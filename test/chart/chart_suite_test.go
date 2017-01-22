/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package chart_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestChart(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chart Suite")
}
