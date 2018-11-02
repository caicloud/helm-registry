/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package types

import (
	"fmt"
	"net/http"

	"github.com/caicloud/helm-registry/pkg/errors"
)

// Save describes the info of a chart
type Save struct {
	// Space name. Be ignored by json.
	Space string `json:"-"`
	// Chart name
	Chart string `json:"chart"`
	// Version number
	Version string `json:"version"`
	// Desc is the description of the chart
	Desc string `json:"description"`
}

// Validate validates whether the info is valid
func (s *Save) Validate() error {
	if len(s.Chart) <= 0 {
		return errors.NewResponError(http.StatusBadRequest, "param.unfound", "${name} unfound", errors.M{
			"name": "save.chart",
		})
	}
	if len(s.Version) <= 0 {
		return errors.NewResponError(http.StatusBadRequest, "param.unfound", "${name} unfound", errors.M{
			"name": "save.version",
		})
	}
	return nil
}

// Path returns the path of current chart
func (s *Save) Path() string {
	return fmt.Sprintf("%s/%s/%s", s.Space, s.Chart, s.Version)
}

// OrchestrationConfig describes an orchestration config
type OrchestrationConfig struct {
	// Save contains the info of current chart
	Save Save `json:"save"`
	// Configs describes a config to orchestrate charts
	Configs map[string]interface{} `json:"configs"`
}

// Validate validates whether the config is valid
func (ac *OrchestrationConfig) Validate() error {
	if err := ac.Save.Validate(); err != nil {
		return err
	}
	if ac.Configs == nil {
		return errors.NewResponError(http.StatusBadRequest, "param.unfound", "${name} unfound", errors.M{
			"name": "configs",
		})
	}
	return nil
}
