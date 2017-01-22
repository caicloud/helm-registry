/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package main

import (
	"github.com/caicloud/helm-registry/cmd/registry/cmd"
	_ "github.com/caicloud/helm-registry/pkg/storage/simple"
	_ "github.com/docker/distribution/registry/storage/driver/filesystem"
)

func main() {
	cmd.Run()
}
