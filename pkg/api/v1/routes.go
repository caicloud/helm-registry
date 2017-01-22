/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package v1

import (
	"github.com/caicloud/helm-registry/pkg/api/definition"
	"github.com/caicloud/helm-registry/pkg/api/v1/descriptor"
	"github.com/emicklei/go-restful"
)

// InstallRouters installs api WebService
func InstallRouters(containers *restful.Container) *restful.WebService {
	service := (&restful.WebService{}).
		ApiVersion("v1").
		Path("/api/v1").
		Doc("v1 API").
		Consumes("*/*", "application/x-www-form-urlencoded", "multipart/form-data", restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML)
	service = definition.GenerateRoutes(service, descriptor.Descriptors)
	containers.Add(service)
	return service
}
