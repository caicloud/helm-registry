/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package space_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSpace(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Space Suite")
}
