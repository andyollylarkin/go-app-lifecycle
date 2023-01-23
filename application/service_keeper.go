package application

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	defaultPingPeriod  = time.Second * 15
	defaultPingTimeout = time.Second * 5
)

var (
	ServiceNotFound = errors.New("service not found")
)

type ServiceNotFoundError struct {
	ServiceName string
	Err         error
}

func (e *ServiceNotFoundError) Error() string {
	return fmt.Sprintf("service %s not found", e.ServiceName)
}

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

func (keeper *ServiceKeeper) Get(serviceName string) (any, error) {
	srv := keeper.Services[serviceName]
	if srv == nil {
		return nil, ServiceNotFound
	}
	return (*srv).Get(), nil
}

func (keeper *ServiceKeeper) HealthCheck() error {
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
