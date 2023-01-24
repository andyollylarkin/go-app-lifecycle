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

func (c *Stage) Init(ctx context.Context, state *State, keeper *ServiceKeeper) error {
	if !IsStateEqual(*state, StateStart) {
		return errors.New("incorrect state. It's only possible to enter the init state from the START state")
	}
	ChangeState(state, StateInit)
	var err error
	var errCh = make(chan error)
	go func() {
		for _, srv := range keeper.Services {
			s := *srv
			err = s.Init(ctx)
			if err != nil {
				errCh <- err
			}
		}
		close(errCh)
	}()
	select {
	case <-ctx.Done():
		return errors.New("initialization timeout exceeded")
	case e := <-errCh:
		return e
	}
}

func (c *Stage) Start(ctx context.Context, state *State, keeper *ServiceKeeper,
	mainFunc MainFunc, waitFunc func()) error {
	if !IsStateEqual(*state, StateInit) {
		return errors.New("incorrect state. It's only possible to enter the start state from the INIT state")
	}
	ChangeState(state, StateRunning)
	err := mainFunc(ctx, waitFunc, keeper)
	ChangeState(state, StateShutdown)
	if err != nil {
		return err
	}
	return nil
}

func (c *Stage) Uninit(ctx context.Context, state *State, keeper *ServiceKeeper) error {
	if !IsStateEqual(*state, StateShutdown) {
		return errors.New("incorrect state. State should be SHUTDOWN")
	}
	ChangeState(state, StateUninit)
	var err error
	var errCh = make(chan error)
	go func() {
		for _, srv := range keeper.Services {
			s := *srv
			err = s.Close(ctx)
			if err != nil {
				errCh <- err
			}
		}
		close(errCh)
	}()
	select {
	case <-ctx.Done():
		ChangeState(state, StateTerminated)
		return errors.New("uninitialization timeout exceeded")
	case e := <-errCh:
		ChangeState(state, StateTerminated)
		return e
	}
}

func (c *Stage) Shutdown(ctx context.Context, state *State, shutdown chan struct{}) error {
	if !IsStateEqual(*state, StateShutdown) {
		return errors.New("incorrect state. State should be SHUTDOWN")
	}
	close(shutdown)
	<-ctx.Done()
	return nil
}
