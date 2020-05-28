package main

import (
	// init config at first
	"github.com/Dowte/domain-health/config"

	"github.com/Dowte/domain-health/api/server"
	"github.com/Dowte/domain-health/common"
	"github.com/Dowte/domain-health/domain_health"
	"github.com/Dowte/domain-health/store"
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
