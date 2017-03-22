/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package lock

import (
	"sync"
	"sync/atomic"
	"time"
)

// Lock describes a simple lock to provide Lock/RLock.
// The lock is very simple and must use Lock/Unlock and RLock/RUnlock correctly. Otherwise
// It will be deadlocked.
type Lock struct {
	sync.RWMutex
	locker, rLocker *cancelableLock
}

// NewMemoryLock creates a lock
func NewMemoryLock() Locker {
	locker := &Lock{}
	locker.locker = newCancelableLock(locker.RWMutex.Lock, locker.RWMutex.Unlock)
	locker.rLocker = newCancelableLock(locker.RWMutex.RLock, locker.RWMutex.RUnlock)
	return locker
}

// Lock tries lock for writing. If locked, return true
func (l *Lock) Lock(timeout time.Duration) bool {
	return l.locker.lockWithTimeout(timeout)
}

// RLock tries lock for reading. If locked, return true
func (l *Lock) RLock(timeout time.Duration) bool {
	return l.rLocker.lockWithTimeout(timeout)
}

// cancelableLock provides a cancelable lock
type cancelableLock struct {
	// lock is the lock method of a anonymous lock
	lock func()
	// unlock is the unlock method of a anonymous lock
	unlock func()
}

// newCancelableLock creates a cancelable lock
func newCancelableLock(lock func(), unlock func()) *cancelableLock {
	return &cancelableLock{lock, unlock}
}

// lockWithTimeout tries lock with a timeout. When time off and the lock can't get
// control, return false and release the lock request automatically.
func (cl *cancelableLock) lockWithTimeout(timeout time.Duration) bool {
	if timeout <= 0 {
		return false
	}
	ch, cancel := cl.goLock()
	select {
	case <-ch:
		return true
	case <-time.After(timeout):
		return !cancel()
	}
}

// goLock tries lock and returns a channel and a cancel function.
// If call the cancel function before get the lock, it will unlock
// immediately when it get the lock.
func (cl *cancelableLock) goLock() (<-chan struct{}, func() bool) {
	canceled := int32(0)
	ch := make(chan struct{})
	go func() {
		cl.lock()
		if atomic.AddInt32(&canceled, 1) == 1 {
			ch <- struct{}{}
		} else {
			cl.unlock()
		}
	}()
	return ch, func() bool {
		return atomic.AddInt32(&canceled, 1) == 1
	}
}

// MemoryLockFactory is a factory for creating ResourceLocker
type MemoryLockFactory struct {
}

// Create creates a new ResourceLocker
func (mlf *MemoryLockFactory) Create(map[string]interface{}) (ResourceLocker, error) {
	return NewResourceLock(NewMemoryLock), nil
}

func init() {
	// register memory ResourceLocker
	Register("memory", &MemoryLockFactory{})
}
