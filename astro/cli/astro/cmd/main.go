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
// that operators use to interact with the project.
package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/uber/astro/astro"
	"github.com/uber/astro/astro/conf"
	"github.com/uber/astro/astro/logger"

	"github.com/spf13/cobra"
)

func init() {
	// silence trace info from terraform/dag by default
	log.SetOutput(ioutil.Discard)
}

// AstroCLI is the main CLI program, where flags and state are stored for the
// running program.
type AstroCLI struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	project *astro.Project
	config  *conf.Project

	// these values are filled in based on runtime flags
	flags struct {
		detach            bool
		moduleNamesString string
		trace             bool
		userCfgFile       string
		verbose           bool

		// projectFlags are special in that the actual flags are dynamic, based
		// on the astro project configuration loaded.
		projectFlags []*projectFlag
	}

	commands struct {
		root  *cobra.Command
		plan  *cobra.Command
		apply *cobra.Command
	}
}

// NewAstroCLI creates a new AstroCLI.
func NewAstroCLI(opts ...Option) (*AstroCLI, error) {
	cli := &AstroCLI{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}

	if err := cli.applyOptions(opts...); err != nil {
		return nil, err
	}

	// Set up Cobra commands and structure
	cli.commands.root = cli.createRootCommand()
	cli.commands.plan = cli.createPlanCmd()
	cli.commands.apply = cli.createApplyCmd()

	cli.commands.root.AddCommand(
		cli.commands.plan,
		cli.commands.apply,
	)

	// Set trace. Note, this will turn tracing on for all instances of astro
	// running in the same process, as the logger is a singleton. This should
	// only be of concern during testing.
	cobra.OnInitialize(func() {
		if cli.flags.trace {
			logger.Trace.SetOutput(cli.stderr)
			log.SetOutput(cli.stderr)
		}
	})

	return cli, nil
}

// Run is the main entry point into the CLI program.
func (cli *AstroCLI) Run(args []string) (exitCode int) {
	cli.commands.root.SetArgs(args)
	cli.commands.root.SetOutput(cli.stderr)

	userSpecifiedConfigFile := configPathFromArgs(args)
	if userSpecifiedConfigFile != "" {
		if err := cli.loadConfig(userSpecifiedConfigFile); err != nil {
			fmt.Fprintln(cli.stderr, err.Error())
			return 1
		}
	}

	cli.configureDynamicUserFlags()

	if err := cli.commands.root.Execute(); err != nil {
		fmt.Fprintln(cli.stderr, err.Error())
		exitCode = 1 // exit with error

		// If we get an unknown flag, it could be because the user expected
		// config to be loaded but it wasn't. Display a message to the user to
		// let them know.
		if cli.config == nil && strings.Contains(err.Error(), "unknown flag") {
			fmt.Fprintln(cli.stderr, "NOTE: No astro config was loaded.")
		}
	}

	return exitCode
}

// configureDynamicUserFlags dynamically adds Cobra flags based on the loaded
// configuration.
func (cli *AstroCLI) configureDynamicUserFlags() {
	projectFlags := flagsFromConfig(cli.config)
	addProjectFlagsToCommands(projectFlags,
		cli.commands.plan,
		cli.commands.apply,
	)
	cli.flags.projectFlags = projectFlags
}
