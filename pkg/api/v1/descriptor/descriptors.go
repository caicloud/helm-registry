/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package descriptor

import (
	"net/http"

	"github.com/caicloud/helm-registry/pkg/api/definition"
)

// Descriptors describes api info
var Descriptors []definition.Descriptor

func registerDescriptors(descriptors []definition.Descriptor) {
	// add common StatusCode
	for _, desc := range descriptors {
		for _, handler := range desc.Handlers {
			handler.StatusCode = append(handler.StatusCode,
				definition.StatusCode{Code: http.StatusBadRequest, Message: "Request params error"},
				definition.StatusCode{Code: http.StatusNotFound, Message: "Resource does not exist"},
				definition.StatusCode{Code: http.StatusConflict, Message: "Conflict. See logs and response"},
				definition.StatusCode{Code: http.StatusLocked, Message: "Resource locked. Can't read or write"},
				definition.StatusCode{Code: http.StatusInternalServerError, Message: "Internal error. See logs"},
			)
		}
	}
	Descriptors = append(Descriptors, descriptors...)
}
