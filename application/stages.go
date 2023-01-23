package application

import (
	"context"
	"errors"
)

type Stage struct {
}

func NewApplicationStage() *Stage {
	stage := &Stage{}
	return stage
}

func (c *Stage) Init(ctx context.Context, state *ApplicationState, keeper *ServiceKeeper) error {
	// TODO: initialization phase
	if *state != StateStart {
		return errors.New("incorrect state. It's only possible to enter the init state from the START state")
	}

	return nil
}

func (c *Stage) Start(ctx context.Context, state *ApplicationState, keeper *ServiceKeeper,
	mainFunc MainFunc, waitFunc func()) error {
	err := mainFunc(ctx, waitFunc)
	if err != nil {
		return err
	}
	return nil
}

func (c *Stage) Uninit(ctx context.Context, state *ApplicationState) error {
	return nil
	//	TODO: запустить горутину с деинициализацией и ждать из ctx.Done(). Деинициализация
	return errors.New("deinitialization timeout expired")
}

func (c *Stage) Shutdown(ctx context.Context, state *ApplicationState, shutdown chan struct{}) error {
	if *state == StateShutdown {
		return errors.New("incorrect state. State already shutdown")
	}
	ChangeState(state, StateShutdown)
	close(shutdown)
	<-ctx.Done()
	ChangeState(state, StateInit)
	return nil
}
