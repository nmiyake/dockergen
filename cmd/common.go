// Copyright 2017 Nick Miyake. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root
// for license information.

package cmd

import (
	"io"
	"sort"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/nmiyake/dockergen/dockergen"
)

func getCommonParams(cmd *cobra.Command, args []string) ([]dockergen.BuildParams, dockergen.Params, io.Writer, error) {
	if len(args) > 0 {
		args = args[1:]
	}

	var all []string
	buildParams := cfg.BuildParams()
	allParamsMap := make(map[string]dockergen.BuildParams)
	for _, param := range buildParams {
		allParamsMap[param.Name] = param
		all = append(all, param.Name)
	}

	// if args were specified, run only the requested builds
	var missing []string
	if len(args) != 0 {
		var requestedBuildParams []dockergen.BuildParams
		for _, curr := range args {
			param, ok := allParamsMap[curr]
			if !ok {
				missing = append(missing, curr)
				continue
			}
			requestedBuildParams = append(requestedBuildParams, param)
		}
		buildParams = requestedBuildParams
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		sort.Strings(all)
		return nil, dockergen.Params{}, nil, errors.Errorf("The following specified entries were not defined in configuration: %v\nValid entries: %v", missing, all)
	}

	return buildParams, cfg.ToParams(), cmd.OutOrStdout(), nil
}
