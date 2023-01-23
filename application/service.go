package application

import (
	"context"
)

type Service interface {
	Init(ctx context.Context) error
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	Get() any
}
