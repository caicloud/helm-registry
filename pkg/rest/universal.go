/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/caicloud/helm-registry/pkg/errors"
	"github.com/caicloud/helm-registry/pkg/log"
)

// UniversalClient implements a basic registry client
type UniversalClient struct {
	client   *http.Client
	EndPoint string
}

// NewUniversalClient creates a UniversalClient
func NewUniversalClient(endpoint string) *UniversalClient {
	return &UniversalClient{&http.Client{}, strings.TrimRight(endpoint, "\\/")}
}

// Do requests specific api and gets response. If there is no error,
// It will returns specific result of corresponding api.
func (cc *UniversalClient) Do(api API) (interface{}, error) {
	if api == nil {
		log.Panic("registry client can't request a nil api")
	}
	req, err := api.Request(cc.EndPoint)
	if err != nil {
		return nil, err
	}
	resp, err := cc.client.Do(req)
	// check err
	if err != nil {
		return nil, ErrorNoResponse.Format(err.Error())
	}
	// not 2xx
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			// translates error
			var merr *errors.Error
			switch resp.StatusCode {
			case ErrorBadRequest.Code:
				merr = ErrorBadRequest.Format("")
			case ErrorNotFound.Code:
				merr = ErrorNotFound.Format("")
			case ErrorConflict.Code:
				merr = ErrorConflict.Format("")
			case ErrorLocked.Code:
				merr = ErrorLocked.Format("")
			case ErrorServer.Code:
				merr = ErrorServer.Format("")
			default:
				merr = &errors.Error{}
			}
			err = json.Unmarshal(data, merr)
			if err == nil {
				merr.Code = resp.StatusCode
				return nil, merr
			}
		}
		// unknown error
		return nil, errors.NewStaticError(resp.StatusCode, errors.ReasonLocal, err.Error())
	}
	return api.Response(resp)
}
