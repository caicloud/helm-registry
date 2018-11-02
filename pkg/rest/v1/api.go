/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"strconv"

	"github.com/caicloud/helm-registry/pkg/errors"
	"github.com/caicloud/helm-registry/pkg/log"
	"github.com/caicloud/helm-registry/pkg/rest"
)

// APIVersion defines api version
const APIVersion = "v1"

// File describes file infomation
type File struct {
	// Path is the path of file
	Path string
	// Data is the data of file. If the field is nil, Path should be a
	// valid path to local file.
	Data []byte
}

// file defines request field of file
type file struct {
	// field is the name of filed
	field string
	// path is the path of file
	path string
	// data is the data of the file
	data []byte
}

// writeTo writes file data to writer
func (f *file) writeTo(writer io.Writer) {
	if f.data == nil {
		data, err := ioutil.ReadFile(f.path)
		if err == nil {
			writer.Write(data)
		} else {
			log.Errorf("can't read file: %s", f.path)
		}
	} else {
		writer.Write(f.data)
	}
}

// baseAPI defines a base api to request registry
type baseAPI struct {
	// object is real api object
	object rest.API
	// method is http method
	method string
	// url is a path of api
	url URL
	// paths is a map of url parameters
	paths map[string]string
	// values is the query parameters of request
	values url.Values
	// files is an array to store file parameters
	files []*file
	// body is data of request body. If method is GET, ignore this field.
	// If this field is not nil and method is not GET, ignore files and append values
	// to url whatever method is.
	body []byte
	// result is a pointer and will be filled by json from body. If result is non-pointer,
	// response will be []byte and ignore result. If result is nil, response do nothing.
	result interface{}
}

// Method returns the http method of current api
func (ba *baseAPI) Method() string {
	return ba.method
}

// Path returns the url path of current api
func (ba *baseAPI) Path() string {
	return path.Join("/api/", APIVersion, string(ba.url))
}

// addPath adds key-value to request
func (ba *baseAPI) addPath(key, value string) {
	if ba.paths == nil {
		ba.paths = make(map[string]string)
	}
	ba.paths[key] = value
}

// addValue adds key-value to request
func (ba *baseAPI) addValue(key, value string) {
	if ba.values == nil {
		ba.values = url.Values{}
	}
	ba.values.Add(key, value)
}

// addFile adds file to request
func (ba *baseAPI) addFile(key, path string, data []byte) {
	ba.files = append(ba.files, &file{
		field: key,
		path:  path,
		data:  data,
	})
}

// addLocalFile adds local file to request
func (ba *baseAPI) addLocalFile(key, path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return rest.ErrorUnknownLocalError.Format(err.Error())
	}
	if info.IsDir() {
		return errors.NewResponError(http.StatusBadRequest, "param.error", "${name} error", errors.M{
			"name": path,
		})
	}
	ba.addFile(key, path, nil)
	return nil
}

var fileType = reflect.TypeOf(new(File))
var bytesType = reflect.TypeOf(make([]byte, 0))
var stringType = reflect.TypeOf("")

// generateParameters generates paths, values, files from object.
// A API object have some fileds, if any field meet the requirements, it can be
// save as parameter:
//  1. Not an anonymous field
//  2. Have tag 'kind' and value is one of: path, query, file, body
//  3. Have tag 'name' and field type is one of: string, int, *v1.File, []byte
//  4. When 'kind' is path or query, field type should be string or int
//  4. When 'kind' is file, field type should be *v1.File
//  5. When 'kind' is body, field type should be string or []byte
//  6. There is at most one body field in an API, If more than one, the last is valid
//
// e.g. SampleAPI: /api/v1/resources/{resource1}
//  type SampleAPI struct {
//		baseAPI
//
//		// fieldA won't be export
//		fieldA int
//
//		// FieldB replaces resource1 in api path
//		FieldB string `kind:"path" name:"resource1"`
//
//		// FieldC will be appended as query parameter
//		FieldC int `kind:"query" name:"field"`
//
//		// FieldD will be set to request body
//		FieldD []byte `kind:"body"`
//
//		// FieldE is a file. and the field will change request content type from
//		// "application/x-www-form-urlencoded" to "multipart/form-data".
//		// If there is any body field, all file fields will be ignored.
//		FieldE *v1.File `kind:"file" name:"file"`
//  }
func (ba *baseAPI) generateParameters() {
	if ba.object == nil {
		return
	}
	apiValue := reflect.ValueOf(ba.object)
	apiType := apiValue.Type()
	// object should be an api
	if apiType.Kind() != reflect.Ptr || apiType.Elem().Kind() != reflect.Struct {
		log.Fatalf("unknown api type: %s", apiType.String())
	}
	elem := apiType.Elem()
	elemValue := apiValue.Elem()
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		if field.Anonymous {
			continue
		}

		// check tag kind
		kind := field.Tag.Get("kind")
		if len(kind) <= 0 {
			continue
		}
		fieldValue := elemValue.Field(i).Interface()
		if kind == "body" {
			// handler body field
			switch field.Type {
			case bytesType:
				ba.body = fieldValue.([]byte)
			case stringType:
				ba.body = []byte(fieldValue.(string))
			default:
				log.Fatalf("field %s.%s should be %s, but got %s", elem.String(), field.Name,
					fileType.String(), field.Type.String())
			}
			continue
		}

		// check tag name
		name := field.Tag.Get("name")
		if len(name) <= 0 {
			continue
		}
		if kind == "file" {
			// handle file field
			if field.Type != fileType {
				log.Fatalf("field %s.%s should be %s, but got %s", elem.String(), field.Name,
					fileType.String(), field.Type.String())
			}
			file := fieldValue.(*File)
			ba.addFile(name, file.Path, file.Data)
			continue
		}
		// handle path and query field
		value := ""
		switch field.Type.Kind() {
		case reflect.String:
			value = fieldValue.(string)
		case reflect.Int:
			intValue := fieldValue.(int)
			value = strconv.Itoa(intValue)
		default:
			log.Fatalf("unknown api kind type shoule be string or int, but got %s", field.Type.Kind())
		}
		switch kind {
		case "path":
			// handle path field
			ba.addPath(name, value)
		case "query":
			// handle query field
			ba.addValue(name, value)
		default:
			log.Fatalf("unknown api kind type: %s", kind)
		}
	}
}

