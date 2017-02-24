/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package rest

import (
	"net/http"
)

// API defines an api of specific version
type API interface {
	// Method returns the http method of current api
	Method() string
	// Path returns the url path of current api
	Path() string
	// Request generates a request for current api. A http endpoint should
	// be http://host:port or https://host:port.
	Request(endpoint string) (*http.Request, error)
	// Response handles *http.Response and return result
	Response(resp *http.Response) (result interface{}, err error)
}

// Client defines a restful client to request registry server
type Client interface {
	// Do request specific api and pass param as parameter. If there is no error,
	// It will returns specific result
	Do(api API) (result interface{}, err error)
}
