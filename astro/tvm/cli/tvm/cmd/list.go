/*
 *  Copyright (c) 2019 Uber Technologies, Inc.
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
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"

	version "github.com/burl/go-version"
	"github.com/spf13/cobra"

	"github.com/uber/astro/astro/tvm"
)

var listCmd = &cobra.Command{
	Use:   "ls",
	Short: "List locally downloaded versions of Terraform",
	Run: func(cmd *cobra.Command, args []string) {
		tvm, err := tvm.NewVersionRepoForCurrentSystem(repoPath)
		if err != nil {
			log.Fatal(err)
		}

		// Get list of downloaded versions and path to binaries
		versionsPaths, err := tvm.List()
		if err != nil {
			log.Fatal(err)
		}

		// Extract just the version strings
		versions := []string{}
		for v := range versionsPaths {
			versions = append(versions, v)
		}

		// Get the path to the current Terraform binary, according to $PATH
		terraformPath, _ := exec.LookPath("terraform")

		// Get path that Terraform binary links to
		terraformLinkPath, _ := os.Readlink(terraformPath)

		// List the versions
		for _, v := range sortedVersions(versions) {
			s := v.String()
			print(s)

			// If the current Terraform binary is linked to a particular
			// path from tvm, then it's active and has been installed by tvm
			if versionsPaths[s] == terraformLinkPath {
				fmt.Printf(" (current, installed at: %s)", terraformPath)
			}

			println()
		}
	},
}

// sortedVersions takes a list of version strings and returns them sorted in
// reverse order.
func sortedVersions(versions []string) (sortedVersions version.Collection) {
	for _, v := range versions {
		semver, err := version.NewVersion(v)
		if err != nil {
			log.Println(err)
			continue
		}

		sortedVersions = append(sortedVersions, semver)
	}

	sort.Sort(sort.Reverse(sortedVersions))

	return
}

func init() {
	rootCmd.AddCommand(listCmd)
}
