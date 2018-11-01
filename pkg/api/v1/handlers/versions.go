/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/caicloud/helm-registry/pkg/api/models"
	"github.com/caicloud/helm-registry/pkg/common"
	"github.com/caicloud/helm-registry/pkg/errors"
	"github.com/caicloud/helm-registry/pkg/storage"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

// ListVersions lists versions in specified chart
func ListVersions(ctx context.Context) (int, []string, error) {
	spaceName, chartName, err := getSpaceAndChartName(ctx)
	if err != nil {
		return 0, nil, err
	}
	chart, err := common.GetChart(ctx, spaceName, chartName)
	if err != nil {
		return 0, nil, err
	}
	return listStrings(ctx, func() ([]string, error) {
		return chart.List(ctx)
	})
}

// DownloadVersion handles a request for getting a version of chart
func DownloadVersion(ctx context.Context) (data []byte, err error) {
	err = managerHelper(ctx, func(space storage.Space, chart storage.Chart, version storage.Version) error {
		data, err = version.GetContent(ctx)
		return err
	})
	return
}

// UpdateVersion handles a request for updating a version of chart. Resource must exist
func UpdateVersion(ctx context.Context) (*models.ChartLink, error) {
	return putVersion(ctx, func(space storage.Space, chart storage.Chart, version storage.Version) error {
		if !version.Exists(ctx) {
			return errors.NewResponError(http.StatusNotFound, "content.unfound", "${name} not found", errors.M{
				"name": fmt.Sprintf("%s/%s/%s", space.Name(), chart.Name(), version.Number()),
			})
		}
		return nil
	})
}

// putVersion handles a version of chart from ctx. canSave is a function and decides whether
// saves the version. If canSave returns nil, putVersion saves the version.
func putVersion(ctx context.Context, canSave managerCallback) (link *models.ChartLink, errx error) {
	errx = managerHelper(ctx, func(space storage.Space, chart storage.Chart, version storage.Version) error {
		data, err := getChartFileData(ctx)
		if err != nil {
			return err
		}
		metadata, err := getMetadataFromArchiveData(data)
		if err != nil {
			return err
		}
		// check chart name and version number
		if metadata.Name != chart.Name() {
			return errors.NewResponError(http.StatusBadRequest, "param.value", "${name} value error", errors.M{
				"name": "chart",
			})
		}
		if metadata.Version != version.Number() {
			return errors.NewResponError(http.StatusBadRequest, "param.value", "${name} value error", errors.M{
				"name": "version",
			})
		}
		// check whether can save
		if err = canSave(space, chart, version); err != nil {
			return err
		}
		err = version.PutContent(ctx, data)
		if err != nil {
			return err
		}
		// construct a chart self-link
		path, err := getRequestPath(ctx)
		if err != nil {
			return err
		}
		link = models.NewChartLink(space.Name(), chart.Name(), version.Number(), path)
		return nil
	})
	return
}

// DeleteVersion deletes specified version
func DeleteVersion(ctx context.Context) error {
	return managerHelper(ctx, func(space storage.Space, chart storage.Chart, version storage.Version) error {
		return chart.Delete(ctx, version.Number())
	})
}

// getChartFileData gets chart file from ctx
func getChartFileData(ctx context.Context) ([]byte, error) {
	request, err := getRequestFromContext(ctx)
	if err != nil {
		return nil, err
	}
	file, _, err := request.Request.FormFile(common.HTTPRequestUploadFileName)
	if err != nil {
		return nil, errors.NewResponError(http.StatusBadRequest, "param.unfound", "${name} unfound", errors.M{
			"name": common.HTTPRequestUploadFileName,
		})
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.NewResponError(http.StatusBadRequest, "param.invalidate", "${name} invalidate", errors.M{
			"name": common.HTTPRequestUploadFileName,
		})
	}
	return data, nil
}

// getMetadataFromArchiveData gets metadata from chart data
func getMetadataFromArchiveData(data []byte) (*chart.Metadata, error) {
	// TODO(optimization): Need not load whole chart
	chart, err := chartutil.LoadArchive(bytes.NewReader(data))
	if err != nil {
		return nil, errors.NewResponError(http.StatusBadRequest, "param.error", "${name} error", errors.M{
			"name": common.HTTPRequestUploadFileName,
		})
	}
	return chart.Metadata, nil
}
