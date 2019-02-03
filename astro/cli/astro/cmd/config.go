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

package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/uber/astro/astro"
	"github.com/uber/astro/astro/utils"
)

// configFileSearchPaths is the default list of paths the astro CLI
// will attempt to find a config file at.
var configFileSearchPaths = []string{
	"astro.yaml",
	"astro.yml",
	"terraform/astro.yaml",
	"terraform/astro.yml",
}

var errCannotFindConfig = errors.New("unable to find config file")

// loadConfig tries to load astro configuration, either from a user specified
// path or from one of the search paths. If no astro config can be found, an
// error is returned.
func (cli *AstroCLI) loadConfig(configFilePath string) error {
	config, err := astro.NewConfigFromFile(configFilePath)
	cli.config = config
	return err
}

// configPathFromArgs reads the command line arguments and returns the value of
// the config option. It returns an empty string if there is no path in the
// args.
func configPathFromArgs(args []string) (configFilePath string) {
	// this is a special cobra command so that we can parse just the config
	// flag early in the program lifecycle.
	findConfig := &cobra.Command{
		SilenceUsage:  true,
		SilenceErrors: true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}

	// Strip the help options from args so that the pre-loading of the config
	// doesn't fail with pflag.ErrHelp
	finalArgs := []string{}
	for _, arg := range args {
		if arg == "-h" || arg == "--help" || arg == "-help" {
			continue
		}
		finalArgs = append(finalArgs, arg)
	}

	// Do an early first parse of the config flag before the main command,
	findConfig.PersistentFlags().StringVar(&configFilePath, "config", "", "config file")
	if err := findConfig.ParseFlags(finalArgs); err != nil {
		return ""
	}

	return configFilePath
}

// resolveConfigFilePath returns the path of the project config file. It
// returns an empty string if no config file is found.
func (cli *AstroCLI) resolveConfigFilePath() (configFilePath string) {
	// Try to find the config file
	if path := firstExistingFilePath(configFileSearchPaths...); path != "" {
		return path
	}

	return ""
}

// firstExistingFilePath takes a list of paths and returns the first one
// where a file exists (or symlink to a file).
func firstExistingFilePath(paths ...string) string {
	for _, f := range paths {
		if utils.FileExists(f) {
			return f
		}
	}
	return ""
}
