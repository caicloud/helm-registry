/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package space_test

import (
	"os"
	"sort"

	"github.com/caicloud/helm-registry/pkg/rest"
	"github.com/caicloud/helm-registry/pkg/rest/v1"
	"github.com/caicloud/helm-registry/test/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// EnvEndpoint is the env variable of registry host
const EnvEndpoint = "ENV_ENGPOINT"

var _ = Describe("Space", func() {
	var (
		endpoint    = ""
		client      *v1.Client
		validSpaces = []string{
			"test",
			"teXa-sd--_t",
			"test_Space",
		}
		invalidSpaces = []string{
			"_srsd",
			"_11srsd",
			"-11srsd",
		}
	)
	sort.Strings(validSpaces)

	BeforeEach(func() {
		By("getting registry host from env")
		endpoint = os.Getenv(EnvEndpoint)
		Expect(endpoint).NotTo(BeEmpty())
		cli, err := v1.NewClient(endpoint)
		Expect(err).To(BeNil())
		client = cli
	})
	Context("create spaces", func() {
		It("should create spaces", utils.Multicase(validSpaces, func(space string) {
			_, err := client.CreateSpace(space)
			Expect(err).To(BeNil())
		}))
		It("shouldn't create any space", utils.Multicase(invalidSpaces, func(space string) {
			_, err := client.CreateSpace(space)
			Expect(rest.ErrorBadRequest.Equal(err)).To(BeTrue())
		}))
	})

	Context("list spaces", func() {
		It("should get spaces", func() {
			spaces, err := client.ListSpaces(0, 100000)
			Expect(err).To(BeNil())
			sort.Strings(spaces.Items)
			Expect(validSpaces).To(Equal(spaces.Items))
		})
	})

	Context("delete spaces", func() {
		It("should delete spaces", utils.Multicase(validSpaces, func(space string) {
			err := client.DeleteSpace(space)
			Expect(err).To(BeNil())
		}))
	})
})
