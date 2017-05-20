// Copyright 2017 Nick Miyake. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root
// for license information.

package dockergen

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	// Environment variable that will be used to determine the unique identifier for this build.
	BuildIDVar string `yaml:"build-id-var"`
	// Variables to set for the templates.
	TemplateVars map[string]string `yaml:"template-vars"`
	// Suffix that will be appended to tags. Can use templates.
	TagSuffix string `yaml:"tag-suffix"`
	// If present, specifies variables that will be looped over for all generation tasks. If more than one key is
	// specified, all of the value slices must have the same length. During any single iteration, the name of the
	// key of the map will be the name of the template variable and the value will be the value for the current
	// iteration.
	For map[string][]string `yaml:"for"`
	// All of the build tasks defined for this configuration.
	Builds BuildYMLs `yaml:"builds"`
}

type BuildYMLs yaml.MapSlice // sorted map[string]BuildConfig

func (s *BuildYMLs) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var mapSlice yaml.MapSlice
	if err := unmarshal(&mapSlice); err != nil {
		return err
	}

	// values of MapSlice are known to be BuildConfig, so read them out as such
	for i, v := range mapSlice {
		bytes, err := yaml.Marshal(v.Value)
		if err != nil {
			return err
		}
		var currBuildConfig BuildConfig
		if err := yaml.Unmarshal(bytes, &currBuildConfig); err != nil {
			return err
		}
		mapSlice[i].Value = currBuildConfig
	}
	*s = BuildYMLs(mapSlice)
	return nil
}

func (c *Config) ToParams() Params {
	return Params{
		BuildIDVar:   c.BuildIDVar,
		TemplateVars: c.TemplateVars,
		TagSuffix:    c.TagSuffix,
		For:          c.For,
	}
}

func (c *Config) BuildParams() []BuildParams {
	var params []BuildParams
	for _, v := range c.Builds {
		val := v.Value.(BuildConfig)
		params = append(params, BuildParams{
			Name: v.Key.(string),
			DockerfileTemplatePath: val.DockerTemplatePath,
			Tag: val.Tag,
			For: val.For,
		})
	}
	return params
}

type Params struct {
	BuildIDVar   string
	TemplateVars map[string]string
	TagSuffix    string
	For          map[string][]string
}

func (p *Params) Validate() error {
	if len(p.For) == 0 {
		return nil
	}

	// fail if any template vars and for vars collide
	var duplicateVars []string
	for k := range p.TemplateVars {
		if _, ok := p.For[k]; !ok {
			continue
		}
		duplicateVars = append(duplicateVars, k)
	}
	sort.Strings(duplicateVars)
	if len(duplicateVars) != 0 {
		return fmt.Errorf("the following variables were defined as both template and for variables: %v", duplicateVars)
	}

	var sortedForVarNames []string
	for k := range p.For {
		sortedForVarNames = append(sortedForVarNames, k)
	}
	sort.Strings(sortedForVarNames)

	// verify all "For" variables are of the same length
	forVarLen := -1
	for _, varName := range sortedForVarNames {
		vals := p.For[varName]
		if forVarLen == -1 {
			forVarLen = len(vals)
			continue
		}
		if len(vals) != forVarLen {
			var parts []string
			parts = append(parts, "Length of all outer 'for' variable arrays must be the same:")
			for _, varName := range sortedForVarNames {
				parts = append(parts, fmt.Sprintf("%s: %d", varName, len(p.For[varName])))
			}
			return fmt.Errorf(strings.Join(parts, "\n\t"))
		}
	}
	return nil
}

type BuildConfig struct {
	// Path to the Dockerfile template that should be used to build the Docker image.
	DockerTemplatePath string `yaml:"docker-template"`
	// Tag that will be used for the generated image. Can use templates.
	Tag string `yaml:"tag"`
	// If present, specifies variables that will be looped over for this generation task. If more than one key is
	// specified, all of the value slices must have the same length. During any single iteration, the name of the
	// key of the map will be the name of the template variable and the value will be the value for the current
	// iteration.
	For map[string][]string `yaml:"for"`
}

type BuildParams struct {
	Name                   string
	DockerfileTemplatePath string
	Tag                    string
	For                    map[string][]string
}
