/*
 *  Copyright (c) 2018 Uber Technologies, Inc.
 *
 *     Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package cmd contains the source for the `astro` command line tool
// that operators use to interact with the project. The layout of files
// in this package is defined by Cobra, which is the library that powers
// the CLI tool.
package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/uber/astro/astro"
	"github.com/uber/astro/astro/conf"
	"github.com/uber/astro/astro/logger"

	"github.com/spf13/cobra"
)

var (
	userCfgFile string
	trace       bool
	userVars    map[string]string
	verbose     bool

	_flags   []*Flag
	_conf    *conf.Project
	_project *astro.Project
)

var rootCmd = &cobra.Command{
	Use:           "astro",
	Short:         "A tool for managing multiple Terraform modules.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	// Best effort to parse flags from config files for now. TODO: we need to
	// deal with errors here.
	initUserFlagsFromConfig()
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&trace, "trace", "", false, "trace output")
	rootCmd.PersistentFlags().StringVar(&userCfgFile, "config", "", "config file")

	// silence trace info from terraform/dag
	log.SetOutput(ioutil.Discard)

	// Set trace
	cobra.OnInitialize(func() {
		if trace {
			logger.Trace.SetOutput(os.Stderr)
		}
	})
}

func initUserFlagsFromConfig() error {
	findConfig := &cobra.Command{
		SilenceUsage:  true,
		SilenceErrors: true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}

	findConfig.PersistentFlags().StringVar(&userCfgFile, "config", "", "config file")

	if err := findConfig.ParseFlags(os.Args); err != nil {
		return err
	}
	config, err := currentConfig()
	if err != nil {
		return err
	}

	_flags, err = commandLineFlags(config)
	if err != nil {
		return err
	}
	for _, flag := range _flags {
		usage := fmt.Sprintf("Set \"%s\" Terraform variable", flag.Variable)
		for _, command := range rootCmd.Commands() {
			if len(flag.AllowedValues) > 0 {
				command.Flags().Var(&StringEnum{Flag: flag}, flag.Flag, usage)
			} else {
				command.Flags().StringVar(&flag.Value, flag.Flag, "", usage)
			}
			if flag.IsRequired {
				command.MarkFlagRequired(flag.Flag)
			}
		}
	}

	return nil
}

// configFile returns the path of the project config file.
func configFile() (string, error) {
	// User provided config file path takes precedence
	if userCfgFile != "" {
		return userCfgFile, nil
	}

	// Try to find the config file
	if path := firstExistingFilePath(configFileSearchPaths...); path != "" {
		return path, nil
	}

	return "", errors.New("unable to find config file")
}

func currentConfig() (*conf.Project, error) {
	if _conf != nil {
		return _conf, nil
	}

	file, err := configFile()
	if err != nil {
		return nil, err
	}
	_conf, err = astro.NewConfigFromFile(file)

	return _conf, err
}

func currentProject() (*astro.Project, error) {
	if _project != nil {
		return _project, nil
	}

	config, err := currentConfig()
	if err != nil {
		return nil, err
	}
	c, err := astro.NewProject(*config)
	if err != nil {
		return nil, fmt.Errorf("unable to load module configuration: %v", err)
	}

	_project = c

	return _project, nil
}
