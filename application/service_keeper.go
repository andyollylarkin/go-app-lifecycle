package application

import (
	"context"
	"time"
)

const (
	defaultPingPeriod  = time.Second * 15
	defaultPingTimeout = time.Second * 5
)

type ServiceKeeper struct {
	Services    map[string]*Service
	PingPeriod  time.Duration
	PingTimeout time.Duration
}

func NewServiceKeeper(services map[string]*Service, pingPeriod time.Duration,
	pingTimeout time.Duration) *ServiceKeeper {
	if pingPeriod == 0 {
		pingPeriod = defaultPingPeriod
	}
	if pingTimeout == 0 {
		pingTimeout = defaultPingTimeout
	}
	return &ServiceKeeper{Services: services, PingPeriod: pingPeriod, PingTimeout: pingTimeout}
}

func (keeper *ServiceKeeper) HealthCheck() error {
	// TODO: services healthcheck
	services := keeper.Services
	var err error
OUTER:
	for {
		time.Sleep(keeper.PingPeriod)
		for _, srv := range services {
			timeoutCtx, cancel := context.WithTimeout(context.Background(), keeper.PingTimeout)
			s := *srv
			err = s.Ping(timeoutCtx)
			cancel()
			if err != nil {
				break OUTER
			}
		}
	}
	return err
}
