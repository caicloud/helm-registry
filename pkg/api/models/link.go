/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package models

// Link describes a normal self-link
type Link struct {
	// Name is object name
	Name string `json:"name"`
	// Link is the uri of object
	Link string `json:"link"`
}

// NewLink creates a self-link
func NewLink(name, link string) *Link {
	return &Link{name, link}
}

// ChartLink describes a chart self-link
type ChartLink struct {
	// Space is the space chart is stored
	Space string `json:"space"`
	// Chart is chart name
	Chart string `json:"chart"`
	// Version is chart version
	Version string `json:"version"`
	// Link is the uri of object
	Link string `json:"link"`
}

// NewChartLink creates a chart self-link
func NewChartLink(space, chart, version, link string) *ChartLink {
	return &ChartLink{space, chart, version, link}
}
