/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/caicloud/helm-registry/pkg/api/models"
	"github.com/caicloud/helm-registry/pkg/rest"
	"github.com/caicloud/helm-registry/pkg/storage"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

// Client is a registry client for managing registry server
type Client struct {
	rest.Client
}

// NewClient creates a new registry client. endpoint is the address of server
func NewClient(endpoint string) (*Client, error) {
	client := &Client{rest.NewUniversalClient(strings.TrimRight(endpoint, "\\/"))}
	return client, nil
}

// NewTransportClient creates a new registry client. endpoint is the address of server with transport
func NewTransportClient(endpoint string, transport http.RoundTripper) (*Client, error) {
	client := &Client{rest.NewUniversalTransportClient(strings.TrimRight(endpoint, "\\/"), transport)}
	return client, nil
}

// ListSpaces lists spaces
func (c *Client) ListSpaces(start, limit int) (*StringCollectionResult, error) {
	api := NewAPIListSpaces()
	api.Start = start
	api.Limit = limit
	return api.Convert(c.Do(api))
}

// CreateSpace creates a space by space name
func (c *Client) CreateSpace(spaceName string) (*models.Link, error) {
	api := NewAPICreateSpace()
	api.Space = spaceName
	return api.Convert(c.Do(api))
}

// DeleteSpace deletes a space by space name
func (c *Client) DeleteSpace(spaceName string) error {
	api := NewAPIDeleteSpace()
	api.Space = spaceName
	return api.Convert(c.Do(api))
}

// ListCharts lists charts in the space
func (c *Client) ListCharts(spaceName string, start, limit int) (*StringCollectionResult, error) {
	api := NewAPIListCharts()
	api.Space = spaceName
	api.Start = start
	api.Limit = limit
	return api.Convert(c.Do(api))
}

// CreateChart creates a chart by config. config is a json string to specify the hierarchical structure of chart.
// Please refer to the descriptor of creating chart.
func (c *Client) CreateChart(spaceName string, config string) (*models.ChartLink, error) {
	api := NewAPICreateChart()
	api.Space = spaceName
	api.Config = config
	return api.Convert(c.Do(api))
}

// UploadChart uploads a chart file. If the chart exists, it produces an error.
func (c *Client) UploadChart(spaceName string, data []byte) (*models.ChartLink, error) {
	api := NewAPIUploadChart()
	api.Space = spaceName
	api.ChartFile.Data = data
	return api.Convert(c.Do(api))
}

// DeleteChart deletes a chart and its all versions
func (c *Client) DeleteChart(spaceName string, chartName string) error {
	api := NewAPIDeleteChart()
	api.Space = spaceName
	api.Chart = chartName
	return api.Convert(c.Do(api))
}

// ListVersions lists versions of the chart
func (c *Client) ListVersions(spaceName string, chartName string, start, limit int) (*StringCollectionResult, error) {
	api := NewAPIListVersions()
	api.Space = spaceName
	api.Chart = chartName
	api.Start = start
	api.Limit = limit
	return api.Convert(c.Do(api))
}

// DownloadVersion downloads a chart file
func (c *Client) DownloadVersion(spaceName string, chartName string, versionNumber string) ([]byte, error) {
	api := NewAPIDownloadVersion()
	api.Space = spaceName
	api.Chart = chartName
	api.Version = versionNumber
	return api.Convert(c.Do(api))
}

// UpdateVersion updates a chart file. If the chart does not exist, it produces an error.
func (c *Client) UpdateVersion(spaceName string, chartName string, versionNumber string, data []byte) (*models.ChartLink, error) {
	api := NewAPIUpdateVersion()
	api.Space = spaceName
	api.Chart = chartName
	api.Version = versionNumber
	api.ChartFile.Data = data
	return api.Convert(c.Do(api))
}

// DeleteChart deletes a version of chart
func (c *Client) DeleteVersion(spaceName string, chartName string, versionNumber string) error {
	api := NewAPIDeleteVersion()
	api.Space = spaceName
	api.Chart = chartName
	api.Version = versionNumber
	return api.Convert(c.Do(api))
}

// FetchChartMetadata fetches all metadata of chart
func (c *Client) FetchChartMetadata(spaceName string, chartName string, start, limit int) (*MetadataCollectionResult, error) {
	api := NewAPIFetchChartMetadata()
	api.Space = spaceName
	api.Chart = chartName
	api.Start = start
	api.Limit = limit
	return api.Convert(c.Do(api))
}

// FetchVersionMetadata fetches metadata of version
func (c *Client) FetchVersionMetadata(spaceName string, chartName string, versionNumber string) (*storage.Metadata, error) {
	api := NewAPIFetchVersionMetadata()
	api.Space = spaceName
	api.Chart = chartName
	api.Version = versionNumber
	return api.Convert(c.Do(api))
}

// UpdateVersionMetadata updates metadata of version
func (c *Client) UpdateVersionMetadata(spaceName string, chartName string, versionNumber string, metadata *chart.Metadata) (*storage.Metadata, error) {
	data, err := json.Marshal(metadata)
	if err != nil {
		return nil, rest.ErrorUnknownLocalError.Format(err.Error())
	}
	api := NewAPIUpdateVersionMetadata()
	api.Space = spaceName
	api.Chart = chartName
	api.Version = versionNumber
	api.Metadata = data
	return api.Convert(c.Do(api))
}

// FetchVersionValues fetches values of version
func (c *Client) FetchVersionValues(spaceName string, chartName string, versionNumber string) ([]byte, error) {
	api := NewAPIFetchVersionValues()
	api.Space = spaceName
	api.Chart = chartName
	api.Version = versionNumber
	return api.Convert(c.Do(api))
}

// UpdateVersionValues updates values of version
func (c *Client) UpdateVersionValues(spaceName string, chartName string, versionNumber string, values []byte) ([]byte, error) {
	api := NewAPIUpdateVersionValues()
	api.Space = spaceName
	api.Chart = chartName
	api.Version = versionNumber
	api.Values = values
	return api.Convert(c.Do(api))
}
