// Copyright 2017 Nick Miyake. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root
// for license information.

package dockergen_test

import (
	"fmt"
	"testing"

	"github.com/nmiyake/dockergen/dockergen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestLoadConfigWithCycles(t *testing.T) {
	for i, tc := range []struct {
		name      string
		yml       string
		wantError string
	}{
		{
			"self-referential cycle",
			`
builds:
  foo:
    docker-template: foo/Dockerfile_template.txt
    tag: test/foo:snapshot
    requires:
      - foo
`,
			"product cycle exists: foo -> foo",
		},
		{
			"transitive cycle",
			`
builds:
  foo:
    docker-template: foo/Dockerfile_template.txt
    tag: test/foo:snapshot
    requires:
      - bar
  bar:
    docker-template: bar/Dockerfile_template.txt
    tag: test/bar:snapshot
    requires:
      - foo
`,
			"product cycle exists: foo -> bar -> foo",
		},
	} {
		var cfg dockergen.Config
		err := yaml.Unmarshal([]byte(tc.yml), &cfg)
		require.NoError(t, err, "Case %d: %s", i, tc.name)

		_, err = cfg.BuildParams()
		require.Error(t, err, fmt.Sprintf("Case %d: %s", i, tc.name))
		assert.Regexp(t, tc.wantError, err.Error(), "Case %d: %s", i, tc.name)
	}
}

func TestLoadConfigWithInvalidRequires(t *testing.T) {
	for i, tc := range []struct {
		name      string
		yml       string
		wantError string
	}{
		{
			"requires refers to image that is not specified",
			`
builds:
  foo:
    requires:
      - bar
`,
			"Image foo requires image bar, which is not defined in configuration",
		},
	} {
		var cfg dockergen.Config
		err := yaml.Unmarshal([]byte(tc.yml), &cfg)
		require.NoError(t, err, "Case %d: %s", i, tc.name)

		_, err = cfg.BuildParams()
		require.Error(t, err, fmt.Sprintf("Case %d: %s", i, tc.name))
		assert.Regexp(t, tc.wantError, err.Error(), "Case %d: %s", i, tc.name)
	}
}
