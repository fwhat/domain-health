package domain_health

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/qjues/domain-health/common"
	"github.com/qjues/domain-health/config"
	"github.com/qjues/domain-health/domain_health/fetcher"
	"github.com/qjues/domain-health/domain_health/fetcher/aliyun"
	"github.com/qjues/domain-health/domain_health/subscriber"
	"github.com/qjues/domain-health/domain_health/subscriber/dingtalk"
	"github.com/qjues/domain-health/store"
	"github.com/qjues/domain-health/store/model"
	"math/rand"
	"net"
	"net/url"
	"sync"
	"time"
)

var log = common.Log

var (
	ErrNotTSL               = errors.New("not TSL domain")
	ErrExpiredTSL           = errors.New("expired TSL domain")
	ErrDomainConnectTimeout = errors.New("domain connect timeout")
)

type Service struct {
	fetchers    map[model.From]fetcher.Fetcher
	subscribers []subscriber.Subscriber

	certWarnDelivered *sync.Map
}

func GetCretInfo(address string) (certInfo model.CertInfo, err error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:443", address), time.Second*time.Duration(config.Instance.ConnectTimeout))
	if err != nil {
		return certInfo, err
	}

	client := tls.Client(conn, &tls.Config{
		Rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
		ServerName: address,
	})

	err = client.Handshake()
	if err != nil {
		return certInfo, err
	}

	err = client.VerifyHostname(address)
	if err != nil {
		return certInfo, err
	}

	certificates := client.ConnectionState().PeerCertificates

	if len(certificates) > 0 {
		certInfo.ExpireTime = certificates[0].NotAfter.Unix()
		certInfo.CommonName = certificates[0].Issuer.CommonName

		return certInfo, nil
	} else {
		return certInfo, ErrNotTSL
	}
}

func (s *Service) initFetcher() {
	if config.Instance.Fetcher.Aliyun.Enable {
		s.fetchers[model.Aliyun] = &aliyun.Fetcher{
			RegionId:        config.Instance.Fetcher.Aliyun.RegionId,
			AccessKeyId:     config.Instance.Fetcher.Aliyun.AccessKeyId,
			AccessKeySecret: config.Instance.Fetcher.Aliyun.AccessKeySecret,
			BlackRR:         config.Instance.Fetcher.Aliyun.BlackRR,
			OnlyType:        config.Instance.Fetcher.Aliyun.OnlyType,
		}
	}
}

func (s *Service) initSubscriber() {
	if config.Instance.Subscriber.DingTalk.Enable {
		s.subscribers = append(s.subscribers, &dingtalk.Subscriber{
			Secret:  config.Instance.Subscriber.DingTalk.Secret,
			WebHook: config.Instance.Subscriber.DingTalk.WebHook,
		})
	}
}

func (s *Service) addSubscriberMessage(message subscriber.Message) {
	for _, sub := range s.subscribers {
		last, ok := s.certWarnDelivered.Load(fmt.Sprintf("%s-%s", message.Domain.Address, message.Type))
		if ok && last.(int64)+config.Instance.SubscribeMessageCalm > time.Now().Unix() {
			log.Debugf("message calm [%s] [%s] last time: [%s] ", message.Domain.Address, message.Type, time.Unix(last.(int64), 0).Format("2006-01-02 15:04:05"))
			continue
		}

		sub.AddMessage(message)
		s.certWarnDelivered.Store(fmt.Sprintf("%s-%s", message.Domain.Address, message.Type), time.Now().Unix())
	}
}

func (s *Service) delivery() {
	for _, sub := range s.subscribers {
		err := sub.Delivery()
		if err != nil {
			log.Error(err)
		}
	}
}

func NewService() *Service {
	s := &Service{
		fetchers:          map[model.From]fetcher.Fetcher{},
		certWarnDelivered: &sync.Map{},
	}
	s.initFetcher()
	s.initSubscriber()

	return s
}

func arrayDiff(array1 []string, array2 []string) []string {
	array2Map := map[string]bool{}
	for _, s := range array2 {
		array2Map[s] = true
	}

	var diff []string

	for _, s := range array1 {
		if _, ok := array2Map[s]; !ok {
			diff = append(diff, s)
		}
	}

	return diff
}

func (s *Service) reFetch() {
	domainStore := store.GetDomainStore()
	for from, f := range s.fetchers {
		records, err := f.Fetch()
		if err != nil {
			log.Error(err)
		}

		var oldArr []string

		for _, old := range domainStore.ReadAllDomainByFrom(from) {
			oldArr = append(oldArr, old.Address)
		}

		domainStore.DeleteAddressArr(arrayDiff(oldArr, records))

		for _, record := range records {
			if !domainStore.HasDomainByAddress(record) {
				domain := model.NewDomain()
				domain.From = from
				_, err := url.Parse(record)
				if err != nil {
					log.Error(err)
				} else {
					domain.Address = record
					domainStore.SaveDomainInfo(domain)
				}
			}
		}
	}
}

func (s *Service) StartCheck() {
	s.reFetch()

	domainStore := store.GetDomainStore()
	list := domainStore.ReadAllDomainList()
	wp := sync.WaitGroup{}
	for _, domain := range list {
		wp.Add(1)
		go func(domain *model.Domain) {
			start := time.Now()
			checkAndSave(domain)
			// 有设置订阅域名证书报警时间 && 域名正常 && 域名的证书过期时间小于报警阀值

			if domain.CheckError == ErrExpiredTSL {
				s.addSubscriberMessage(subscriber.Message{
					Type:   subscriber.CretExpired,
					Domain: domain,
				})
			}

			if domain.CheckError == ErrDomainConnectTimeout {
				s.addSubscriberMessage(subscriber.Message{
					Type:   subscriber.ConnectTimeout,
					Domain: domain,
				})
			}

			if config.Instance.SubscribeCertWarning > 0 && domain.CheckError == nil && domain.CertInfo.ExpireTime < start.Unix()+config.Instance.SubscribeCertWarning {
				s.addSubscriberMessage(subscriber.Message{
					Type:   subscriber.CretPreExpired,
					Domain: domain,
				})
			}
			log.Debugf("checked record [%s] speed %v", domain.Address, time.Now().Sub(start))
			wp.Done()
		}(domain)
	}
	wp.Wait()
	s.delivery()
}

func checkAndSave(domain *model.Domain) {
	certInfo, err := GetCretInfo(domain.Address)
	if err != nil {
		if hostnameErr, ok := err.(x509.HostnameError); ok {
			domain.CheckError = ErrNotTSL
			domain.OriginError = hostnameErr
		} else if certificateInvalidError, ok := err.(x509.CertificateInvalidError); ok {
			domain.CheckError = ErrNotTSL
			domain.OriginError = nil

			for _, name := range certificateInvalidError.Cert.DNSNames {
				if name == domain.Address {
					domain.CheckError = ErrExpiredTSL
					domain.OriginError = certificateInvalidError
					domain.CertInfo.ExpireTime = certificateInvalidError.Cert.NotAfter.Unix()
					domain.CertInfo.CommonName = certificateInvalidError.Cert.Issuer.CommonName
				}
			}

		} else if opError, ok := err.(*net.OpError); ok && opError.Timeout() {
			domain.CheckError = ErrDomainConnectTimeout
			domain.OriginError = opError

		} else {
			domain.CheckError = err
			domain.OriginError = err
			domain.LastCheckTime = time.Now().Unix()
		}

	} else {
		domain.CheckError = nil
		domain.CertInfo = certInfo
		domain.LastCheckTime = time.Now().Unix()
	}

	store.GetDomainStore().SaveDomainInfo(domain)
}
