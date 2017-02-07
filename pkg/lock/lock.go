/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package lock

import (
	"fmt"
	"math"
	"sync"
	"time"
)

var (
	// TimeoutImmediate stands for a no-waitting operation. It lets Locker try
	// to lock and return immediately if failed.
	// In fact it use 5 milliseconds to instead the concept of immediate.
	// A timeout with 0 can't acquire the lock.
	TimeoutImmediate = time.Duration(5) * time.Millisecond
	// TimeoutInfinite stands for a infinited waitting operation. It lets Locker
	// try to lock and return while success.
	TimeoutInfinite = time.Duration(math.MaxInt64)
)

// Locker describes a interface for a resource lock
type Locker interface {
	// Lock tries lock for writing. If success to lock, return true.
	// If timeout less or equal than 0, It always return false.
	Lock(timeout time.Duration) bool
	// Unlock unlock write lock
	Unlock()
	// RLock tries lock for reading. If success to lock, return true
	// If timeout less or equal than 0, It always return false.
	RLock(timeout time.Duration) bool
	// RUnlock unlock read lock
	RUnlock()
}

// ResourceLocker describes a interface to manage resource locks
type ResourceLocker interface {
	// Get gets a lock for resources. The locker can lock multiple level resources.
	Get(res ...string) Locker
	// Close releases all lockers
	Close()
}

// ResourceLockerFactory is a factory for creating ResourceLocker
type ResourceLockerFactory interface {
	// Create creates a new ResourceLocker
	Create(map[string]interface{}) (ResourceLocker, error)
}

var (
	// factoriesMu is used for protecting factories
	factoriesMu sync.RWMutex
	// factories stores all registered ResourceLocker
	factories = make(map[string]ResourceLockerFactory)
)

// Register registers a ResourceLockerFactory
func Register(name string, factory ResourceLockerFactory) {
	if factory == nil {
		panic("Must not provide nil ResourceLockerFactory")
	}
	factoriesMu.Lock()
	defer factoriesMu.Unlock()
	_, registered := factories[name]
	if registered {
		panic(fmt.Sprintf("ResourceLockerFactory named %s already registered", name))
	}
	factories[name] = factory
}

// Create creates a new ResourceLocker with the given name and parameters.
func Create(name string, parameters map[string]interface{}) (ResourceLocker, error) {
	factoriesMu.RLock()
	factory, ok := factories[name]
	factoriesMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("ResourceLockerFactory not registered: %s", name)
	}
	return factory.Create(parameters)
}
