package server

import (
	"github.com/Dowte/domain-health/api/web"
	"github.com/Dowte/domain-health/common"
	"github.com/Dowte/domain-health/config"
	"net/http"
)

const (
	apiPrefix       = "/domain_health"
	apiDomainAdd    = apiPrefix + "/add"
	apiDomainDelete = apiPrefix + "/delete"
	apiDomainList   = apiPrefix + "/list"
)

var log = common.Log

func Start() error {
	http.HandleFunc(apiDomainAdd, web.DomainAdd)
	http.HandleFunc(apiDomainList, web.DomainList)

	log.Noticef("start api server on %s", config.Instance.ListenAddress)

	return http.ListenAndServe(config.Instance.ListenAddress, nil)
}
