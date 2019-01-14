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
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/uber/astro/astro"
)

var (
	detach            bool
	moduleNamesString string
)

var applyCmd = &cobra.Command{
	Use:                   "apply [flags] [-- [Terraform argument]...]",
	DisableFlagsInUseLine: true,
	Short:                 "Run Terraform apply on all modules",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := currentProject()
		if err != nil {
			return err
		}

		vars := flagsToUserVariables()

		var moduleNames []string
		if moduleNamesString != "" {
			moduleNames = strings.Split(moduleNamesString, ",")
		}

		status, results, err := c.Apply(
			astro.ApplyExecutionParameters{
				ExecutionParameters: astro.ExecutionParameters{
					ModuleNames:         moduleNames,
					UserVars:            vars,
					TerraformParameters: args,
				},
			},
		)
		if err != nil {
			return fmt.Errorf("ERROR: %v", processError(err))
		}

		err = printExecStatus(status, results)
		if err != nil {
			return fmt.Errorf("Done; there were errors; some modules may not have been applied")
		}

		fmt.Println("Done")

		return nil
	},
}

var planCmd = &cobra.Command{
	Use:                   "plan [flags] [-- [Terraform argument]...]",
	DisableFlagsInUseLine: true,
	Short:                 "Generate execution plans for modules",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := currentProject()
		if err != nil {
			return err
		}

		vars := flagsToUserVariables()

		var moduleNames []string
		if moduleNamesString != "" {
			moduleNames = strings.Split(moduleNamesString, ",")
		}

		status, results, err := c.Plan(
			astro.PlanExecutionParameters{
				ExecutionParameters: astro.ExecutionParameters{
					ModuleNames:         moduleNames,
					UserVars:            vars,
					TerraformParameters: args,
				},
				Detach: detach,
			},
		)
		if err != nil {
			return fmt.Errorf("ERROR: %v", processError(err))
		}

		err = printExecStatus(status, results)
		if err != nil {
			return errors.New("Done; there were errors")
		}

		fmt.Println("Done")

		return nil
	},
}

func init() {
	applyCmd.PersistentFlags().StringVar(&moduleNamesString, "modules", "", "list of modules to apply")
	rootCmd.AddCommand(applyCmd)

	planCmd.PersistentFlags().BoolVar(&detach, "detach", false, "disconnect remote state before planning")
	planCmd.PersistentFlags().StringVar(&moduleNamesString, "modules", "", "list of modules to plan")
	rootCmd.AddCommand(planCmd)
}

// processError interprets certain astro errors and embellishes them for
// display on the CLI.
func processError(err error) error {
	switch e := err.(type) {
	case astro.MissingRequiredVarsError:
		// reverse map variables to CLI flags
		return fmt.Errorf("missing required flags: %s", strings.Join(varsToFlagNames(e.MissingVars()), ", "))
	default:
		return err
	}
}
