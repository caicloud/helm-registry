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
	registerDescriptors(charts)
}

// charts descriptors
var charts = []definition.Descriptor{
	{
		Path: "/spaces/{space}/charts",
		Handlers: []definition.Handler{
			{
				HTTPMethod: http.MethodGet,
				Handler:    definition.NewHandlerDecoration(definition.VerbList, handlers.ListCharts).Handle,
				Doc:        "List all charts in space",
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
					definition.StatusCode{Code: http.StatusOK, Message: "Success and respond with a array of chart names",
						Sample: &models.ListResponse{
							Metadata: models.Metadata{
								Total:       10,
								ItemsLength: 1,
							},
							Items: []string{
								"chartName",
							},
						}},
				},
			},
			{
				HTTPMethod: http.MethodPost,
				Handler:    definition.NewHandlerDecoration(definition.VerbCreate, handlers.CreateChart).Handle,
				Doc:        "Create a chart by config",
				Note: `
The hanlder regards whole request body as an orchestration config. The config should be an json string.
An config sample:
{
    "save":{                            // key, required
        "chart":"chart name",           // string, required
        "version":"1.0.0",              // string, required    
		"desc":"description"            // string, optional
    },
    "configs":{                         // key, required
        "package":{                     // key, required
            "independent":true,         // boolean, required
            "space":"space name",       // string, required
            "chart":"chart name",       // string, required
            "version":"version number"  // string, required
        },
        "_config": {                    // key, required
        // root chart config
        },
        "chartB": {
            "package":{
                "independent":true,        
                "space":"space name",
                "chart":"chart name",
                "version":"version number"
            },
            "_config": {
                // chartB config
            },
            "chartD":{
                "package":{
                    "independent":false,
                    "space":"space name",
                    "chart":"chart name",
                    "version":"version number"
                },
                "_config": {
                    // chartD config
                }
            }
        },
        "chartC": {
            "package":{
                "independent":false,
                "space":"space name",
                "chart":"chart name",
                "version":"version number"
            },
            "_config": {
                // chartC config
            }
        }
    }
}
`,
				StatusCode: []definition.StatusCode{
					definition.StatusCode{Code: http.StatusCreated, Message: "Create successfully",
						Sample: &models.ChartLink{
							Space:   "spaceName",
							Chart:   "chartName",
							Version: "1.0.0",
							Link:    "/spaces/spaceName/charts/chartName/versions/1.0.0",
						}},
				},
			},
		},
	},
	{
		Path: "/spaces/{space}/charts/{chart}",
		Handlers: []definition.Handler{
			{
				HTTPMethod: http.MethodDelete,
				Handler:    definition.NewHandlerDecoration(definition.VerbDelete, handlers.DeleteChart).Handle,
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
				},
				StatusCode: []definition.StatusCode{
					definition.StatusCode{Code: http.StatusNoContent, Message: "Delete successfully"},
				},
			},
		},
	},
}
