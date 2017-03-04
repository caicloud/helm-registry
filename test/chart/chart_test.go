/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package chart_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/caicloud/helm-registry/pkg/log"
	"github.com/caicloud/helm-registry/pkg/rest/v1"
	"github.com/caicloud/helm-registry/test/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/helm/pkg/chartutil"
)

// EnvEndpoint is the env variable of registry host
const EnvEndpoint = "ENV_ENGPOINT"

type ChartPath struct {
	Space   string
	Chart   string
	Version string
	Path    string
}

var _ = Describe("Chart", func() {
	var (
		endpoint       = ""
		client         *v1.Client
		space          = "library"
		chart          = "test"
		version        = "1.0.0"
		combination    = [][]string{{space, chart, version}}
		validChartPath = []ChartPath{
			{space, chart, version, "./testdata/test1.tgz"},
		}
		invalidChartPath = []ChartPath{
			{space, chart, version, "./testdata/test1.tgz"},
			{space, "xxxx", "1.0.1", "./testdata/test1.tgz"},
		}
		validUpdatedChartPath = []ChartPath{
			{space, chart, "1.0.0", "./testdata/test2.tgz"},
		}
		invalidUpdatedChartPath = []ChartPath{
			{space, "xxxx", "1.0.1", "./testdata/test2.tgz"},
		}
	)
	BeforeEach(func() {
		By("getting registry host from env")
		endpoint = os.Getenv(EnvEndpoint)
		Expect(endpoint).NotTo(BeEmpty())
		cli, err := v1.NewClient(endpoint)
		Expect(err).To(BeNil())
		client = cli
	})
	Context("upload chart", func() {
		It("should upload chart", utils.Multicase(validChartPath, func(path ChartPath) {
			data, err := ioutil.ReadFile(path.Path)
			Expect(err).To(BeNil())
			link, err := client.UploadVersion(path.Space, path.Chart, path.Version, data)
			Expect(err).To(BeNil())
			Expect([]string{link.Space, link.Chart, link.Version}).
				To(Equal([]string{path.Space, path.Chart, path.Version}))
			log.Infoln("upload chart", link)
		}))
		It("shouldn't upload chart", utils.Multicase(invalidChartPath, func(path ChartPath) {
			data, err := ioutil.ReadFile(path.Path)
			Expect(err).To(BeNil())
			_, err = client.UploadVersion(path.Space, path.Chart, path.Version, data)
			Expect(err).NotTo(BeNil())
		}))
	})

	Context("list chart and version", func() {
		It("should list chart", utils.Multicase(combination, func(space, chart, version string) {
			result, err := client.ListCharts(space, 0, 100000)
			Expect(err).To(BeNil())
			Expect(result.Metadata.Total).To(Equal(1))
			Expect(result.Metadata.ItemsLength).To(Equal(1))
			Expect(result.Items[0]).To(Equal(chart))
			log.Infoln("list chart", result)
		}))
		It("should list chart version", utils.Multicase(combination, func(space, chart, version string) {
			result, err := client.ListVersions(space, chart, 0, 100000)
			Expect(err).To(BeNil())
			Expect(result.Metadata.Total).To(Equal(1))
			Expect(result.Metadata.ItemsLength).To(Equal(1))
			Expect(result.Items[0]).To(Equal(version))
			log.Infoln("list chart version", result)
		}))
	})

	Context("fetch metadata and values", func() {
		It("should get chart metadata", utils.Multicase(combination, func(space, chart, version string) {
			result, err := client.FetchChartMetadata(space, chart, 0, 100000)
			Expect(err).To(BeNil())
			Expect(result.Metadata.Total).To(Equal(1))
			Expect(result.Metadata.ItemsLength).To(Equal(1))
			md := result.Items[0]
			Expect(md.Name).To(Equal(chart))
			Expect(md.Version).To(Equal(version))
			log.Infoln("get chart metadata", result)
		}))
		It("should get version metadata", utils.Multicase(combination, func(space, chart, version string) {
			md, err := client.FetchVersionMetadata(space, chart, version)
			Expect(err).To(BeNil())
			Expect(md.Name).To(Equal(chart))
			Expect(md.Version).To(Equal(version))
			log.Infoln("get version metadata", md)
		}))
		It("should get version values", utils.Multicase(combination, func(space, chart, version string) {
			values, err := client.FetchVersionValues(space, chart, version)
			Expect(err).To(BeNil())
			obj := map[string]interface{}{}
			err = json.Unmarshal(values, &obj)
			Expect(err).To(BeNil())
			Expect(obj["replicaCount"]).To(Equal(3.0))
			cpu, ok := obj["cpu"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(cpu["requests"]).To(Equal("1000m"))
			Expect(cpu["limits"]).To(Equal("1000m"))
			memory, ok := obj["memory"].(map[string]interface{})
			Expect(ok).To(BeTrue())
			Expect(memory["requests"]).To(Equal("128Mi"))
			Expect(memory["limits"]).To(Equal("128Mi"))
			log.Infoln("get version values", string(values))
		}))
	})

	Context("update chart", func() {
		It("should update chart", utils.Multicase(validUpdatedChartPath, func(path ChartPath) {
			data, err := ioutil.ReadFile(path.Path)
			Expect(err).To(BeNil())
			link, err := client.UpdateVersion(path.Space, path.Chart, path.Version, data)
			Expect(err).To(BeNil())
			Expect([]string{link.Space, link.Chart, link.Version}).
				To(Equal([]string{path.Space, path.Chart, path.Version}))
			log.Infoln("update chart", link)
		}))
		It("shouldn't update chart", utils.Multicase(invalidUpdatedChartPath, func(path ChartPath) {
			data, err := ioutil.ReadFile(path.Path)
			Expect(err).To(BeNil())
			_, err = client.UpdateVersion(path.Space, path.Chart, path.Version, data)
			Expect(err).NotTo(BeNil())
		}))
	})

	Context("download chart", func() {
		It("should download chart", utils.Multicase(combination, func(space, chart, version string) {
			data, err := client.DownloadVersion(space, chart, version)
			Expect(err).To(BeNil())
			downloaded, err := chartutil.LoadArchive(bytes.NewReader(data))
			Expect(err).To(BeNil())
			Expect(downloaded.Metadata.Name).To(Equal(chart))
			Expect(downloaded.Metadata.Version).To(Equal(version))
			Expect(downloaded.Values.Raw).To(Equal(`replicaCount: 3
cpu:
  requests: 2000m
  limits: 2000m
memory:
  requests: 256Mi
  limits: 256Mi`))
			log.Infoln("download chart", downloaded.Metadata)
		}))
	})

	Context("create chart", func() {
		It("should create chart", utils.Multicase(combination, func(space, chart, version string) {
			config := `
{
    "save":{         
        "space":"library",                  
        "chart":"testX",           
        "version":"1.0.1",             
        "description":"a test for creating chart"           
    },
    "configs":{                        
        "package":{                    
            "independent":true,        
            "space":"library",      
            "chart":"test",      
            "version":"1.0.0" 
        },
        "_config": {
            "storage1": "ssd"            
        },
        "chartB": {
            "package":{
                "independent":true,        
                "space":"library",
                "chart":"test",
                "version":"1.0.0"
            },
            "_config": {
                "storage2": "ssd"
            },
            "chartD":{
                "package":{
                    "independent":true,
                    "space":"library",
                    "chart":"test",
                    "version":"1.0.0"
                },
                "_config": {
                    "storage3": "ssd"
                }
            }
        },
        "chartC": {
            "package":{
                "independent":true,
                "space":"library",
                "chart":"test",
                "version":"1.0.0"
            },
            "_config": {
                "storage4": "ssd"
            }
        }
    }
}
`
			const (
				chartName     = "testX"
				versionNumber = "1.0.1"
			)
			link, err := client.CreateChart(space, config)
			Expect(err).To(BeNil())
			Expect([]string{link.Space, link.Chart, link.Version}).
				To(Equal([]string{space, chartName, versionNumber}))
			// check metadata
			md, err := client.FetchVersionMetadata(space, chartName, versionNumber)
			Expect(err).To(BeNil())
			Expect(md.Name).To(Equal(chartName))
			Expect(md.Version).To(Equal(versionNumber))
			log.Infoln("get creation metadata", md)

			// check values
			values, err := client.FetchVersionValues(space, chartName, versionNumber)
			Expect(err).To(BeNil())
			log.Infoln("get creation values", string(values))

			// delete creation
			err = client.DeleteChart(space, chartName)
			Expect(err).To(BeNil())
		}))
	})

	Context("delete chart and version", func() {
		It("should delete chart version", utils.Multicase(combination, func(space, chart, version string) {
			err := client.DeleteVersion(space, chart, version)
			Expect(err).To(BeNil())
			result, err := client.ListCharts(space, 0, 100000)
			Expect(err).To(BeNil())
			Expect(result.Metadata.Total).To(Equal(0))
			Expect(result.Metadata.ItemsLength).To(Equal(0))
			log.Infoln("delete chart version", result)
		}))
		It("should delete chart", utils.Multicase(validChartPath, func(path ChartPath) {
			// upload
			data, err := ioutil.ReadFile(path.Path)
			Expect(err).To(BeNil())
			link, err := client.UploadVersion(path.Space, path.Chart, path.Version, data)
			Expect(err).To(BeNil())
			Expect([]string{link.Space, link.Chart, link.Version}).
				To(Equal([]string{path.Space, path.Chart, path.Version}))

			// delete
			err = client.DeleteChart(path.Space, path.Chart)
			Expect(err).To(BeNil())
			result, err := client.ListCharts(path.Space, 0, 100000)
			Expect(err).To(BeNil())
			Expect(result.Metadata.Total).To(Equal(0))
			Expect(result.Metadata.ItemsLength).To(Equal(0))
			log.Infoln("delete chart", result)
		}))

	})

})
