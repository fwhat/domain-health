package domain_health

import (
	"errors"
	"github.com/Dowte/domain-health/store"
	"github.com/Dowte/domain-health/store/model"
	"net/http"
	"net/url"
	"time"
)

var (
	ErrGetExpireTimeErr = errors.New("get domain expire time err")
	ErrNotTSL           = errors.New("not TSL domain")
	ErrTimeOut          = errors.New("check domain timeout")
)

func GetExpireTime(address string) (time time.Time, err error) {
	resp, err := http.Get(address)
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

func StartCheck() {
	list := store.GetDomainStore().ReadAllDomainList()
	for _, domain := range list {
		go checkAndSave(domain)
	}
}

func checkAndSave(domain *model.Domain) {
	expireTime, err := GetExpireTime(domain.Address)
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
