/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package common

// kvStore stores global data
var kvStore = make(map[string]interface{})

// Set saves a key-value pair in global scope. It's not thread-safe
func Set(key string, value interface{}) {
	kvStore[key] = value
}

// Get gets a value by key from global scope. It's not thread-safe
func Get(key string) (interface{}, bool) {
	v, ok := kvStore[key]
	return v, ok
}
