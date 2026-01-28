package httpapi

import (
	"sync"
	"sync/atomic"
)

// CallRegistry tracks active call sessions and supports graceful draining.
// When draining is enabled, new calls are rejected while in-flight calls
// finish naturally.
//
// The mu mutex makes the draining check and wg.Add atomic in Add(), preventing
// a TOCTOU race where StartDraining+Wait could be called between the draining
// check and wg.Add.
type CallRegistry struct {
	mu       sync.Mutex
	draining bool
	wg       sync.WaitGroup
	count    atomic.Int64
}

// NewCallRegistry creates a new CallRegistry.
func NewCallRegistry() *CallRegistry {
	return &CallRegistry{}
}

// Add registers a new active call. Returns false if the registry is draining,
// meaning no new calls should be accepted. The draining check and WaitGroup
// increment are performed atomically under a mutex.
func (cr *CallRegistry) Add() bool {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	if cr.draining {
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
// This is safe to call concurrently with Add — the mutex ensures no Add can
// slip through after StartDraining returns.
func (cr *CallRegistry) StartDraining() {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.draining = true
}

// IsDraining reports whether the registry is in draining mode.
func (cr *CallRegistry) IsDraining() bool {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	return cr.draining
}

// ActiveCount returns the number of currently active calls.
func (cr *CallRegistry) ActiveCount() int64 {
	return cr.count.Load()
}

// Wait blocks until all active calls have completed (all Done calls matched Add calls).
func (cr *CallRegistry) Wait() {
	cr.wg.Wait()
}
