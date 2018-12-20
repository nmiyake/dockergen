// Copyright 2017 Nick Miyake. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root
// for license information.

package cmd

import (
	"github.com/nmiyake/dockergen/dockergen"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds and tags the Docker files specified in the configuration",
	Long: `Builds and tags images. If no arguments are provided, all of the images
in the configuration are built. If arguments are provided, they specify the names of the
images that should be built.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		executor, builds, params, err := getCommonParams(args)
		if err != nil {
			return err
		}
		return dockergen.Build(executor, builds, params, cmd.OutOrStdout())
	},
}

func init() {
	RootCmd.AddCommand(buildCmd)
}
