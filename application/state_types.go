package application

import "sync/atomic"

type ApplicationState int32

const (
	StateStart = ApplicationState(iota) // начальное состояние приложения
	StateInit
	StateRunning
	StateDeInit
	StateShutdown
)

func ChangeState(oldState *ApplicationState, newState ApplicationState) {
	var stateAddr = (*int32)(oldState)
	atomic.CompareAndSwapInt32(stateAddr, int32(*oldState), int32(newState))
}
