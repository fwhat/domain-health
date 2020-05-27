package main

import (
	"github.com/Dowte/domain-health/config"
	_ "github.com/Dowte/domain-health/config"
	"github.com/Dowte/domain-health/domain_health"
	"time"

	"github.com/Dowte/domain-health/api/server"
	"github.com/Dowte/domain-health/store"
)

func main() {
	store.InitDomainStore()

	go server.Start()

	for {
		domain_health.StartCheck()

		time.Sleep(time.Duration(config.Instance.HealthTime) * time.Second)
	}
}
