/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package handlers

import (
	"context"
	"fmt"

	"github.com/caicloud/helm-registry/pkg/api/models"
	"github.com/caicloud/helm-registry/pkg/common"
	"github.com/caicloud/helm-registry/pkg/errors"
	"github.com/caicloud/helm-registry/pkg/orchestration"
	"gopkg.in/yaml.v2"
)

// ListCharts lists charts in specified space
func ListCharts(ctx context.Context) (int, []string, error) {
	spaceName, err := getSpaceName(ctx)
	if err != nil {
		return 0, nil, err
	}
	space, err := common.GetSpace(ctx, spaceName)
	if err != nil {
		return 0, nil, err
	}
	return listStrings(ctx, func() ([]string, error) {
		return space.List(ctx)
	})
}

// DeleteChart deletes specified chart
func DeleteChart(ctx context.Context) error {
	spaceName, chartName, err := getSpaceAndChartName(ctx)
	if err != nil {
		return err
	}
	space, err := common.GetSpace(ctx, spaceName)
	if err != nil {
		return err
	}
	return space.Delete(ctx, chartName)
}

// CreateChart creates a chart by a json config
func CreateChart(ctx context.Context) (*models.ChartLink, error) {
	config, err := getChartConfig(ctx)
	if err != nil {
		return nil, err
	}
	space, _, version, err := common.GetSpaceChartAndVersion(ctx, config.Save.Space, config.Save.Chart, config.Save.Version)
	if err != nil {
		return nil, err
	}
	if !space.Exists(ctx) {
		return nil, errors.ErrorContentNotFound.Format(config.Save.Space)
	}
	if version.Exists(ctx) {
		return nil, errors.ErrorResourceExist.Format(config.Save.Path())
	}
	configs, values, err := separateConfigs(config.Configs)
	if err != nil {
		return nil, err
	}
	// create chart
	newChart, err := orchestration.Create(configs)
	if err != nil {
		return nil, err
	}
	orchestration.ClearValues(newChart)
	// set values
	rawValues, err := yaml.Marshal(values)
	if err != nil {
		return nil, errors.ErrorInternalUnknown.Format(err.Error())
	}
	newChart.Values.Raw = string(rawValues)
	// set chart
	newChart.Metadata.Name = config.Save.Chart
	newChart.Metadata.Version = config.Save.Version
	newChart.Metadata.Description = config.Save.Desc
	// archive chart
	data, err := orchestration.Archive(newChart)
	if err != nil {
		return nil, err
	}
	// save chart
	err = version.PutContent(ctx, data)
	if err != nil {
		return nil, err
	}
	// construct a chart self-link
	path, err := getRequestPath(ctx)
	if err != nil {
		return nil, err
	}
	return models.NewChartLink(config.Save.Space, config.Save.Chart, config.Save.Version,
		fmt.Sprintf("%s/%s/versions/%s", path, config.Save.Chart, config.Save.Version)), nil
}

// valuesConfigName is the key of values
const valuesConfigName = "_config"

// packageName is the key of package
const packageName = "package"

// separateConfigs separates configs and values from original configs
func separateConfigs(originalConfigs map[string]interface{}) (configs map[string]interface{}, values map[string]interface{}, err error) {
	configs = make(map[string]interface{})
	values = make(map[string]interface{})
	for key, value := range originalConfigs {
		switch key {
		case valuesConfigName:
			values[key] = value
		case packageName:
			configs[key] = value
		default:
			data, ok := value.(map[string]interface{})
			if !ok {
				return nil, nil, errors.ErrorParamTypeError.Format(key,
					"map", "unknown")
			}
			a, v, err := separateConfigs(data)
			if err != nil {
				return nil, nil, err
			}
			if len(a) >= 0 {
				configs[key] = a
			}
			values[key] = v
		}
	}
	return
}
