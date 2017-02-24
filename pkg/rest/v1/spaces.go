/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package v1

import (
	"net/http"

	"github.com/caicloud/helm-registry/pkg/api/models"
)

// APIListSpace defines an api of listing spaces
type APIListSpaces struct {
	baseAPI
	// Start is the start index of list
	Start int `kind:"query" name:"start"`
	// Limit is the max length of list
	Limit int `kind:"query" name:"limit"`
}

// NewAPIListSpaces creates an instance of APIListSpace
func NewAPIListSpaces() *APIListSpaces {
	api := &APIListSpaces{}
	api.object = api
	api.method = http.MethodGet
	api.url = URLSpaces
	api.result = &StringCollectionResult{}
	return api
}

// Convert converts result to *StringCollectionResult
func (api *APIListSpaces) Convert(result interface{}, err error) (*StringCollectionResult, error) {
	if err != nil {
		return nil, err
	}
	return result.(*StringCollectionResult), nil
}

// APICreateSpace defines an api of creating space
type APICreateSpace struct {
	baseAPI
	// Space is the name of space
	Space string `kind:"query" name:"space"`
}

// APICreateSpace creates an instance of APICreateSpace
func NewAPICreateSpace() *APICreateSpace {
	api := &APICreateSpace{}
	api.object = api
	api.method = http.MethodPost
	api.url = URLSpaces
	api.result = &models.Link{}
	return api
}

// Convert converts result to *models.Link
func (api *APICreateSpace) Convert(result interface{}, err error) (*models.Link, error) {
	if err != nil {
		return nil, err
	}
	return result.(*models.Link), nil
}

// APIDeleteSpace defines an api of deleting space
type APIDeleteSpace struct {
	baseAPI
	// Space is the name of space
	Space string `kind:"path" name:"space"`
}

// APICreateSpace creates an instance of APICreateSpace
func NewAPIDeleteSpace() *APIDeleteSpace {
	api := &APIDeleteSpace{}
	api.object = api
	api.method = http.MethodDelete
	api.url = URLSpace
	return api
}

// Convert converts result to *models.Link
func (api *APIDeleteSpace) Convert(result interface{}, err error) error {
	return err
}
