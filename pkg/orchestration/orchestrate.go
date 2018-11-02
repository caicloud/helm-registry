/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package orchestration

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/caicloud/helm-registry/pkg/common"
	"github.com/caicloud/helm-registry/pkg/errors"
	"github.com/caicloud/helm-registry/pkg/storage"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

const (
	// packageKey is the key of package in configs
	packageKey = "package"
)

// convertInterface converts interface{} to map[string]interface{}
func convertInterface(name string, inf interface{}) (map[string]interface{}, error) {
	if inf == nil {
		return map[string]interface{}{}, nil
	}
	data, ok := inf.(map[string]interface{})
	if !ok {
		return nil, errors.NewResponError(http.StatusBadRequest, "param.error", "${param} error", errors.M{
			"name": name,
		})
	}
	return data, nil
}

// Create creates a new chart from configs.
// An example:
// {
//     "package": {                     // It's the fixed description of current chart
//         "independent":true,          // It means that the chart is an independent chart in registry
//         "space":"space name",        // Chart space
//         "chart":"chart name",        // Original chart name
//         "version":"version number"   // Original chart version
//                                      // We will be able to find the root chart by space/chart/version
//     },
//     "chartB": {                      // The original chart is library/chartA/1.0.2, but is renamed to chartB.
//                                      // You can rename it or not. It's depend on you.
//         "package":{                  // We will find the chart by library/chartA/1.0.2
//             "independent":true,
//             "space":"library",
//             "chart":"chartA",        // The original name of chartB is chartA
//             "version":"1.0.2"
//         },
//         "chartD": {                  // This chart is the subchart of chartA and original name is chartD
//             "package":{              // We will find the subchart from library/chartA/1.0.2
//                 "independent":false,
//                 "space":"library",   // When independent is false, the space should be same as its parent
//                 "chart":"chartD",    // The original name of the subchart
//                 "version":"2.3.4"    // The version of subchart
//             }
//         }
//     },
//     "chartC": {
//         "package":{
//             "independent":false,
//             "space":"library",
//             "chart":"chartC",
//             "version":"1.0.0"
//         }
//     }
// }
func Create(configs map[string]interface{}) (*chart.Chart, error) {
	return create(nil, configs)
}

// ClearValues removes all values in a chart
func ClearValues(chrt *chart.Chart) {
	chrt.Values = &chart.Config{}
	for _, child := range chrt.Dependencies {
		ClearValues(child)
	}
}

// create creates a new chart from configs.
func create(parent *chart.Chart, configs map[string]interface{}) (*chart.Chart, error) {
	// packageConfig is the config of current package
	var packageConfig *Package

	// deps is the children of parent
	deps := make(map[string]map[string]interface{})
	// find packageConfig and deps from configs
	for key, value := range configs {
		data, err := convertInterface(key, value)
		if err != nil {
			return nil, err
		}
		if key == packageKey {
			// catch package
			packageConfig, err = NewPackage(data)
			if err != nil {
				return nil, err
			}
		} else {
			// filter invalid chart name
			if !common.MustGetSpaceManager().Validate(context.Background(),
				storage.ValidationTypeChartName, key) {
				return nil, errors.NewResponError(http.StatusBadRequest, "charts.invalidate", "${name} invalidate", errors.M{
					"name": key,
				})
			}
			deps[key] = data
		}
	}
	currentChart, err := getChartByPackage(parent, packageConfig)
	if err != nil {
		return nil, err
	}
	// generate charts recursively
	if len(deps) > 0 {
		children := make([]*chart.Chart, 0, len(deps))
		for name, cfg := range deps {
			child, err := create(currentChart, cfg)
			if err != nil {
				return nil, err
			}
			child.Metadata.Name = name
			children = append(children, child)
		}
		currentChart.Dependencies = children
	} else {
		currentChart.Dependencies = []*chart.Chart{}
	}
	return currentChart, nil
}

// getChartByPackage returns a chart via package configs
func getChartByPackage(parent *chart.Chart, pkg *Package) (*chart.Chart, error) {
	chartName := fmt.Sprintf("%s/%s", pkg.Chart, pkg.Version)
	if pkg.Independent {
		return getChart(pkg.Space, pkg.Chart, pkg.Version)
	}
	if parent == nil {
		return nil, errors.NewResponError(http.StatusBadRequest, "charts.invalidate", "${name} invalidate", errors.M{
			"name": chartName,
		})
	}
	for _, dep := range parent.GetDependencies() {
		if dep.GetMetadata().Name == pkg.Chart {
			return dep, nil
		}
	}
	return nil, errors.NewResponError(http.StatusNotFound, "content.unfound", "${name} not found", errors.M{
		"name": fmt.Sprintf("%s in %s/%s", chartName, parent.Metadata.Name, parent.Metadata.Version),
	})
}

// getChart gets a chart
func getChart(spaceName, chartName, versionNumber string) (*chart.Chart, error) {
	ctx := context.Background()
	space, err := common.MustGetSpaceManager().Space(ctx, spaceName)
	if err != nil {
		return nil, err
	}
	chart, err := space.Chart(ctx, chartName)
	if err != nil {
		return nil, err
	}
	version, err := chart.Version(ctx, versionNumber)
	if err != nil {
		return nil, err
	}
	data, err := version.GetContent(ctx)
	if err != nil {
		return nil, err
	}
	c, err := chartutil.LoadArchive(bytes.NewReader(data))
	if err != nil {
		return nil, errors.NewResponError(http.StatusInternalServerError, "param.error", "${name} error", errors.M{
			"name": fmt.Sprintf("%s/%s", chartName, versionNumber),
		})
	}
	return c, nil
}