// request generates a request for current api. A http endpoint should
// be http://host:port or https://host:port.
func (ba *baseAPI) request(endpoint string) (*http.Request, error) {
	// generate api path by ba.paths
	path := URL(ba.Path()).Format(ba.paths)
	contentType := ""
	var body io.Reader
	if ba.Method() == http.MethodGet || body != nil {
		// append values to url
		if len(ba.values) > 0 {
			path += "?" + ba.values.Encode()
		}
	}
	if ba.body != nil {
		// use ba.body as request body
		// ignore ba.files
		body = bytes.NewBuffer(ba.body)
	} else {
		if len(ba.files) <= 0 {
			// application/x-www-form-urlencoded
			encodedValues := ba.values.Encode()
			if len(encodedValues) > 0 {
				contentType = "application/x-www-form-urlencoded"
				body = bytes.NewBufferString(encodedValues)
			}
		} else {
			// multipart/form-data
			buf := bytes.NewBuffer(nil)
			writer := multipart.NewWriter(buf)
			// write values
			for key, array := range ba.values {
				for _, v := range array {
					w, err := writer.CreateFormField(key)
					if err != nil {
						return nil, rest.ErrorUnknownLocalError.Format(err.Error())
					}
					fmt.Fprint(w, v)
				}
			}
			// write files
			for _, file := range ba.files {
				if len(file.path) <= 0 {
					// file path must be a valid string
					file.path = file.field
				}
				w, err := writer.CreateFormFile(file.field, file.path)
				if err != nil {
					return nil, rest.ErrorUnknownLocalError.Format(err.Error())
				}
				file.writeTo(w)
			}
			// close
			writer.Close()
			contentType = "multipart/form-data; boundary=" + writer.Boundary()
			body = buf
		}
	}
	// generate http request
	req, err := http.NewRequest(ba.Method(), endpoint+path, body)
	if len(contentType) > 0 {
		req.Header.Set("Content-Type", contentType)
	}
	if err != nil {
		return nil, rest.ErrorUnknownLocalError.Format(err.Error())
	}
	return req, nil
}

// Request generates a request from object for current api. A http endpoint should
// be http://host:port or https://host:port.
func (ba *baseAPI) Request(endpoint string) (*http.Request, error) {
	ba.generateParameters()
	return ba.request(endpoint)
}

// responseData handles *http.Response and return response data
func (ba *baseAPI) responseData(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, rest.ErrorUnknownLocalError.Format(err.Error())
	}
	return data, nil
}

// unmarshal get data from resp and unmarshal
func (ba *baseAPI) unmarshal(resp *http.Response, obj interface{}) error {
	data, err := ba.responseData(resp)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, obj)
	if err != nil {
		return rest.ErrorUnknownLocalError.Format(err.Error())
	}
	return nil
}

// Response handles *http.Response and return result
func (ba *baseAPI) Response(resp *http.Response) (interface{}, error) {
	if ba.result != nil {
		value := reflect.ValueOf(ba.result)
		if value.Kind() != reflect.Ptr {
			return ba.responseData(resp)
		}
		return ba.result, ba.unmarshal(resp, ba.result)
	}
	return nil, nil
}
