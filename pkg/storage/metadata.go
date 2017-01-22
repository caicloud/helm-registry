/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package storage

import (
	"k8s.io/helm/pkg/proto/hapi/chart"
)

// Metadata describes the dependencies of chart metadata
type Metadata struct {
	chart.Metadata
	Dependencies []*Metadata `json:"dependencies,omitempty"`
}

// CoalesceMetadata coalesces all metadata in chart
func CoalesceMetadata(chart *chart.Chart) (*Metadata, error) {
	metadata := &Metadata{}
	metadata.Metadata = *chart.Metadata
	for _, dep := range chart.Dependencies {
		m, err := CoalesceMetadata(dep)
		if err != nil {
			return nil, err
		}
		metadata.Dependencies = append(metadata.Dependencies, m)
	}
	return metadata, nil
}
