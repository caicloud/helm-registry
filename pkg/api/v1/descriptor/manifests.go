/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package descriptor

import (
	"net/http"

	"github.com/caicloud/helm-registry/pkg/api/definition"
	"github.com/caicloud/helm-registry/pkg/api/models"
	"github.com/caicloud/helm-registry/pkg/api/v1/handlers"
	"github.com/caicloud/helm-registry/pkg/common"
	"github.com/caicloud/helm-registry/pkg/storage"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

func init() {
	registerDescriptors(manifests)
}

// manifests descriptors
var manifests = []definition.Descriptor{
	{
		Path: "/spaces/{space}/metadata",
		Handlers: []definition.Handler{
			{
				HTTPMethod: http.MethodGet,
				Handler:    definition.NewHandlerDecoration(definition.VerbList, handlers.ListMetadataInSpace).Handle,
				Doc:        "List all metadata in a space",
				PathParams: []definition.Param{
					{
						Name:     "space",
						Type:     "string",
						Doc:      "space name",
						Required: true,
					},
				},
				QueryParams: []definition.Param{
					{
						Name:     "start",
						Type:     "number",
						Doc:      "Query start index",
						Required: false,
						Default:  0,
					},
					{
						Name:     "limit",
						Type:     "number",
						Doc:      "Specify the number of records to return",
						Required: false,
						Default:  common.DefaultPagingLimit,
					},
				},
				StatusCode: []definition.StatusCode{
					definition.StatusCode{Code: http.StatusOK, Message: "Success and respond with a array of metadata",
						Sample: &models.ListResponse{
							Metadata: models.Metadata{
								Total:       10,
								ItemsLength: 1,
							},
							Items: []*storage.Metadata{
								{
									Metadata: chart.Metadata{
										Name:        "A",
										Version:     "1.0.0",
										Description: "A chart named A and has dependency with ChartB",
									},
									Dependencies: []*storage.Metadata{
										{
											Metadata: chart.Metadata{
												Name:        "B",
												Version:     "2.1.0",
												Description: "A chart is named B",
											},
										},
									},
								},
							},
						}},
				},
			},
		},
	},
	{
		Path: "/spaces/{space}/metadata/latest",
		Handlers: []definition.Handler{
			{
				HTTPMethod: http.MethodGet,
				Handler:    definition.NewHandlerDecoration(definition.VerbList, handlers.ListLatestMetadataInSpace).Handle,
				Doc:        "List latest metadata in a space",
				PathParams: []definition.Param{
					{
						Name:     "space",
						Type:     "string",
						Doc:      "space name",
						Required: true,
					},
				},
				QueryParams: []definition.Param{
					{
						Name:     "start",
						Type:     "number",
						Doc:      "Query start index",
						Required: false,
						Default:  0,
					},
					{
						Name:     "limit",
						Type:     "number",
						Doc:      "Specify the number of records to return",
						Required: false,
						Default:  common.DefaultPagingLimit,
					},
				},
				StatusCode: []definition.StatusCode{
					definition.StatusCode{Code: http.StatusOK, Message: "Success and respond with a array of latest metadata",
						Sample: &models.ListResponse{
							Metadata: models.Metadata{
								Total:       10,
								ItemsLength: 1,
							},
							Items: []*storage.Metadata{
								{
									Metadata: chart.Metadata{
										Name:        "A",
										Version:     "1.0.0",
										Description: "A chart named A and has dependency with ChartB",
									},
									Dependencies: []*storage.Metadata{
										{
											Metadata: chart.Metadata{
												Name:        "B",
												Version:     "2.1.0",
												Description: "A chart is named B",
											},
										},
									},
								},
							},
						}},
				},
			},
		},
	},
	{
		Path: "/spaces/{space}/charts/{chart}/metadata",
		Handlers: []definition.Handler{
			{
				HTTPMethod: http.MethodGet,
				Handler:    definition.NewHandlerDecoration(definition.VerbList, handlers.ListMetadataInChart).Handle,
				Doc:        "List all metadata in a chart",
				PathParams: []definition.Param{
					{
						Name:     "space",
						Type:     "string",
						Doc:      "space name",
						Required: true,
					},
					{
						Name:     "chart",
						Type:     "string",
						Doc:      "chart name",
						Required: true,
					},
				},
				QueryParams: []definition.Param{
					{
						Name:     "start",
						Type:     "number",
						Doc:      "Query start index",
						Required: false,
						Default:  0,
					},
					{
						Name:     "limit",
						Type:     "number",
						Doc:      "Specify the number of records to return",
						Required: false,
						Default:  common.DefaultPagingLimit,
					},
				},
				StatusCode: []definition.StatusCode{
					definition.StatusCode{Code: http.StatusOK, Message: "Success and respond with a array of metadata",
						Sample: &models.ListResponse{
							Metadata: models.Metadata{
								Total:       10,
								ItemsLength: 1,
							},
							Items: []*storage.Metadata{
								{
									Metadata: chart.Metadata{
										Name:        "A",
										Version:     "1.0.0",
										Description: "A chart named A and has dependency with ChartB",
									},
									Dependencies: []*storage.Metadata{
										{
											Metadata: chart.Metadata{
												Name:        "B",
												Version:     "2.1.0",
												Description: "A chart is named B",
											},
										},
									},
								},
							},
						}},
				},
			},
		},
	},
	{
		Path: "/spaces/{space}/charts/{chart}/metadata/latest",
		Handlers: []definition.Handler{
			{
				HTTPMethod: http.MethodGet,
				Handler:    definition.NewHandlerDecoration(definition.VerbGet, handlers.GetLatestMetadataInChart).Handle,
				Doc:        "Get metadata of the latest version in a chart",
				PathParams: []definition.Param{
					{
						Name:     "space",
						Type:     "string",
						Doc:      "space name",
						Required: true,
					},
					{
						Name:     "chart",
						Type:     "string",
						Doc:      "chart name",
						Required: true,
					},
				},
				StatusCode: []definition.StatusCode{
					definition.StatusCode{Code: http.StatusOK, Message: "Success and respond with metadata of latest version",
						Sample: &storage.Metadata{
							Metadata: chart.Metadata{
								Name:        "A",
								Version:     "1.0.0",
								Description: "A chart named A and has dependency with ChartB",
							},
							Dependencies: []*storage.Metadata{
								{
									Metadata: chart.Metadata{
										Name:        "B",
										Version:     "2.1.0",
										Description: "A chart is named B",
									},
								},
							},
						}},
				},
			},
		},
	},
	{
		Path: "/spaces/{space}/charts/{chart}/versions/{version}/manifests/metadata",
		Handlers: []definition.Handler{
			{
				HTTPMethod: http.MethodGet,
				Handler:    definition.NewHandlerDecoration(definition.VerbGet, handlers.FetchMetadata).Handle,
				Doc:        "Get metadata of a version",
				PathParams: []definition.Param{
					{
						Name:     "space",
						Type:     "string",
						Doc:      "space name",
						Required: true,
					},
					{
						Name:     "chart",
						Type:     "string",
						Doc:      "chart name",
						Required: true,
					},
					{
						Name:     "version",
						Type:     "string",
						Doc:      "version number",
						Required: true,
					},
				},
				StatusCode: []definition.StatusCode{
					definition.StatusCode{Code: http.StatusOK, Message: "Success and respond with a metadata of a version",
						Sample: &storage.Metadata{
							Metadata: chart.Metadata{
								Name:        "A",
								Version:     "1.0.0",
								Description: "A chart named A and has dependency with ChartB",
							},
							Dependencies: []*storage.Metadata{
								{
									Metadata: chart.Metadata{
										Name:        "B",
										Version:     "2.1.0",
										Description: "A chart is named B",
									},
								},
							},
						}},
				},
			},
		},
	},
	{
		Path: "/spaces/{space}/charts/{chart}/versions/{version}/manifests/values",
		Handlers: []definition.Handler{
			{
				HTTPMethod: http.MethodGet,
				Handler:    definition.NewHandlerDecoration(definition.VerbGet, handlers.FetchValues).Handle,
				Doc:        "Get values of a version",
				PathParams: []definition.Param{
					{
						Name:     "space",
						Type:     "string",
						Doc:      "space name",
						Required: true,
					},
					{
						Name:     "chart",
						Type:     "string",
						Doc:      "chart name",
						Required: true,
					},
					{
						Name:     "version",
						Type:     "string",
						Doc:      "version number",
						Required: true,
					},
				},
				StatusCode: []definition.StatusCode{
					definition.StatusCode{Code: http.StatusOK, Message: "Success and respond with values of a version"},
				},
			},
		},
	},
}
