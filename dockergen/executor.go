// Copyright 2017 Nick Miyake. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root
// for license information.

package dockergen

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Executor interface {
	Run(w io.Writer, cmd string, args ...string) error
}

func NewCmdExecutor() Executor {
	return &cmdExecutor{}
}

type cmdExecutor struct{}

func (e *cmdExecutor) Run(w io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = w
	cmd.Stderr = w
	return cmd.Run()
}

func NewPrintCmdExecutor() Executor {
	return &printCmdExecutor{}
}

type printCmdExecutor struct{}

func (e *printCmdExecutor) Run(w io.Writer, name string, args ...string) error {
	_, err := io.WriteString(w, fmt.Sprintln(name, strings.Join(args, " ")))
	return err
}
