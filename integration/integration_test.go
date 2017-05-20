// Copyright 2017 Nick Miyake. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root
// for license information.

package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/nmiyake/pkg/dirs"
	"github.com/palantir/godel/pkg/products"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {
	cli, err := products.Bin("dockergen")
	require.NoError(t, err)

	tmpDir, cleanup, err := dirs.TempDir("", "")
	defer cleanup()
	require.NoError(t, err)

	for i, currCase := range []struct {
		name         string
		config       string
		filesToWrite map[string]string
		wantRegexp   string
	}{
		{
			name: "simple build with defaults",
			config: `
builds:
  test-template:
    docker-template: Dockerfile_template.txt
    tag: testuser/foo:bar
`,
			filesToWrite: map[string]string{
				"Dockerfile_template.txt": `FROM scratch
ENV foo bar
`,
			},
			wantRegexp: `(?s).+
Step 2/2 : ENV foo bar.+
Successfully built [0-9a-f]+
Successfully tagged testuser/foo:bar-unspecified
`,
		},
		{
			name: "simple build with variables",
			config: `
build-id-var: CIRCLE_BUILD_NUM
template-vars:
  myTmplVar1: myTmplVal1
builds:
  test-template:
    docker-template: Dockerfile_template.txt
    tag: testuser/foo:bar
`,
			filesToWrite: map[string]string{
				"Dockerfile_template.txt": `FROM scratch
ENV foo {{.myTmplVar1}}
`,
			},
			wantRegexp: `(?s).+
Step 2/2 : ENV foo myTmplVal1.+
Successfully built [0-9a-f]+
Successfully tagged testuser/foo:bar-13
`,
		},
		{
			name: "build with outer for loop only",
			config: `
build-id-var: CIRCLE_BUILD_NUM
tag-suffix: -t{{BuildID}}
template-vars:
  myTmplVar1: myTmplVal1
for:
  loopVar1:
    - hello
    - world
  loopVar2:
    - farewell
    - friends
builds:
  test-template:
    docker-template: Dockerfile_template.txt
    tag: testuser/foo:bar-{{.loopVar1}}
`,
			filesToWrite: map[string]string{
				"Dockerfile_template.txt": `FROM scratch
ENV foo {{.myTmplVar1}}
ENV {{.loopVar1}} {{.loopVar2}}
`,
			},
			wantRegexp: `(?s).+
Step 2/3 : ENV foo myTmplVal1.+
Step 3/3 : ENV hello farewell.+
Successfully built [0-9a-f]+
Successfully tagged testuser/foo:bar-hello-t13.+
Step 2/3 : ENV foo myTmplVal1.+
Step 3/3 : ENV world friends.+
Successfully built [0-9a-f]+
Successfully tagged testuser/foo:bar-world-t13
`,
		},
		{
			name: "build with inner for loop only",
			config: `
build-id-var: CIRCLE_BUILD_NUM
tag-suffix: -t{{BuildID}}
template-vars:
  myTmplVar1: myTmplVal1
builds:
  test-template:
    docker-template: Dockerfile_template.txt
    tag: testuser/foo:bar-{{.loopVar1}}
    for:
      loopVar1:
        - hello
        - world
      loopVar2:
        - farewell
        - friends
`,
			filesToWrite: map[string]string{
				"Dockerfile_template.txt": `FROM scratch
ENV foo {{.myTmplVar1}}
ENV {{.loopVar1}} {{.loopVar2}}
`,
			},
			wantRegexp: `(?s).+
Step 2/3 : ENV foo myTmplVal1.+
Step 3/3 : ENV hello farewell.+
Successfully built [0-9a-f]+
Successfully tagged testuser/foo:bar-hello-t13.+
Step 2/3 : ENV foo myTmplVal1.+
Step 3/3 : ENV world friends.+
Successfully built [0-9a-f]+
Successfully tagged testuser/foo:bar-world-t13
`,
		},
		{
			name: "build with outer and inner for loop",
			config: `
build-id-var: CIRCLE_BUILD_NUM
tag-suffix: -t{{BuildID}}
template-vars:
  myTmplVar1: myTmplVal1
for:
  outerLoopVar1:
    - outer-hello
    - outer-world
  outerLoopVar2:
    - outer-farewell
    - outer-friends
builds:
  foo-template:
    docker-template: foo/Dockerfile_template.txt
    tag: testuser/test:foo-{{.outerLoopVar1}}-{{.innerLoopVar1}}
    for:
      innerLoopVar1:
        - inner-hello
        - inner-world
      innerLoopVar2:
        - inner-farewell
        - inner-friends
  bar-template:
    docker-template: bar/Dockerfile_template.txt
    tag: testuser/test:bar-{{.outerLoopVar1}}
`,
			filesToWrite: map[string]string{
				"foo/Dockerfile_template.txt": `FROM scratch
ENV foo {{.myTmplVar1}}
ENV {{.outerLoopVar1}} {{.outerLoopVar2}}
ENV {{.innerLoopVar1}} {{.innerLoopVar2}}
`,
				"bar/Dockerfile_template.txt": `FROM scratch
ENV bar {{.myTmplVar1}}
ENV {{.outerLoopVar1}} {{.outerLoopVar2}}
`,
			},
			wantRegexp: `(?s).+
Step 2/4 : ENV foo myTmplVal1.+
Step 3/4 : ENV outer-hello outer-farewell.+
Step 4/4 : ENV inner-hello inner-farewell.+
Successfully built [0-9a-f]+
Successfully tagged testuser/test:foo-outer-hello-inner-hello-t13.+
Step 2/4 : ENV foo myTmplVal1.+
Step 3/4 : ENV outer-hello outer-farewell.+
Step 4/4 : ENV inner-world inner-friends.+
Successfully built [0-9a-f]+
Successfully tagged testuser/test:foo-outer-hello-inner-world-t13.+
Step 2/3 : ENV bar myTmplVal1.+
Step 3/3 : ENV outer-hello outer-farewell.+
Successfully built [0-9a-f]+
Successfully tagged testuser/test:bar-outer-hello-t13.+
Step 2/4 : ENV foo myTmplVal1.+
Step 3/4 : ENV outer-world outer-friends.+
Step 4/4 : ENV inner-hello inner-farewell.+
Successfully built [0-9a-f]+
Successfully tagged testuser/test:foo-outer-world-inner-hello-t13.+
Step 2/4 : ENV foo myTmplVal1.+
Step 3/4 : ENV outer-world outer-friends.+
Step 4/4 : ENV inner-world inner-friends.+
Successfully built [0-9a-f]+
Successfully tagged testuser/test:foo-outer-world-inner-world-t13.+
Step 2/3 : ENV bar myTmplVal1.+
Step 3/3 : ENV outer-world outer-friends.+
Successfully built [0-9a-f]+
Successfully tagged testuser/test:bar-outer-world-t13
`,
		},
	} {
		currCaseDir, err := ioutil.TempDir(tmpDir, fmt.Sprintf("case-%d", i))
		require.NoError(t, err, "Case %d: %s", i, currCase.name)

		configFile := path.Join(currCaseDir, "config.yml")
		err = ioutil.WriteFile(configFile, []byte(currCase.config), 0644)
		require.NoError(t, err)

		for k, v := range currCase.filesToWrite {
			currPath := path.Join(currCaseDir, k)
			if dirPath := path.Dir(currPath); dirPath != "." {
				err := os.MkdirAll(dirPath, 0755)
				require.NoError(t, err, "Case %d: %s", i, currCase.name)
			}
			err := ioutil.WriteFile(path.Join(currCaseDir, k), []byte(v), 0644)
			require.NoError(t, err, "Case %d: %s", i, currCase.name)
		}

		args := []string{"--config", configFile}
		args = append(args, "build")
		cmd := exec.Command(cli, args...)
		cmd.Dir = currCaseDir
		cmd.Env = append(os.Environ(), "CIRCLE_BUILD_NUM=13")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Output: %s", string(output))

		assert.Regexp(t, currCase.wantRegexp, string(output), "Case %d: %s", i, currCase.name)
	}
}
