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
)

func init() {
	registerDescriptors(versions)
}

// versions descriptors
var versions = []definition.Descriptor{
	{
		Path: "/spaces/{space}/charts/{chart}/versions",
		Handlers: []definition.Handler{
			{
				HTTPMethod: http.MethodGet,
				Handler:    definition.NewHandlerDecoration(definition.VerbList, handlers.ListVersions).Handle,
				Doc:        "List all versions in a chart",
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
					definition.StatusCode{Code: http.StatusOK, Message: "Success and respond with a array of version numbers",
						Sample: &models.ListResponse{
							Metadata: models.Metadata{
								Total:       10,
								ItemsLength: 1,
							},
							Items: []string{
								"versionNumber",
							},
						}},
				},
			},
		},
	},
	{
		Path: "/spaces/{space}/charts/{chart}/versions/{version}",
		Handlers: []definition.Handler{
			{
				HTTPMethod: http.MethodGet,
				Handler:    definition.NewHandlerDecoration(definition.VerbGet, handlers.DownloadVersion).Handle,
				Doc:        "Download a version of a chart",
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
					definition.StatusCode{Code: http.StatusOK, Message: "Download with an archive file of chart"},
				},
			},
			{
				HTTPMethod: http.MethodPut,
				Handler:    definition.NewHandlerDecoration(definition.VerbUpdate, handlers.UpdateVersion).Handle,
				Doc:        "Update a version of a chart",
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
				QueryParams: []definition.Param{
					{
						Name:     "chartfile",
						Type:     "multipart/form-data",
						Doc:      "An archive file of chart",
						Required: true,
					},
				},
				StatusCode: []definition.StatusCode{
					definition.StatusCode{Code: http.StatusOK, Message: "Update successfully",
						Sample: &models.ChartLink{
							Space:   "spaceName",
							Chart:   "chartName",
							Version: "1.0.0",
							Link:    "/spaces/spaceName/charts/chartName/versions/1.0.0",
						}},
				},
			},
			{
				HTTPMethod: http.MethodDelete,
				Handler:    definition.NewHandlerDecoration(definition.VerbDelete, handlers.DeleteVersion).Handle,
				Doc:        "Delete a chart and its all versions",
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
					definition.StatusCode{Code: http.StatusNoContent, Message: "Delete successfully"},
				},
			},
		},
	},
}
