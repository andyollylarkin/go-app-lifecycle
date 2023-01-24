package application

import "sync/atomic"

type State int32

const (
	StateStart = State(iota) // initial state
	StateInit
	StateRunning
	StateShutdown
	StateUninit
	StateTerminated
)

// ChangeState atomically changes the application state
func ChangeState(oldState *State, newState State) {
	var stateAddr = (*int32)(oldState)
	atomic.CompareAndSwapInt32(stateAddr, int32(*oldState), int32(newState))
}

// IsStateEqual compare application state with target state. True if states is equal, false otherwise
func IsStateEqual(current State, target State) bool {
	if current == target {
		return true
	}
	return false
}
