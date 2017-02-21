/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package handlers

import (
	"context"

	"github.com/caicloud/helm-registry/pkg/common"
	"github.com/caicloud/helm-registry/pkg/storage"
)

// ListMetadata lists all metadata of a chart
func ListMetadata(ctx context.Context) (int, []*storage.Metadata, error) {
	spaceName, chartName, err := getSpaceAndChartName(ctx)
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
	start, limit, err := getPaging(ctx)
	if err != nil {
		return 0, nil, err
	}
	total := len(metadata)
	start, end := standardizeRange(total, start, limit)
	return total, metadata[start:end], nil
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
