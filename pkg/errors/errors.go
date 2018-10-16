/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package errors

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

//APIReseaonPrefix prefix
const APIReseaonPrefix = "helm:"

//M error map
type M map[string]interface{}

// Error defines error with code
type Error struct {
	ID      int    `json:"id"`
	Code    int    `json:"code"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
	Data    M      `json:"data"`
	format  string
}

// Error returns error reason
func (e *Error) Error() string {
	if e.Data != nil {
		for k, v := range e.Data {
			e.Message = strings.Replace(e.Message, "%("+k+")", fmt.Sprintf("%+v", v), 1)
		}
	}
	return e.Message
}

// Equal returns whether err.ID equal to e.ID
func (e *Error) Equal(err error) bool {
	if errx, ok := (err).(*Error); ok {
		return errx.ID == e.ID
	}
	return false
}

// Format generate an specified error
func (e *Error) Format(params ...interface{}) *Error {
	if len(e.format) <= 0 {
		return e
	}
	return &Error{
		e.ID,
		e.Code,
		e.Reason,
		fmt.Sprintf(e.format, params...),
		e.Detail,
		nil,
		e.format,
	}
}

// id counter
var counter int32

// NewErrorID generates an unique id in current runtime
func NewErrorID() int {
	return int(atomic.AddInt32(&counter, 1))
}

// NewStaticError creates a static error
func NewStaticError(code int, reason string, message string) *Error {
	return &Error{NewErrorID(), code, reason, message, "", nil, ""}
}

// NewFormatError creates a format error
func NewFormatError(code int, reason string, format string) *Error {
	return &Error{NewErrorID(), code, reason, "", "", nil, format}
}

//NewConflict create a conflict error zh-cn
func NewConflict(reason, message string, data M) *Error {
	return &Error{NewErrorID(), http.StatusConflict, reason, message, "", data, ""}
}
