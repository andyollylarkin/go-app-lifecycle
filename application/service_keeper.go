package application

import (
	"errors"
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
	//attemps := 3
	//attempsDelay := time.Second
	//timeout, cancel := context.WithTimeout(context.Background(), keeper.PingTimeout)
	//defer cancel()
	//for {
	//	time.Sleep(keeper.PingPeriod) // pool every service after timeout
	//}
	println("start healthcheck")
	time.Sleep(time.Second * 4)
	println("done healthcheck")
	return errors.New("error when health check")
}
