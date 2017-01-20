/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package errors

import "net/http"

// defines reason types
const (
	// ReasonInternal is a type about internal errors
	ReasonInternal = "ReasonInternal"
	// ReasonRequest is a type about request errors
	ReasonRequest = "ReasonRequest"
	// ReasonLocking is a type about race errors
	ReasonLocking = "ResourceLocking"
	// ReasonLocal is a type about local errors (for client)
	ReasonLocal = "ReasonLocal"
	// ReasonServer is a type about server errors (for client)
	ReasonServer = "ReasonServer"
)

var (
	// ErrorParamTypeError defines param type error
	ErrorParamTypeError = NewFormatError(http.StatusBadRequest, ReasonRequest, "param %s should be %s, but got %s")
	// ErrorParamNotFound defines request param error
	ErrorParamNotFound = NewFormatError(http.StatusBadRequest, ReasonRequest, "can't find param %s in request")
	// ErrorContentNotFound defines not found error
	ErrorContentNotFound = NewFormatError(http.StatusNotFound, ReasonInternal, "content %s not found")
	// ErrorInvalidParam defines invalid error
	ErrorInvalidParam = NewFormatError(http.StatusBadRequest, ReasonRequest, "%s is invalid: %v")
	// ErrorResourceExist defines resource conflict error
	ErrorResourceExist = NewFormatError(http.StatusConflict, ReasonInternal, "resource conflict because %s exist")
	// ErrorLocking defines locking error
	ErrorLocking = NewFormatError(http.StatusConflict, ReasonLocking, "%s is locking and can't be handled: %v")
	// ErrorInvalidStatus defines invalid status error
	ErrorInvalidStatus = NewFormatError(http.StatusConflict, ReasonInternal, "%s status is invalid: %v")

	// ErrorInternalTypeError defines internal type error
	ErrorInternalTypeError = NewFormatError(http.StatusInternalServerError, ReasonInternal, "type of %s should be %s, but got %s")
	// ErrorUnknownNotFoundError defines not found error that we can't find a reason
	ErrorUnknownNotFoundError = NewFormatError(http.StatusInternalServerError, ReasonInternal, "content %s not found, may be it's a serious error")
	// ErrorInternalUnknown defines internal unknown error that we can't find a reason
	ErrorInternalUnknown = NewFormatError(http.StatusInternalServerError, ReasonInternal, "%v")
)
