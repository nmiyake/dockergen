// Copyright 2017 Nick Miyake. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root
// for license information.

package cmd

import (
	"sort"

	"github.com/nmiyake/dockergen/dockergen"
	"github.com/pkg/errors"
)

// getCommonParams returns the parameters for the action based on the specified image names. If the names are empty, all
// images are returned. If the names are non-empty and any of the specified names are not valid images, an error is
// returned. Otherwise, the returned build parameters are all of the builds required to build the requested images
// (including dependencies) sorted in topological order.
func getCommonParams(imageNames []string) (map[string]dockergen.Executor, []dockergen.BuildParams, dockergen.Params, error) {
	var all []string
	allBuildParams, err := cfg.BuildParams()
	if err != nil {
		return nil, nil, dockergen.Params{}, errors.WithStack(err)
	}

	executor := dockergen.NewCmdExecutor()
	if dryRun {
		executor = dockergen.NewPrintCmdExecutor()
	}

	allParamsMap := make(map[string]dockergen.BuildParams)
	allExecutorsMap := make(map[string]dockergen.Executor)
	for _, param := range allBuildParams {
		allParamsMap[param.Name] = param
		all = append(all, param.Name)
		allExecutorsMap[param.Name] = executor
	}

	imagesToBuild := allBuildParams
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
		imagesToBuild = requestedBuildParams

		dependentExecutor := executor
		if noDeps {
			// if dependencies should not be built, use a no-op executor for them
			dependentExecutor = dockergen.NoopExecutor()
		}

		// expand imagesToBuild to include all required dependent builds. If "noDeps" is true, this is still run, but
		// will use a no-op executor and thus will not perform the actual operations. However, the dependent builds must
		// still be executed to properly populate the tag map.
		seen := make(map[string]struct{})
		for _, image := range imagesToBuild {
			seen[image.Name] = struct{}{}
		}
		for _, param := range requestedBuildParams {
			for _, currReq := range dockergen.RequiredBuilds(param, allBuildParams) {
				if _, ok := seen[currReq.Name]; ok {
					continue
				}
				imagesToBuild = append(imagesToBuild, currReq)
				seen[currReq.Name] = struct{}{}
				allExecutorsMap[currReq.Name] = dependentExecutor
			}
		}
	}
	return allExecutorsMap, dockergen.TopologicalSort(imagesToBuild), cfg.ToParams(), nil
}
