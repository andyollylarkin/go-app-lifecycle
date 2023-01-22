package application

import "context"

type MainFunc func(ctx context.Context, shutdown chan struct{}) error
type RecoverFunc func() error
