/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package handlers

import (
	"context"
	"strings"

	"github.com/caicloud/helm-registry/pkg/common"
	"github.com/caicloud/helm-registry/pkg/errors"
	"github.com/caicloud/helm-registry/pkg/storage"
)

// ListMetadataInSpace lists all metadata in a space
func ListMetadataInSpace(ctx context.Context) (int, []*storage.Metadata, error) {
	spaceName, err := getSpaceName(ctx)
	if err != nil {
		return 0, nil, err
	}
	start, limit, err := getPaging(ctx)
	if err != nil {
		return 0, nil, err
	}
	space, err := common.GetSpace(ctx, spaceName)
	if err != nil {
		return 0, nil, err
	}
	metadata, err := space.VersionMetadata(ctx)
	if err != nil {
		return 0, nil, err
	}
	total := len(metadata)
	start, end := standardizeRange(total, start, limit)
	return total, metadata[start:end], nil
}

// ListLatestMetadataInSpace lists all metadata of the latest version of charts in space
func ListLatestMetadataInSpace(ctx context.Context) (int, []*storage.Metadata, error) {
	spaceName, err := getSpaceName(ctx)
	if err != nil {
		return 0, nil, err
	}
	start, limit, err := getPaging(ctx)
	if err != nil {
		return 0, nil, err
	}
	kind, sub, err := getFilterCondition(ctx)
	if err != nil {
		return 0, nil, err
	}
	space, err := common.GetSpace(ctx, spaceName)
	if err != nil {
		return 0, nil, err
	}
	chartNames, err := space.List(ctx)
	if err != nil {
		return 0, nil, err
	}
	metadata := make([]*storage.Metadata, 0, len(chartNames))
	for _, chartName := range chartNames {
		if len(sub) != 0 && !strings.Contains(chartName, sub) {
			continue
		}
		md, err := getLatestMetadata(ctx, spaceName, chartName)
		if err != nil {
			return 0, nil, err
		}
		if len(kind) != 0 && md.Type != kind {
			continue
		}
		metadata = append(metadata, md)
	}

	total := len(metadata)
	start, end := standardizeRange(total, start, limit)
	return total, metadata[start:end], nil
}

// ListMetadataInChart lists all metadata in a chart
func ListMetadataInChart(ctx context.Context) (int, []*storage.Metadata, error) {
	spaceName, chartName, err := getSpaceAndChartName(ctx)
	if err != nil {
		return 0, nil, err
	}
	start, limit, err := getPaging(ctx)
	if err != nil {
		return 0, nil, err
	}
	chart, err := common.GetChart(ctx, spaceName, chartName)
	if err != nil {
		return 0, nil, err
	}
	// get all metadata of versions
	metadata, err := chart.VersionMetadata(ctx)
	if err != nil {
		return 0, nil, err
	}
	total := len(metadata)
	start, end := standardizeRange(total, start, limit)
	return total, metadata[start:end], nil
}

// GetLatestMetadataInChart gets metadata of the latest version in a chart
func GetLatestMetadataInChart(ctx context.Context) (metadata *storage.Metadata, err error) {
	spaceName, chartName, err := getSpaceAndChartName(ctx)
	if err != nil {
		return nil, err
	}
	return getLatestMetadata(ctx, spaceName, chartName)
}

// FetchMetadata fetches metadata of specified version
func FetchMetadata(ctx context.Context) (metadata *storage.Metadata, err error) {
	err = managerHelper(ctx, func(space storage.Space, chart storage.Chart, version storage.Version) error {
		metadata, err = version.Metadata(ctx)
		return err
	})
	return
}

// FetchValues fetches values of specified version
func FetchValues(ctx context.Context) (data []byte, err error) {
	err = managerHelper(ctx, func(space storage.Space, chart storage.Chart, version storage.Version) error {
		data, err = version.Values(ctx)
		return err
	})
	return
}

// getLatestMetadata gets latest metadata in a chart
func getLatestMetadata(ctx context.Context, spaceName, chartName string) (metadata *storage.Metadata, err error) {
	chart, err := common.GetChart(ctx, spaceName, chartName)
	if err != nil {
		return nil, err
	}
	versionNumbers, err := chart.List(ctx)
	if err != nil {
		return nil, err
	}
	if len(versionNumbers) <= 0 {
		return nil, errors.ErrorContentNotFound.Format("metadata")
	}
	version, err := chart.Version(ctx, versionNumbers[len(versionNumbers)-1])
	if err != nil {
		return nil, err
	}
	return version.Metadata(ctx)
}
