// Copyright 2017 Nick Miyake. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root
// for license information.

package dockergen

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/pkg/errors"
)

const (
	defaultTagSuffix = "-{{BuildID}}"
	defaultBuildID   = "unspecified"
)

func Build(executors map[string]Executor, builds []BuildParams, dockerGenParams Params, stdout io.Writer) error {
	return runActionLogic(runBuildAction, executors, builds, dockerGenParams, stdout)
}

func Push(executors map[string]Executor, builds []BuildParams, dockerGenParams Params, stdout io.Writer) error {
	return runActionLogic(runPushAction, executors, builds, dockerGenParams, stdout)
}

func Tags(executors map[string]Executor, builds []BuildParams, dockerGenParams Params, stdout io.Writer) error {
	return runActionLogic(runTagAction, executors, builds, dockerGenParams, stdout)
}

func runActionLogic(action runActionFunc, executors map[string]Executor, builds []BuildParams, dockerGenParams Params, stdout io.Writer) error {
	if err := dockerGenParams.Validate(); err != nil {
		return errors.Wrapf(err, "invalid Docker generator params")
	}

	// evaluate the build variable
	buildID := defaultBuildID
	if dockerGenParams.BuildIDVar != "" {
		if envVar := os.Getenv(dockerGenParams.BuildIDVar); envVar != "" {
			buildID = envVar
		}
	}

	tagSuffixTmpl := defaultTagSuffix
	if dockerGenParams.TagSuffix != "" {
		tagSuffixTmpl = dockerGenParams.TagSuffix
	}

	evaluatedVarMap := make(map[string]string)
	for k, v := range dockerGenParams.TemplateVars {
		valResult, err := executeGoTemplate(v, buildID, nil, nil, -1, -1)
		if err != nil {
			return errors.Wrapf(err, "failed to execute template for variable %s", k)
		}
		evaluatedVarMap[k] = valResult
	}

	tags := make(map[string][][]string)
	return runInFor(func(idx int, curEvalVarMap map[string]string) error {
		for _, currBuild := range builds {
			innerTags, err := runAction(action, executors[currBuild.Name], currBuild, buildID, tagSuffixTmpl, curEvalVarMap, tags, idx, stdout)
			if err != nil {
				return errors.Wrapf(err, "failed to build %s", currBuild.Name)
			}
			tags[currBuild.Name] = append(tags[currBuild.Name], innerTags)
		}
		return nil
	}, dockerGenParams.For, buildID, evaluatedVarMap, tags)
}

func runInFor(f func(int, map[string]string) error, forVars map[string][]string, buildID string, evaluatedVarsIn map[string]string, inputTags map[string][][]string) error {
	// copy input map so that modifications made in for loop are not persisted
	evaluatedVars := make(map[string]string, len(evaluatedVarsIn))
	for k, v := range evaluatedVarsIn {
		evaluatedVars[k] = v
	}

	// if there are no "for" variables, run once
	if len(forVars) == 0 {
		return f(0, evaluatedVars)
	}

	// run build for each for loop variable
	var sortedForVarNames []string
	for k := range forVars {
		sortedForVarNames = append(sortedForVarNames, k)
	}
	sort.Strings(sortedForVarNames)

	for i := 0; i < len(forVars[sortedForVarNames[0]]); i++ {
		// set variable values for this iteration
		for _, currForVar := range sortedForVarNames {
			currForVarResult, err := executeGoTemplate(forVars[currForVar][i], buildID, evaluatedVars, inputTags, -1, -1)
			if err != nil {
				return errors.Wrapf(err, "failed to execute template for 'for' variable %s at index %d", currForVar, i)
			}
			evaluatedVars[currForVar] = currForVarResult
		}
		if err := f(i, evaluatedVars); err != nil {
			return err
		}
	}
	return nil
}

func runAction(action runActionFunc, executor Executor, build BuildParams, buildID, tagSuffixTmpl string, evaluatedVars map[string]string, inputTags map[string][][]string, outerIdx int, stdout io.Writer) ([]string, error) {
	var tags []string
	err := runInFor(func(innerIdx int, curEvalVarMap map[string]string) error {
		renderedTag, err := executeGoTemplate(build.Tag, buildID, curEvalVarMap, inputTags, outerIdx, innerIdx)
		if err != nil {
			return errors.Wrapf(err, "failed to execute template for tag")
		}
		renderedTagSuffix, err := executeGoTemplate(tagSuffixTmpl, buildID, curEvalVarMap, inputTags, outerIdx, innerIdx)
		if err != nil {
			return errors.Wrapf(err, "failed to execute template for tag suffix")
		}
		tag := renderedTag + renderedTagSuffix
		if tag == "" {
			return errors.Errorf("tag must be non-empty")
		}
		tags = append(tags, tag)
		return action(runParams{
			executor:   executor,
			build:      build,
			buildID:    buildID,
			tag:        tag,
			evalVarMap: curEvalVarMap,
			inputTags:  inputTags,
			outerIdx:   outerIdx,
			innerIdx:   innerIdx,
			stdout:     stdout,
		})
	}, build.For, buildID, evaluatedVars, inputTags)
	return tags, err
}

