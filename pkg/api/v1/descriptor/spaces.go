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
	registerDescriptors(spaces)
}

// spaces descriptors
var spaces = []definition.Descriptor{
	{
		Path: "/spaces",
		Handlers: []definition.Handler{
			{
				HTTPMethod: http.MethodGet,
				Handler:    definition.NewHandlerDecoration(definition.VerbList, handlers.ListSpaces).Handle,
				Doc:        "List all spaces in the registry",
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
					definition.StatusCode{Code: http.StatusOK, Message: "Success and respond with a array of space names",
						Sample: &models.ListResponse{
							Metadata: models.Metadata{
								Total:       10,
								ItemsLength: 1,
							},
							Items: []string{
								"spaceName",
							},
						}},
				},
			},
			{
				HTTPMethod: http.MethodPost,
				Handler:    definition.NewHandlerDecoration(definition.VerbCreate, handlers.CreateSpace).Handle,
				Doc:        "Create a space",
				QueryParams: []definition.Param{
					{
						Name:     "space",
						Type:     "string",
						Doc:      "space name",
						Required: true,
					},
				},
				StatusCode: []definition.StatusCode{
					definition.StatusCode{Code: http.StatusCreated, Message: "Create successfully",
						Sample: &models.Link{
							Name: "spaceName",
							Link: "/spaces/spaceName",
						}},
				},
			},
		},
	},
	{
		Path: "/spaces/{space}",
		Handlers: []definition.Handler{
			{
				HTTPMethod: http.MethodDelete,
				Handler:    definition.NewHandlerDecoration(definition.VerbDelete, handlers.DeleteSpace).Handle,
				Doc:        "Delete space",
				PathParams: []definition.Param{
					{
						Name:     "space",
						Type:     "string",
						Doc:      "space name",
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
