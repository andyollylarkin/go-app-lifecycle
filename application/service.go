package application

import (
	"context"
)

type Service interface {
	Init(ctx context.Context) error
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	CheckHealthFunc(ctx context.Context) error
	GetService() any
}
