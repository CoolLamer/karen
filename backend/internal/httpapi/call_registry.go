package httpapi

import (
	"sync"
	"sync/atomic"
)

// CallRegistry tracks active call sessions and supports graceful draining.
// When draining is enabled, new calls are rejected while in-flight calls
// finish naturally.
type CallRegistry struct {
	draining atomic.Bool
	wg       sync.WaitGroup
	count    atomic.Int64
}

// NewCallRegistry creates a new CallRegistry.
func NewCallRegistry() *CallRegistry {
	return &CallRegistry{}
}

// Add registers a new active call. Returns false if the registry is draining,
// meaning no new calls should be accepted.
func (cr *CallRegistry) Add() bool {
	if cr.draining.Load() {
		return false
	}
	cr.wg.Add(1)
	cr.count.Add(1)
	return true
}

// Done marks a call as completed. Must be called exactly once per successful Add.
func (cr *CallRegistry) Done() {
	cr.count.Add(-1)
	cr.wg.Done()
}

// StartDraining sets the draining flag so that future Add calls return false.
func (cr *CallRegistry) StartDraining() {
	cr.draining.Store(true)
}

// IsDraining reports whether the registry is in draining mode.
func (cr *CallRegistry) IsDraining() bool {
	return cr.draining.Load()
}

// ActiveCount returns the number of currently active calls.
func (cr *CallRegistry) ActiveCount() int64 {
	return cr.count.Load()
}

// Wait blocks until all active calls have completed (all Done calls matched Add calls).
func (cr *CallRegistry) Wait() {
	cr.wg.Wait()
}
