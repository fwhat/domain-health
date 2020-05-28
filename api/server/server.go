package server

import (
	"github.com/qjues/domain-health/api/web"
	"github.com/qjues/domain-health/common"
	"github.com/qjues/domain-health/config"
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
