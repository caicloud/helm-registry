/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package cmd

import (
	"time"

	"github.com/caicloud/helm-registry/pkg/api"
	"github.com/caicloud/helm-registry/pkg/common"
	"github.com/caicloud/helm-registry/pkg/log"
	"github.com/emicklei/go-restful"
	"github.com/spf13/cobra"
	"gopkg.in/tylerb/graceful.v1"
)

// config path
var configPath = ""

// serveCmd starts a http server for managing charts
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts a http server for managing a charts repository",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		// read config
		config, err := newConfig(configPath)
		if err != nil {
			log.Fatal(err)
		}

		// init SpaceManager
		common.Set(common.ContextNameSpaceManager, config.Manager.Name)
		common.Set(common.ContextNameSpaceParameters, config.Manager.Parameters)
		common.MustGetSpaceManager()

		// start server
		api.Initialize()
		log.Infof("Listening address %s", config.Listen)
		graceful.Run(config.Listen, 5*time.Minute, restful.DefaultContainer)
		log.Error("Server stopped")
	},
}

func init() {
	// bind variable configPath with flag --config or -c
	serveCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "path of config.yaml")
	rootCmd.AddCommand(serveCmd)
}
