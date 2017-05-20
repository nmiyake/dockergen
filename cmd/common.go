// Copyright 2017 Nick Miyake. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root
// for license information.

package cmd

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/nmiyake/dockergen/dockergen"
)

func getCommonParams(cmd *cobra.Command, args []string) ([]dockergen.BuildParams, dockergen.Params, io.Writer) {
	buildParams := cfg.BuildParams()
	allParamsMap := make(map[string]dockergen.BuildParams)
	for _, param := range buildParams {
		allParamsMap[param.Name] = param
	}

	// if args were specified, run only the requested builds
	if len(args) != 0 {
		var requestedBuildParams []dockergen.BuildParams
		for _, curr := range args {
			param, ok := allParamsMap[curr]
			if !ok {
				continue
			}
			requestedBuildParams = append(requestedBuildParams, param)
		}
		buildParams = requestedBuildParams
	}

	return buildParams, cfg.ToParams(), cmd.OutOrStdout()
}
