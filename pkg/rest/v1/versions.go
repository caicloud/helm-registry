/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package v1

import (
	"net/http"

	"github.com/caicloud/helm-registry/pkg/api/models"
)

// APIListVersions defines an api of listing versions
type APIListVersions struct {
	baseAPI
	// Space is the name of space
	Space string `kind:"path" name:"space"`
	// Chart is the name of chart
	Chart string `kind:"path" name:"chart"`
	// Start is the start index of list
	Start int `kind:"query" name:"start"`
	// Limit is the max length of list
	Limit int `kind:"query" name:"limit"`
}

// NewAPIListVersions creates an instance of APIListVersions
func NewAPIListVersions() *APIListVersions {
	api := &APIListVersions{}
	api.object = api
	api.method = http.MethodGet
	api.url = URLVersions
	api.result = &StringCollectionResult{}
	return api
}

// Convert converts result to *StringCollectionResult
func (api *APIListVersions) Convert(result interface{}, err error) (*StringCollectionResult, error) {
	if err != nil {
		return nil, err
	}
	return result.(*StringCollectionResult), nil
}

// APIDownloadVersion defines an api of downloading version
type APIDownloadVersion struct {
	baseAPI
	// Space is the name of space
	Space string `kind:"path" name:"space"`
	// Chart is the name of chart
	Chart string `kind:"path" name:"chart"`
	// Version is the name of Version
	Version string `kind:"path" name:"version"`
}

// NewAPIDownloadVersion creates an instance of APIDownloadVersion
func NewAPIDownloadVersion() *APIDownloadVersion {
	api := &APIDownloadVersion{}
	api.object = api
	api.method = http.MethodGet
	api.url = URLVersion
	api.result = []byte{}
	return api
}

// Convert converts result to []byte
func (api *APIDownloadVersion) Convert(result interface{}, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	return result.([]byte), nil
}

// APIUpdateVersion defines an api of updating version
type APIUpdateVersion struct {
	baseAPI
	// Space is the name of space
	Space string `kind:"path" name:"space"`
	// Chart is the name of chart
	Chart string `kind:"path" name:"chart"`
	// Version is the name of Version
	Version string `kind:"path" name:"version"`
	// ChartFile is a chart file
	ChartFile *File `kind:"file" name:"chartfile"`
}

// NewAPIUpdateVersion creates an instance of APIUpdateVersion
func NewAPIUpdateVersion() *APIUpdateVersion {
	api := &APIUpdateVersion{}
	api.object = api
	api.method = http.MethodPut
	api.url = URLVersion
	api.result = &models.ChartLink{}
	api.ChartFile = &File{}
	return api
}

// Convert converts result to *models.ChartLink
func (api *APIUpdateVersion) Convert(result interface{}, err error) (*models.ChartLink, error) {
	if err != nil {
		return nil, err
	}
	return result.(*models.ChartLink), nil
}

// APIDeleteVersion defines an api of deleting version
type APIDeleteVersion struct {
	baseAPI
	// Space is the name of space
	Space string `kind:"path" name:"space"`
	// Chart is the name of chart
	Chart string `kind:"path" name:"chart"`
	// Version is the name of Version
	Version string `kind:"path" name:"version"`
}

// APICreateVersion creates an instance of APICreateVersion
func NewAPIDeleteVersion() *APIDeleteVersion {
	api := &APIDeleteVersion{}
	api.object = api
	api.method = http.MethodDelete
	api.url = URLVersion
	return api
}

// Convert converts result to *models.Link
func (api *APIDeleteVersion) Convert(result interface{}, err error) error {
	return err
}
