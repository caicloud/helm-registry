/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package orchestration

import "github.com/caicloud/helm-registry/pkg/errors"

// Package describes infomation of a chart in registry
type Package struct {
	// Independent identifies whether the chart is an independent and complete chart package.
	// If the field is false, It means that the package is contained by its parent
	// package. On other word, the package is not an independent package.
	Independent bool
	// Space is the name of space where the package is stored. If Independent is false, the
	// space is same as its parent's space
	Space string
	// Chart is the original name of the chart
	Chart string
	// Version is the version number of the chart.
	Version string
}

// NewPackage creates a package from config
func NewPackage(config map[string]interface{}) (*Package, error) {
	independent, err := findBoolFromConfig(config, "independent")
	if err != nil {
		return nil, err
	}
	space, err := findStringFromConfig(config, "space")
	if err != nil {
		return nil, err
	}
	chart, err := findStringFromConfig(config, "chart")
	if err != nil {
		return nil, err
	}
	version, err := findStringFromConfig(config, "version")
	if err != nil {
		return nil, err
	}
	return &Package{
		Independent: independent,
		Space:       space,
		Chart:       chart,
		Version:     version,
	}, nil
}

// findBoolFromConfig finds param from config
func findBoolFromConfig(config map[string]interface{}, param string) (bool, error) {
	value, ok := config[param]
	if !ok {
		return false, errors.ErrorParamNotFound.Format(param)
	}
	v, ok := value.(bool)
	if !ok {
		return false, errors.ErrorParamTypeError.Format(param, "string", "unknown")
	}
	return v, nil
}

// findStringFromConfig finds param from config
func findStringFromConfig(config map[string]interface{}, param string) (string, error) {
	value, ok := config[param]
	if !ok {
		return "", errors.ErrorParamNotFound.Format(param)
	}
	v, ok := value.(string)
	if !ok {
		return "", errors.ErrorParamTypeError.Format(param, "string", "unknown")
	}
	return v, nil
}
