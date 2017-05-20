// Copyright 2017 Nick Miyake. All rights reserved.
// Licensed under the MIT License. See LICENSE in the project root
// for license information.

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/nmiyake/dockergen/dockergen"
)

var (
	cfgFile string
	cfg     dockergen.Config
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dockergen",
	Short: "Builds, tags and publishes Dockerfiles based on templates",
	Long: `Dockergen allows Dockerfiles to be generated programatically from templates
based on declarative configuration.`,
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.SilenceErrors = true
	RootCmd.SilenceUsage = true

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if cfgFile == "" {
			return errors.Errorf("config flag is required")
		}
		cfgBytes, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			return errors.Wrapf(err, "failed to read config file")
		}
		if err := yaml.Unmarshal(cfgBytes, &cfg); err != nil {
			return errors.Wrapf(err, "failed to unmarshal configuration")
		}
		return nil
	}
}
