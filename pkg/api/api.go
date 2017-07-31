/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package api

import (
	"time"

	"github.com/caicloud/helm-registry/pkg/api/v1"
	"github.com/caicloud/helm-registry/pkg/log"
	"github.com/emicklei/go-restful"
)

// Initialize initializes apis of all versions
func Initialize() {
	v1.InstallRouters(restful.DefaultContainer)
	restful.EnableTracing(true)
	restful.DefaultContainer.Filter(NCSACommonLogFormatLogger())
}

// NCSACommonLogFormatLogger adds logs for every request using common log format.
func NCSACommonLogFormatLogger() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		r := req.Request
		start := time.Now()
		log.Printf("Started %s - [%s] %s %s",
			req.Request.RemoteAddr,
			time.Now().Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			req.Request.URL.RequestURI(),
		)
		chain.ProcessFilter(req, resp)
		log.Printf("%s - [%s] \"%s %s %s\" %d %d %v",
			req.Request.RemoteAddr,
			time.Now().Format("02/Jan/2006:15:04:05 -0700"),
			req.Request.Method,
			req.Request.URL.RequestURI(),
			req.Request.Proto,
			resp.StatusCode(),
			resp.ContentLength(),
			time.Since(start),
		)
	}
}
