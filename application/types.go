package application

import "context"

type MainFunc func(ctx context.Context, wait func()) error
type RecoverFunc func() error
