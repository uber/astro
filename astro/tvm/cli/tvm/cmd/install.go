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
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/uber/astro/astro/tvm"
)

// defaultInstallPath is the path that the Terraform binary will be
// linked on the system.
const defaultInstallPath = "/usr/local/bin/terraform"

var (
	installPath string
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Download and link the specified version of Terraform",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tvm, err := tvm.NewVersionRepoForCurrentSystem(repoPath)
		if err != nil {
			log.Fatal(err)
		}

		version := args[0]

		if err := tvm.Link(version, viper.GetString("installPath"), true); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	installCmd.PersistentFlags().StringVar(
		&installPath, "path", "",
		fmt.Sprintf("path to link Terraform binary to (default: %s )", defaultInstallPath),
	)

	viper.BindPFlag("path", installCmd.PersistentFlags().Lookup("path"))
	viper.SetDefault("installPath", defaultInstallPath)

	rootCmd.AddCommand(installCmd)
}
