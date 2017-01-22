/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package storage

import "context"

// ValidationType defines a type for Validating in SpaceManager
type ValidationType string

// Basic ValidationType
const (
	// ValidationTypeSpaceName is the validation type of space name
	ValidationTypeSpaceName = "SpaceName"
	// ValidationTypeChartName is the validation type of chart name
	ValidationTypeChartName = "ChartName"
	// ValidationTypeVersionNumber is the validation type of version number
	ValidationTypeVersionNumber = "VersionNumber"
)

// base defines common methods
type base interface {
	// Kind returns factory name of instance
	Kind() string
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

	// Validate validates whether the value of vType is valid.
	// An instance of SpaceManager should validate Basic ValidationType at least
	Validate(ctx context.Context, vType ValidationType, value interface{}) bool
}

// Space defines methods for managing specific chart space
type Space interface {
	base

	// Name returns name of instance
	Name() string

	// Delete deletes specific chart
	Delete(ctx context.Context, chart string) error

	// List lists all chart names in current space
	List(ctx context.Context) ([]string, error)

	// Exists returns whether the space exists
	Exists(ctx context.Context) bool

	// VersionMetadata returns all version metadata in the current space
	VersionMetadata(ctx context.Context) ([]*Metadata, error)

	// Chart returns a Chart for managing specific chart
	Chart(ctx context.Context, chart string) (Chart, error)
}

// Chart defines methods for managing specific chart
type Chart interface {
	base

	// Name returns name of instance
	Name() string

	// Delete deletes specific chart
	Delete(ctx context.Context, version string) error

	// List lists all version numbers in current chart
	List(ctx context.Context) ([]string, error)

	// Exists returns whether the chart exists
	Exists(ctx context.Context) bool

	// VersionMetadata returns all version metadata in the current chart
	VersionMetadata(ctx context.Context) ([]*Metadata, error)

	// Version returns a Version for managing specific version
	Version(ctx context.Context, version string) (Version, error)
}

// Version defines methods for managing specific version of a chart
type Version interface {
	base

	// Number returns version number
	Number() string

	// PutContent stores chart data
	PutContent(ctx context.Context, data []byte) error

	// GetContent gets chart data
	GetContent(ctx context.Context) ([]byte, error)

	// Exists returns whether the version exists
	Exists(ctx context.Context) bool

	// Validate validates whether the chart is valid and can be modified
	Validate(ctx context.Context) error

	// Metadata returns a Metadata of the current chart
	Metadata(ctx context.Context) (*Metadata, error)

	// Values gets data from values.yaml file which in the current chart data
	Values(ctx context.Context) ([]byte, error)
}
