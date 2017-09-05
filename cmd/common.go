// Copyright 2017 Nick Miyake. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root
// for license information.

package cmd

import (
	"sort"

	"github.com/pkg/errors"

	"github.com/nmiyake/dockergen/dockergen"
)

// getCommonParams returns the parameters for the action based on the specified image names. If the names are empty, all
// images are returned. If the names are non-empty and any of the specified names are not valid images, an error is
// returned. Otherwise, the returned build parameters are all of the builds required to build the requested images
// (including dependencies) sorted in topological order.
func getCommonParams(imageNames []string) (dockergen.Executor, []dockergen.BuildParams, dockergen.Params, error) {
	var all []string
	buildParams, err := cfg.BuildParams()
	if err != nil {
		return nil, nil, dockergen.Params{}, errors.WithStack(err)
	}

	allParamsMap := make(map[string]dockergen.BuildParams)
	for _, param := range buildParams {
		allParamsMap[param.Name] = param
		all = append(all, param.Name)
	}

	// if args were specified, run only the requested builds
	if len(imageNames) != 0 {
		var missing []string
		var requestedBuildParams []dockergen.BuildParams
		for _, curr := range imageNames {
			param, ok := allParamsMap[curr]
			if !ok {
				missing = append(missing, curr)
				continue
			}
			requestedBuildParams = append(requestedBuildParams, param)
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			sort.Strings(all)
			return nil, nil, dockergen.Params{}, errors.Errorf("The following specified entries were not defined in configuration: %v\nValid entries: %v", missing, all)
		}

		// expand parameters to include all required builds
		var expandedParams []dockergen.BuildParams
		currParamsMap := make(map[string]struct{})
		for _, param := range requestedBuildParams {
			for _, currReq := range dockergen.RequiredBuilds(param, buildParams) {
				if _, ok := currParamsMap[currReq.Name]; ok {
					continue
				}
				expandedParams = append(expandedParams, currReq)
				currParamsMap[currReq.Name] = struct{}{}
			}
		}
		buildParams = expandedParams
	}

	executor := dockergen.NewCmdExecutor()
	if dryRun {
		executor = dockergen.NewPrintCmdExecutor()
	}
	return executor, dockergen.TopologicalSort(buildParams), cfg.ToParams(), nil
}
