/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package rest

import (
	"net/http"

	"github.com/caicloud/helm-registry/pkg/errors"
)

// client common error definition
var (
	// ErrorUnknownLocalError defines unknown error
	ErrorUnknownLocalError = errors.NewFormatError(http.StatusBadRequest, errors.ReasonLocal, "%s")
	// ErrorNoResponse defines that can't get a response
	ErrorNoResponse = errors.NewFormatError(http.StatusNotFound, errors.ReasonLocal, "%s")
	// ErrorBadRequest defines a bad request error
	ErrorBadRequest = errors.NewFormatError(http.StatusBadRequest, errors.ReasonRequest, "%s")
	// ErrorNotFound defines that a resource not found
	ErrorNotFound = errors.NewFormatError(http.StatusNotFound, errors.ReasonServer, "%s")
	// ErrorConflict defines that a resource conflict
	ErrorConflict = errors.NewFormatError(http.StatusConflict, errors.ReasonRequest, "%s")
	// ErrorLocked defines that a resource is locked
	ErrorLocked = errors.NewFormatError(http.StatusLocked, errors.ReasonLocking, "%s")
	// ErrorServer defines server error
	ErrorServer = errors.NewFormatError(http.StatusInternalServerError, errors.ReasonServer, "%s")
	// ErrorParamValueError defines param value error
	ErrorParamValueError = errors.ErrorParamValueError
	// ErrorParamTypeError defines param type error
	ErrorParamTypeError = errors.ErrorParamTypeError
)
