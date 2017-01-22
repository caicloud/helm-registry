/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package utils

import (
	"reflect"

	"github.com/caicloud/helm-registry/pkg/log"
)

// Multicase test with multiple test cases
func Multicase(cases interface{}, function interface{}) func() {
	arr := reflect.ValueOf(cases)
	if arr.Kind() != reflect.Array && arr.Kind() != reflect.Slice {
		panic("cases must be an array")
	}
	if arr.Len() <= 0 {
		panic("there is at least one value in cases")
	}
	f := reflect.ValueOf(function)
	if f.Kind() != reflect.Func {
		panic("body must be a function")
	}
	if f.Type().IsVariadic() {
		panic("variadic function is not supported")
	}
	paramsCount := f.Type().NumIn()
	switch paramsCount {
	case 0:
		return func() {
			f.Call([]reflect.Value{})
		}
	case 1:
		return func() {
			for i := 0; i < arr.Len(); i++ {
				value := arr.Index(i)
				log.Infoln("test case:", i, "params:", value.Interface())
				f.Call([]reflect.Value{value})
			}
		}
	}
	return func() {
		values := make([]reflect.Value, paramsCount)
		for i := 0; i < arr.Len(); i++ {
			value := arr.Index(i)
			log.Infoln("test case:", i, "params:", value.Interface())
			if value.Kind() != reflect.Array && value.Kind() != reflect.Slice && value.Len() != paramsCount {
				panic("the case is not compatibal to the function")
			}
			for j := 0; j < paramsCount; j++ {
				values[j] = value.Index(j)
			}
			f.Call(values)
		}
	}
}
