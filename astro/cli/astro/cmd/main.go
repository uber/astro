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
	"io/ioutil"
	"log"
	"os"

	"github.com/uber/astro/astro/logger"

	"github.com/spf13/cobra"
)

// CLI flags
var (
	trace        bool
	userCfgFile  string
	projectFlags []*ProjectFlag
	verbose      bool
)

var rootCmd = &cobra.Command{
	Use:           "astro",
	Short:         "A tool for managing multiple Terraform modules.",
	SilenceUsage:  true,
	SilenceErrors: true,
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

// Main is the main entry point into the CLI program.
func Main() (exitCode int) {
	// Try to parse user flags from their astro config file. Reading the astro
	// config could fail with an error, e.g. if there is no config file found,
	// but this is not a hard failure. Save the error for later, so we can let
	// the user know about the error in certain cases.
	projectFlags, projectFlagsLoadErr := loadProjectFlagsFromConfig()
	if projectFlags != nil && len(projectFlags) > 0 {
		addProjectFlagsToCommands(projectFlags, applyCmd, planCmd)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())

		// If there was an error when parsing the user's project config file,
		// then display a message to let them know in case they're wondering
		// why their CLI flags are not working.
		if projectFlagsLoadErr != nil && projectFlagsLoadErr != errCannotFindConfig {
			fmt.Fprintf(os.Stderr, "\nThere was an error loading astro config:\n")
			fmt.Fprintln(os.Stderr, projectFlagsLoadErr.Error())
		}

		// exit with error
		return 1
	}

	// success
	return 0
}
