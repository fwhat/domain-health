package main

import (
	// init config at first
	"github.com/qjues/domain-health/config"

	"github.com/qjues/domain-health/api/server"
	"github.com/qjues/domain-health/common"
	"github.com/qjues/domain-health/domain_health"
	"github.com/qjues/domain-health/store"
	"time"
)

var log = common.Log

func main() {
	store.InitDomainStore()

	go server.Start()

	service := domain_health.NewService()

	for {
		service.StartCheck()

		log.Info("check success.")

		time.Sleep(time.Duration(config.Instance.HealthTime) * time.Second)
	}
}
