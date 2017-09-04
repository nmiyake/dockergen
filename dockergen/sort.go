// Copyright 2017 Nick Miyake. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root
// for license information.

package dockergen

import (
	"sort"
)

func RequiredBuilds(build BuildParams, allParams []BuildParams) []BuildParams {
	allParamsMap := make(map[string]BuildParams)
	for _, k := range allParams {
		allParamsMap[k.Name] = k
	}

	allDeps := make(map[string]struct{})
	var remainingDeps []string
	remainingDeps = append(remainingDeps, build.Name)
	for len(remainingDeps) > 0 {
		currDep := remainingDeps[0]
		remainingDeps = remainingDeps[1:]
		allDeps[currDep] = struct{}{}
		currBuild := allParamsMap[currDep]
		for _, dep := range currBuild.Requires {
			remainingDeps = append(remainingDeps, dep)
		}
	}

	var output []BuildParams
	for k := range allDeps {
		output = append(output, allParamsMap[k])
	}
	sort.Slice(output, func(i, j int) bool {
		return output[i].Name < output[j].Name
	})
	return output
}

func TopologicalSort(builds []BuildParams) []BuildParams {
	reverseDeps := make(map[string][]string)
	for _, currBuild := range builds {
		for _, currReq := range currBuild.Requires {
			reverseDeps[currReq] = append(reverseDeps[currReq], currBuild.Name)
		}
	}
	for k, v := range reverseDeps {
		seen := make(map[string]struct{})
		var uniqVals []string
		for _, curr := range v {
			if _, ok := seen[curr]; ok {
				continue
			}
			uniqVals = append(uniqVals, curr)
			seen[curr] = struct{}{}
		}
		reverseDeps[k] = uniqVals
	}

	var sorted []BuildParams
	paramsMap := make(map[string]BuildParams)
	visited := make(map[string]bool)
	for _, curr := range builds {
		visited[curr.Name] = false
		paramsMap[curr.Name] = curr
	}
	for i := len(builds) - 1; i >= 0; i-- {
		sorted = topoVisit(builds[i], paramsMap, reverseDeps, visited, sorted)
	}
	// reverse
	for i, j := 0, len(sorted)-1; i < j; i, j = i+1, j-1 {
		sorted[i], sorted[j] = sorted[j], sorted[i]
	}
	return sorted
}

func topoVisit(curr BuildParams, params map[string]BuildParams, deps map[string][]string, visited map[string]bool, sorted []BuildParams) []BuildParams {
	if visited[curr.Name] {
		return sorted
	}
	visited[curr.Name] = true
	for _, currChild := range deps[curr.Name] {
		sorted = topoVisit(params[currChild], params, deps, visited, sorted)
	}
	return append(sorted, curr)
}
