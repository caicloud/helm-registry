/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package storage

import (
	"fmt"
	"sync"
)

// SpaceManagerFactory is a factory for creating SpaceManager
type SpaceManagerFactory interface {
	// Create creates a new SpaceManager
	Create(map[string]interface{}) (SpaceManager, error)
}

var (
	// factoriesMu is used for protecting factories
	factoriesMu sync.RWMutex
	// factories stores all registered SpaceManagerFactory
	factories = make(map[string]SpaceManagerFactory)
)

// Register registers a SpaceManagerFactory
func Register(name string, factory SpaceManagerFactory) {
	if factory == nil {
		panic("Must not provide nil SpaceManagerFactory")
	}
	factoriesMu.Lock()
	defer factoriesMu.Unlock()
	_, registered := factories[name]
	if registered {
		panic(fmt.Sprintf("SpaceManagerFactory named %s already registered", name))
	}
	factories[name] = factory
}

// Create creates a new SpaceManagerFactory with the given name and parameters.
func Create(name string, parameters map[string]interface{}) (SpaceManager, error) {
	factoriesMu.RLock()
	factory, ok := factories[name]
	factoriesMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("SpaceManagerFactory not registered: %s", name)
	}
	return factory.Create(parameters)
}
