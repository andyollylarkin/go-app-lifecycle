package application

import "context"

type MainFunc func(ctx context.Context, wait func(), keeper *ServiceKeeper) error
type RecoverFunc func() error
