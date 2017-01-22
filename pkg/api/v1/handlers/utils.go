/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/caicloud/helm-registry/pkg/api/definition"
	"github.com/caicloud/helm-registry/pkg/api/v1/types"
	"github.com/caicloud/helm-registry/pkg/common"
	"github.com/caicloud/helm-registry/pkg/errors"
	"github.com/caicloud/helm-registry/pkg/storage"
	"github.com/emicklei/go-restful"
)

// getRequestFromContext get request from context
func getRequestFromContext(ctx context.Context) (*restful.Request, error) {
	value := ctx.Value(definition.KeyRequest)
	if v, ok := value.(*restful.Request); ok {
		return v, nil
	}
	return nil, errors.ErrorUnknownNotFoundError.Format(definition.KeyRequest)
}

// getPathParameter gets value from request.PathParameter
func getPathParameter(ctx context.Context, name string) (string, error) {
	request, err := getRequestFromContext(ctx)
	if err != nil {
		return "", err
	}
	value := request.PathParameter(name)
	if len(value) <= 0 {
		return "", errors.ErrorParamNotFound.Format(name)
	}
	return value, nil
}

// getHeaderParameter gets value from request.HeaderParameter
func getHeaderParameter(ctx context.Context, name string) (string, error) {
	request, err := getRequestFromContext(ctx)
	if err != nil {
		return "", err
	}
	value := request.HeaderParameter(name)
	if len(value) <= 0 {
		return "", errors.ErrorParamNotFound.Format(name)
	}
	return value, nil
}

// getQueryParameter gets value from request.QueryParameter
func getQueryParameter(ctx context.Context, name string) (string, error) {
	request, err := getRequestFromContext(ctx)
	if err != nil {
		return "", err
	}
	value := request.QueryParameter(name)
	if len(value) <= 0 {
		return "", errors.ErrorParamNotFound.Format(name)
	}
	return value, nil
}

// getSpaceName gets space name
func getSpaceName(ctx context.Context) (string, error) {
	const field = "space"
	name, err := getPathParameter(ctx, field)
	if err != nil {
		name, err = getQueryParameter(ctx, field)
	}
	return name, err
}

// getChartName gets chart name
func getChartName(ctx context.Context) (string, error) {
	const field = "chart"
	return getPathParameter(ctx, field)
}

// getSpaceAndChartName gets space and chart name
func getSpaceAndChartName(ctx context.Context) (string, string, error) {
	space, err := getSpaceName(ctx)
	if err != nil {
		return "", "", err
	}
	chart, err := getChartName(ctx)
	if err != nil {
		return "", "", err
	}
	return space, chart, nil
}

// getVersionNumber gets version number
func getVersionNumber(ctx context.Context) (string, error) {
	const field = "version"
	return getPathParameter(ctx, field)
}

// getSpaceChartNameAndVersionNumber gets space, chart name and version number
func getSpaceChartNameAndVersionNumber(ctx context.Context) (string, string, string, error) {
	space, chart, err := getSpaceAndChartName(ctx)
	if err != nil {
		return "", "", "", err
	}
	version, err := getVersionNumber(ctx)
	if err != nil {
		return "", "", "", err
	}
	return space, chart, version, nil
}

// getRequestPath returns the path of request
func getRequestPath(ctx context.Context) (string, error) {
	request, err := getRequestFromContext(ctx)
	if err != nil {
		return "", err
	}
	return request.Request.URL.Path, nil
}

// getPaging gets paging info from context and return start and limit
func getPaging(ctx context.Context) (int, int, error) {
	request, err := getRequestFromContext(ctx)
	if err != nil {
		return 0, 0, err
	}
	const startName = "start"
	start := request.QueryParameter(startName)
	if len(start) <= 0 {
		start = "0"
	}
	const limitName = "limit"
	limit := request.QueryParameter(limitName)
	if len(limit) <= 0 {
		limit = "10"
	}
	s, err := strconv.Atoi(start)
	if err != nil {
		return 0, 0, errors.ErrorParamTypeError.Format(startName, "number", "string")
	}
	l, err := strconv.Atoi(limit)
	if err != nil {
		return 0, 0, errors.ErrorParamTypeError.Format(limitName, "number", "string")
	}
	return s, l, nil
}

// listStrings used by ListSpaces, ListCharts
func listStrings(ctx context.Context, name string, f func() ([]string, error)) (int, []string, error) {
	start, limit, err := getPaging(ctx)
	if err != nil {
		return 0, nil, err
	}
	strings, err := f()
	if err != nil {
		return 0, nil, err
	}
	total := len(strings)
	if start >= total {
		return total, []string{}, nil
	}
	end := start + limit
	if end > total {
		end = total
	}
	return total, strings[start:end], nil
}

// readDataFromBody reads data from the body of request
func readDataFromBody(ctx context.Context) ([]byte, error) {
	request, err := getRequestFromContext(ctx)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(request.Request.Body)
	if err != nil {
		return nil, errors.ErrorInvalidParam.Format("config", string(data))
	}
	return data, nil
}

// getChartConfig gets a config
func getChartConfig(ctx context.Context) (*types.OrchestrationConfig, error) {
	data, err := readDataFromBody(ctx)
	if err != nil {
		return nil, err
	}
	config := &types.OrchestrationConfig{}
	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, errors.ErrorParamTypeError.Format("config", "orchestration config", "unknown")
	}
	if err = config.Validate(); err != nil {
		return nil, err
	}
	space, err := getSpaceName(ctx)
	if err != nil {
		return nil, err
	}
	config.Save.Space = space
	return config, err
}

// managerCallback is used for pass space,chart and version
type managerCallback func(space storage.Space, chart storage.Chart, version storage.Version) error

// managerHelper is helper for getting a space/chart/version info from ctx
func managerHelper(ctx context.Context, f managerCallback) error {
	spaceName, chartName, versionNumber, err := getSpaceChartNameAndVersionNumber(ctx)
	if err != nil {
		return err
	}
	space, chart, version, err := common.GetSpaceChartAndVersion(ctx, spaceName, chartName, versionNumber)
	if err != nil {
		return err
	}
	return f(space, chart, version)
}
