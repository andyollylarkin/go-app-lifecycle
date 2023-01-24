package application

import (
	"context"
)

type Service interface {
	// Init initialization service
	Init(ctx context.Context) error
	// Ping check if service keep alive
	Ping(ctx context.Context) error
	// Close resources
	Close(ctx context.Context) error
	// Get getting service
	Get() any
}
