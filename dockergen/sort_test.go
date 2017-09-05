// Copyright 2017 Nick Miyake. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root
// for license information.

package dockergen_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/nmiyake/dockergen/dockergen"
)

func TestTopologicalSort(t *testing.T) {
	for i, tc := range []struct {
		name string
		yml  string
		want []string
	}{
		{
			"no dependencies maintains order",
			`
builds:
  foo:
  baz:
  bar:
`,
			[]string{
				"foo",
				"baz",
				"bar",
			},
		},
		{
			"dependencies maintains original order",
			`
builds:
  foo:
    requires:
      - "other"
      - "bar"
      - "baz"
      - "abc"
  baz:
  bar:
  foo-2:
    requires:
      - bar-2
  abc:
  other:
  bar-2:
`,
			[]string{
				"baz",
				"bar",
				"abc",
				"other",
				"foo",
				"bar-2",
				"foo-2",
			},
		},
		{
			"dependencies maintains original order alternate",
			`
builds:
  foo:
    requires:
      - "baz"
      - "other"
      - "abc"
      - "bar"
  other:
  abc:
  bar:
  baz:
`,
			[]string{
				"other",
				"abc",
				"bar",
				"baz",
				"foo",
			},
		},
		{
			"simple sort",
			`
builds:
  bar:
  foo:
    requires:
      - bar
`,
			[]string{
				"bar",
				"foo",
			},
		},
		{
			"multi-level sort",
			`
builds:
  foo:
    requires:
      - bar
      - baz
  baz:
  bar:
    requires:
      - baz
`,
			[]string{
				"baz",
				"bar",
				"foo",
			},
		},
		{
			"multi-level complicated sort",
			`
builds:
  five:
    requires:
      - two
      - zero
  four:
    requires:
      - zero
      - one
  two:
    requires:
      - three
  zero:
  one:
  three:
    requires:
      - one
`,
			[]string{
				"zero",
				"one",
				"four",
				"three",
				"two",
				"five",
			},
		},
	} {
		var cfg dockergen.Config
		err := yaml.Unmarshal([]byte(tc.yml), &cfg)
		require.NoError(t, err, "Case %d: %s", i, tc.name)

		bParams, err := cfg.BuildParams()
		require.NoError(t, err, "Case %d: %s", i, tc.name)

		sorted := dockergen.TopologicalSort(bParams)
		var got []string
		for _, curr := range sorted {
			got = append(got, curr.Name)
		}
		assert.Equal(t, tc.want, got, "Case %d: %s", i, tc.name)
	}
}

func TestRequiredBuilds(t *testing.T) {
	for i, tc := range []struct {
		name  string
		yml   string
		input string
		want  []string
	}{
		{
			"no dependencies returns self",
			`
builds:
  foo:
  baz:
  bar:
`,
			"foo",
			[]string{
				"foo",
			},
		},
		{
			"single dependency returns self and dependency",
			`
builds:
  bar:
  foo:
    requires:
      - bar
`,
			"foo",
			[]string{
				"bar",
				"foo",
			},
		},
		{
			"multi-level",
			`
builds:
  foo:
    requires:
      - bar
      - baz
  baz:
  bar:
    requires:
      - baz
`,
			"foo",
			[]string{
				"bar",
				"baz",
				"foo",
			},
		},
		{
			"multi-level complicated",
			`
builds:
  five:
    requires:
      - two
      - zero
  four:
    requires:
      - zero
      - one
  two:
    requires:
      - three
  zero:
  one:
  three:
    requires:
      - one
`,
			"five",
			[]string{
				"five",
				"one",
				"three",
				"two",
				"zero",
			},
		},
	} {
		var cfg dockergen.Config
		err := yaml.Unmarshal([]byte(tc.yml), &cfg)
		require.NoError(t, err, "Case %d: %s", i, tc.name)

		bParams, err := cfg.BuildParams()
		require.NoError(t, err, "Case %d: %s", i, tc.name)

		paramsMap := make(map[string]dockergen.BuildParams)
		for _, param := range bParams {
			paramsMap[param.Name] = param
		}

		requires := dockergen.RequiredBuilds(paramsMap[tc.input], bParams)
		var got []string
		for _, param := range requires {
			got = append(got, param.Name)
		}
		assert.Equal(t, tc.want, got, "Case %d: %s", i, tc.name)
	}
}
