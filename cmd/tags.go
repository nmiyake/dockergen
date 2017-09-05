// Copyright 2017 Nick Miyake. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root
// for license information.

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nmiyake/dockergen/dockergen"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Prints the tags for the Docker files specified in the configuration",
	Long: `Prints the tags for the images. If no arguments are provided, the tags for
all of the images in the configuration are printed. If arguments are provided, they
specify the names of the images for which tags are printed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		executor, builds, params, err := getCommonParams(args)
		if err != nil {
			return err
		}
		return dockergen.Tags(executor, builds, params, cmd.OutOrStdout())
	},
}

func init() {
	RootCmd.AddCommand(tagsCmd)
}
