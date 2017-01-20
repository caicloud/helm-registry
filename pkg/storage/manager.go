/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package storage

import (
	"context"

	"k8s.io/helm/pkg/proto/hapi/chart"
)

// base defines common methods
type base interface {
	// Name returns name of instance
	Name() string
}

// SpaceManager defines methods for managing chart spaces
type SpaceManager interface {
	base

	// Create creates specific space
	Create(ctx context.Context, space string) (Space, error)

	// Delete deletes specific space.
	Delete(ctx context.Context, space string) error

	// List lists all space names in current space manager
	List(ctx context.Context) ([]string, error)

	// Space returns a Space to manage specific space
	Space(ctx context.Context, space string) (Space, error)
}

// Space defines methods for managing specific chart space
type Space interface {
	base

	// Delete deletes specific chart
	Delete(ctx context.Context, chart string) error

	// List lists all chart names in current space
	List(ctx context.Context) ([]string, error)

	// Charts returns all metadatas in the current space
	Charts(ctx context.Context) ([]*chart.Metadata, error)

	// Chart returns a Chart for managing specific chart
	Chart(ctx context.Context, chart string) (Chart, error)
}

// Chart defines methods for managing specific chart
type Chart interface {
	base

	// Delete deletes specific chart
	Delete(ctx context.Context, version string) error

	// List lists all version numbers in current chart
	List(ctx context.Context) ([]string, error)

	// Versions returns all metadatas in the current chart
	Versions(ctx context.Context) ([]*chart.Metadata, error)

	// Version returns a Version for managing specific version
	Version(ctx context.Context, version string) (Version, error)
}

// Version defines methods for managing specific version of a chart
type Version interface {
	base

	// PutContent stores chart data
	PutContent(ctx context.Context, data []byte) error

	// GetContent gets chart data
	GetContent(ctx context.Context) ([]byte, error)

	// Validate validates whether the chart is valid
	Validate(ctx context.Context) error

	// Metadata returns a Metadata of the current chart
	Metadata(ctx context.Context) (*chart.Metadata, error)

	// Values gets data from values.yaml file which in the current chart data
	Values(ctx context.Context) ([]byte, error)
}
