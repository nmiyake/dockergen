// Copyright 2017 Nick Miyake. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root
// for license information.

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nmiyake/dockergen/dockergen"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Pushes the tags for the Dockerfiles specified in the configuration",
	Long: `Pushes tags for images. If no arguments are provided, all of the tags for the
images in the configuration are pushed. If arguments are provided, they specify the names of
the images whose tags should be pushed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		builds, params, stdout, err := getCommonParams(cmd, args)
		if err != nil {
			return err
		}
		return dockergen.Push(builds, params, stdout)
	},
}

func init() {
	RootCmd.AddCommand(pushCmd)
}
