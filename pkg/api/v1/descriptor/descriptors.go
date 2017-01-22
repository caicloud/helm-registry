/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package descriptor

import "github.com/caicloud/helm-registry/pkg/api/definition"

// Descriptors describes api info
var Descriptors []definition.Descriptor

func registerDescriptors(descriptors []definition.Descriptor) {
	// add common StatusCode
	for _, desc := range descriptors {
		for _, handler := range desc.Handlers {
			handler.StatusCode = append(handler.StatusCode,
				definition.StatusCode{Code: 400, Message: "Request params error"},
				definition.StatusCode{Code: 404, Message: "Resource does not exist"},
				definition.StatusCode{Code: 409, Message: "Conflict. See logs and response"},
				definition.StatusCode{Code: 500, Message: "Internal error. See logs"},
			)
		}
	}
	Descriptors = append(Descriptors, descriptors...)
}