type runActionFunc func(params runParams) error

type runParams struct {
	executor   Executor
	build      BuildParams
	buildID    string
	tag        string
	evalVarMap map[string]string
	inputTags  map[string][][]string
	outerIdx   int
	innerIdx   int
	stdout     io.Writer
}

func runBuildAction(params runParams) error {
	bytes, err := ioutil.ReadFile(params.build.DockerfileTemplatePath)
	if err != nil {
		return errors.Wrapf(err, "failed to read Dockerfile template")
	}

	renderedDockerfile, err := executeGoTemplate(string(bytes), params.buildID, params.evalVarMap, params.inputTags, params.outerIdx, params.innerIdx)
	if err != nil {
		return errors.Wrapf(err, "failed to execute template for Dockerfile")
	}

	return executeDockerBuild(params.executor, renderedDockerfile, params.tag, params.build.DockerfileTemplatePath, params.stdout)
}

func runPushAction(params runParams) error {
	args := []string{
		"push", params.tag,
	}
	if err := params.executor.Run(params.stdout, "docker", args...); err != nil {
		return errors.Wrapf(err, "failed to execute command %v", args)
	}
	return nil
}

func runTagAction(params runParams) error {
	fmt.Fprintln(params.stdout, params.tag)
	return nil
}

func executeDockerBuild(executor Executor, dockerfileContents, tag, dockerfileTemplatePath string, stdout io.Writer) (rerr error) {
	if dockerfileTemplatePath == "" {
		return errors.Errorf("dockerFileLoc must be non-empty")
	}

	f, err := ioutil.TempFile(filepath.Dir(dockerfileTemplatePath), "Dockerfile")
	if err != nil {
		return errors.Wrapf(err, "failed to create temporary file for rendered Dockerfile")
	}
	defer func() {
		if err := os.Remove(f.Name()); err != nil && rerr == nil {
			rerr = errors.Wrapf(err, "failed to remove temporary file for rendered Dockerfile")
		}
	}()

	if _, err := f.WriteString(dockerfileContents); err != nil {
		return errors.Wrapf(err, "failed to write Dockerfile")
	}
	if err := f.Close(); err != nil {
		return errors.Wrapf(err, "failed to close file")
	}

	args := []string{
		"build", "-t", tag, "-f", f.Name(), filepath.Dir(dockerfileTemplatePath),
	}
	if err := executor.Run(stdout, "docker", args...); err != nil {
		return errors.Wrapf(err, "failed to execute command %v", args)
	}
	return nil
}

func executeGoTemplate(tmplContent, buildID string, vars map[string]string, inputTags map[string][][]string, outerIdx, innerIdx int) (string, error) {
	funcs := template.FuncMap{
		"Getenv":  os.Getenv,
		"BuildID": func() string { return buildID },
		"Tag": func(image string, i, j int) (string, error) {
			tagSlice, ok := inputTags[image]
			if !ok {
				return "", fmt.Errorf("unknown image name %s", image)
			}
			if i >= len(tagSlice) {
				return "", fmt.Errorf("outer index out of bounds: %d > %d", i, len(tagSlice))
			}
			if j >= len(tagSlice[i]) {
				return "", fmt.Errorf("inner index out of bounds: %d > %d", j, len(tagSlice[i]))
			}
			return tagSlice[i][j], nil
		},
		"OuterIdx": func() (int, error) {
			if outerIdx < 0 {
				return 0, fmt.Errorf("OuterIdx was not set")
			}
			return outerIdx, nil
		},
		"InnerIdx": func() (int, error) {
			if innerIdx < 0 {
				return 0, fmt.Errorf("InnerIdx was not set")
			}
			return innerIdx, nil
		},
	}
	tmpl, err := template.New("env").Funcs(funcs).Parse(tmplContent)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse template")
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, vars); err != nil {
		return "", errors.Wrapf(err, "failed to execute template")
	}
	return buf.String(), nil
}
