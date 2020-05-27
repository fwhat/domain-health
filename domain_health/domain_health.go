package domain_health

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/Dowte/domain-health/common"
	"github.com/Dowte/domain-health/config"
	"github.com/Dowte/domain-health/domain_health/fetcher"
	"github.com/Dowte/domain-health/domain_health/fetcher/aliyun"
	"github.com/Dowte/domain-health/store"
	"github.com/Dowte/domain-health/store/model"
	"math/rand"
	"net"
	"net/url"
	"sync"
	"time"
)

var log = common.Log

var (
	ErrNotTSL = errors.New("not TSL domain")
)

type Service struct {
	fetchers map[model.From]fetcher.Fetcher
}

func GetCretInfo(address string) (certInfo model.CertInfo, err error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:443", address), time.Second*10)
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

func NewService() *Service {
	s := &Service{
		fetchers: map[model.From]fetcher.Fetcher{},
	}
	if config.Instance.Fetcher.Aliyun.EnableFetch {
		s.fetchers[model.Aliyun] = &aliyun.Fetcher{
			RegionId:        config.Instance.Fetcher.Aliyun.RegionId,
			AccessKeyId:     config.Instance.Fetcher.Aliyun.AccessKeyId,
			AccessKeySecret: config.Instance.Fetcher.Aliyun.AccessKeySecret,
			BlackRR:         config.Instance.Fetcher.Aliyun.BlackRR,
			OnlyType:        config.Instance.Fetcher.Aliyun.OnlyType,
		}
	}

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
			checkAndSave(domain)
			log.Debugf("checked record [%s]", domain.Address)
			wp.Done()
		}(domain)
	}
	wp.Wait()
}

func checkAndSave(domain *model.Domain) {
	certInfo, err := GetCretInfo(domain.Address)
	if err != nil {
		domain.CheckError = err.Error()
		domain.LastCheckTime = time.Now().Unix()
	} else {
		domain.CheckError = ""
		domain.CertInfo = certInfo
		domain.LastCheckTime = time.Now().Unix()
	}

	store.GetDomainStore().SaveDomainInfo(domain)
}
