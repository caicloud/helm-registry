/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package common

import (
	"context"
	"fmt"
	"net/http"

	"github.com/caicloud/helm-registry/pkg/errors"
	"github.com/caicloud/helm-registry/pkg/log"
	"github.com/caicloud/helm-registry/pkg/storage"
)

// a global SpaceManager
var globalSpaceManager storage.SpaceManager

// GetSpaceManager gets a SpaceManager with configs from default Context.
// kvStore should have two keys:
//  ContextNameSpaceManager: specify the name of SpaceManager
//  ContextNameSpaceParameters: specify the parameters of SpaceManager
func GetSpaceManager() (storage.SpaceManager, error) {
	if globalSpaceManager != nil {
		return globalSpaceManager, nil
	}
	name, ok := Get(ContextNameSpaceManager)
	if !ok {
		return nil, errors.NewResponError(http.StatusInternalServerError, "error.unknown", "${name} error", errors.M{
			"name": ContextNameSpaceManager,
		})
	}
	value, ok := Get(ContextNameSpaceParameters)
	if !ok {
		return nil, errors.NewResponError(http.StatusInternalServerError, "error.unknown", "${name} error", errors.M{
			"name": ContextNameSpaceParameters,
		})
	}
	parameters, ok := value.(map[string]interface{})
	if !ok {
		return nil, errors.NewResponError(http.StatusInternalServerError, "param.error", "${name} error", errors.M{
			"name": ContextNameSpaceParameters,
		})
	}
	manager, err := storage.Create(fmt.Sprint(name), parameters)
	if err != nil {
		return nil, err
	}
	globalSpaceManager = manager
	return manager, nil
}

// MustGetSpaceManager must get a SpaceManager. If not, panic.
func MustGetSpaceManager() storage.SpaceManager {
	manager, err := GetSpaceManager()
	if err != nil {
		log.Panic(err)
	}
	return manager
}

// GetSpace gets an instance of specific space
func GetSpace(ctx context.Context, space string) (storage.Space, error) {
	return MustGetSpaceManager().Space(ctx, space)
}

// GetSpaceAndChart gets instances of space and chart
func GetSpaceAndChart(ctx context.Context, space string, chart string) (storage.Space, storage.Chart, error) {
	iSpace, err := GetSpace(ctx, space)
	if err != nil {
		return nil, nil, err
	}
	iChart, err := iSpace.Chart(ctx, chart)
	return iSpace, iChart, err
}

// GetSpaceChartAndVersion gets instance of space, chart and version
func GetSpaceChartAndVersion(ctx context.Context, space string, chart string, version string) (storage.Space, storage.Chart, storage.Version, error) {
	iSpace, iChart, err := GetSpaceAndChart(ctx, space, chart)
	if err != nil {
		return nil, nil, nil, err
	}
	iVersion, err := iChart.Version(ctx, version)
	return iSpace, iChart, iVersion, err
}

// GetChart gets an instance of specific space
func GetChart(ctx context.Context, space string, chart string) (storage.Chart, error) {
	_, iChart, err := GetSpaceAndChart(ctx, space, chart)
	return iChart, err
}

// GetVersion gets an instance of specific space
func GetVersion(ctx context.Context, space string, chart string, version string) (storage.Version, error) {
	_, _, iVersion, err := GetSpaceChartAndVersion(ctx, space, chart, version)
	return iVersion, err
}
