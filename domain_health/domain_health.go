package domain_health

import (
	"errors"
	"github.com/Dowte/domain-health/common"
	"github.com/Dowte/domain-health/config"
	"github.com/Dowte/domain-health/domain_health/fetch"
	"github.com/Dowte/domain-health/domain_health/fetch/aliyun"
	"github.com/Dowte/domain-health/store"
	"github.com/Dowte/domain-health/store/model"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var log = common.Log

var (
	ErrGetExpireTimeErr = errors.New("get domain expire time err")
	ErrNotTSL           = errors.New("not TSL domain")
	ErrTimeOut          = errors.New("check domain timeout")
)

type Service struct {
	fetchers map[model.From]fetch.Fetcher
}

func GetExpireTime(record url.URL) (time time.Time, err error) {
	record.Scheme = "https"
	resp, err := http.Get(record.String())
	if err == nil {
		defer resp.Body.Close()

		if info := resp.TLS; info != nil {
			if len(info.PeerCertificates) > 0 {
				return info.PeerCertificates[0].NotAfter, nil
			} else {
				err = ErrNotTSL
				return
			}
		} else {
			err = ErrGetExpireTimeErr
			return
		}
	} else {
		if e, ok := err.(*url.Error); ok && e.Timeout() {
			err = ErrTimeOut
		} else {
			return
		}
	}
	return
}

func NewService() *Service {
	s := &Service{
		fetchers: map[model.From]fetch.Fetcher{},
	}
	if config.Instance.Fetcher.Aliyun.EnableFetch {
		s.fetchers[model.Aliyun] = &aliyun.Fetcher{
			RegionId:        config.Instance.Fetcher.Aliyun.RegionId,
			AccessKeyId:     config.Instance.Fetcher.Aliyun.AccessKeyId,
			AccessKeySecret: config.Instance.Fetcher.Aliyun.AccessKeySecret,
			BlackRR:         config.Instance.Fetcher.Aliyun.BlackRR,
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
	for from, fetcher := range s.fetchers {
		records, err := fetcher.Fetch()
		if err != nil {
			log.Error(err)
		}

		var oldArr []string

		for _, old := range domainStore.ReadAllDomainByFrom(from) {
			oldArr = append(oldArr, old.Record.String())
		}

		domainStore.DeleteAddressArr(arrayDiff(oldArr, records))

		for _, record := range records {
			if !domainStore.HasDomainByAddress(record) {
				domain := model.NewDomain()
				domain.From = from
				parse, err := url.Parse(record)
				if err != nil {
					log.Error(err)
				} else {
					domain.Record = parse
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
			log.Debugf("checked record [%s]", domain.Record.String())
			wp.Done()
		}(domain)
	}
	wp.Wait()
}

func checkAndSave(domain *model.Domain) {
	expireTime, err := GetExpireTime(*domain.Record)
	if err != nil {
		domain.CheckError = err.Error()
		domain.LastCheckTime = time.Now()
	} else {
		domain.CheckError = ""
		domain.ExpireTime = expireTime
		domain.LastCheckTime = time.Now()
	}

	store.GetDomainStore().SaveDomainInfo(domain)
}
