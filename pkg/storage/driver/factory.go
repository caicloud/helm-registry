/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package driver

import "github.com/docker/distribution/registry/storage/driver/factory"

// Create creates a specific StorageDriver
func Create(name string, parameters map[string]interface{}) (StorageDriver, error) {
	return factory.Create(name, parameters)
}
