/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package v1

import (
	"github.com/caicloud/helm-registry/pkg/api/models"
	"github.com/caicloud/helm-registry/pkg/storage"
)

// StringCollectionResult describes a collection of []string
type StringCollectionResult struct {
	Metadata models.Metadata `json:"metadata"`
	Items    []string        `json:"items"`
}

// MetadataCollectionResult describes a collection of []*chart.Metadata
type MetadataCollectionResult struct {
	Metadata models.Metadata     `json:"metadata"`
	Items    []*storage.Metadata `json:"items"`
}
