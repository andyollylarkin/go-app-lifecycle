package application

import (
	"context"
	"errors"
)

type ApplicationStage struct {
	currentState *ApplicationState
}

func NewApplicationStage() *ApplicationStage {
	stage := &ApplicationStage{}
	return stage
}

func (c *ApplicationStage) Init(ctx context.Context, state *ApplicationState, keeper *ServiceKeeper) error {
	// TODO: initialization phase
	if *state != StateStart {
		return errors.New("incorrect state. It's only possible to enter the init state from the START state")
	}

	return nil
}

func (c *ApplicationStage) Start(ctx context.Context, state *ApplicationState, keeper *ServiceKeeper,
	shutdown chan struct{}, mainFunc MainFunc) error {
	err := mainFunc(ctx, shutdown)
	if err != nil {
		return err
	}
	return nil
}

func (c *ApplicationStage) Uninit(ctx context.Context, state *ApplicationState) error {
	return nil
	//	TODO: запустить горутину с деинициализацией и ждать из ctx.Done(). Деинициализация
}

func (c *ApplicationStage) Shutdown(ctx context.Context, state *ApplicationState, shutdown chan struct{}) error {
	if *state == StateShutdown {
		return errors.New("incorrect state. State already shutdown")
	}
	ChangeState(state, StateShutdown)
	close(shutdown)
	<-ctx.Done()
	ChangeState(state, StateInit)
	return nil
}
