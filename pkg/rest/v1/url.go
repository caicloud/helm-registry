/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package v1

import "strings"

// URL defines a url to generate api
type URL string

const (
	URLSpaces          URL = "/spaces"
	URLSpace           URL = "/spaces/{space}"
	URLCharts          URL = "/spaces/{space}/charts"
	URLChart           URL = "/spaces/{space}/charts/{chart}"
	URLChartMetadata   URL = "/spaces/{space}/charts/{chart}/metadata"
	URLVersions        URL = "/spaces/{space}/charts/{chart}/versions"
	URLVersion         URL = "/spaces/{space}/charts/{chart}/versions/{version}"
	URLVersionMetadata URL = "/spaces/{space}/charts/{chart}/versions/{version}/manifests/metadata"
	URLVersionValues   URL = "/spaces/{space}/charts/{chart}/versions/{version}/manifests/values"
)

// Format generates url. values should contain all keys in url.
// If any key does not exist, it generates wrong url.
func (u URL) Format(values map[string]string) string {
	oldnew := make([]string, 0, len(values)*2)
	for k, v := range values {
		oldnew = append(oldnew, "{"+k+"}", v)
	}
	replacer := strings.NewReplacer(oldnew...)
	return replacer.Replace(string(u))
}
