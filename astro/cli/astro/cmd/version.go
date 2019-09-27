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
	"strings"

	"github.com/spf13/cobra"
)

var (
	// When a release happens, the value of this variable will be overwritten
	// by the linker to match the release version.
	version = "dev"
	commit  = ""
	date    = ""
)

func (cli *AstroCLI) createVersionCmd() {
	versionCmd := &cobra.Command{
		Use:                   "version",
		DisableFlagsInUseLine: true,
		Short:                 "Print astro version",
		RunE: func(cmd *cobra.Command, args []string) error {
			versionString := []string{
				"astro version",
				version,
			}

			if commit != "" {
				versionString = append(versionString, fmt.Sprintf("(%s)", commit))
			}

			if date != "" {
				versionString = append(versionString, fmt.Sprintf("built %s", date))
			}

			println(strings.Join(versionString, " "))

			return nil
		},
	}
	cli.commands.version = versionCmd
}
