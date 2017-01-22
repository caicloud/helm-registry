/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package models

import "reflect"

// Metadata describes the structure of list metadata
type Metadata struct {
	Total       int `json:"total"`
	ItemsLength int `json:"itemsLength"`
}

// ListResponse describes a list
type ListResponse struct {
	Metadata Metadata    `json:"metadata"`
	Items    interface{} `json:"items"`
}

// NewListResponse create a ListResponse of a list.
func NewListResponse(total int, items interface{}) *ListResponse {
	value := reflect.ValueOf(items)
	typ := value.Type()
	kind := typ.Kind()
	if kind != reflect.Array && kind != reflect.Slice {
		items = []interface{}{items}
	}
	return &ListResponse{
		Metadata{
			Total:       total,
			ItemsLength: value.Len(),
		},
		items,
	}
}
