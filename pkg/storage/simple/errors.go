/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package simple

import (
	"net/http"

	"github.com/caicloud/helm-registry/pkg/errors"
)

var (
	// ErrorContentMissing defines content must be specified
	ErrorContentMissing = errors.NewFormatError(http.StatusInternalServerError, errors.ReasonInternal, "%s does not specify")
	// ErrorNoParameter defines parameter can't be nil or empty
	ErrorNoParameter = errors.NewFormatError(http.StatusInternalServerError, errors.ReasonInternal, "%s can't be nil or empty")
	// ErrorNoResource defines can't find resource from package
	ErrorNoResource = errors.NewFormatError(http.StatusConflict, errors.ReasonInternal, "can't find %s from package: %v")
	// ErrorNeedForcedDelete defines need to force to delete resource
	ErrorNeedForcedDelete = errors.NewFormatError(http.StatusInternalServerError, errors.ReasonInternal, "need force to delete %s")
	// ErrorInvalidParam defines invalid error
	ErrorInvalidParam = errors.ErrorInvalidParam
	// ErrorResourceExist defines resource conflict error
	ErrorResourceExist = errors.ErrorResourceExist
	// ErrorInternalUnknown defines internal unknown error that we can't find a reason
	ErrorInternalUnknown = errors.ErrorInternalUnknown
	// ErrorLocking defines locking error
	ErrorLocking = errors.ErrorLocking
	// ErrorInvalidStatus defines invalid status error
	ErrorInvalidStatus = errors.ErrorInvalidStatus
	// ErrorParamTypeError defines param type error
	ErrorParamTypeError = errors.ErrorParamTypeError
	// ErrorContentNotFound defines not found error
	ErrorContentNotFound = errors.ErrorContentNotFound
)
