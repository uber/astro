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
	"github.com/uber/astro/astro/logger"
)

func (cli *AstroCLI) createRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "astro",
		Short:             "A tool for managing multiple Terraform modules.",
		SilenceUsage:      true,
		SilenceErrors:     true,
		PersistentPreRunE: cli.preRun,
	}

	rootCmd.PersistentFlags().BoolVarP(&cli.flags.verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&cli.flags.trace, "trace", "", false, "trace output")
	rootCmd.PersistentFlags().StringVar(&cli.flags.userCfgFile, "config", "", "config file")

	return rootCmd
}

func (cli *AstroCLI) createApplyCmd() *cobra.Command {
	applyCmd := &cobra.Command{
		Use:                   "apply [flags] [-- [Terraform argument]...]",
		DisableFlagsInUseLine: true,
		Short:                 "Run Terraform apply on all modules",
		RunE:                  cli.runApply,
	}

	applyCmd.PersistentFlags().StringVar(&cli.flags.moduleNamesString, "modules", "", "list of modules to apply")

	return applyCmd
}

func (cli *AstroCLI) createPlanCmd() *cobra.Command {
	planCmd := &cobra.Command{
		Use:                   "plan [flags] [-- [Terraform argument]...]",
		DisableFlagsInUseLine: true,
		Short:                 "Generate execution plans for modules",
		RunE:                  cli.runPlan,
	}

	planCmd.PersistentFlags().BoolVar(&cli.flags.detach, "detach", false, "disconnect remote state before planning")
	planCmd.PersistentFlags().StringVar(&cli.flags.moduleNamesString, "modules", "", "list of modules to plan")

	return planCmd
}

func (cli *AstroCLI) preRun(cmd *cobra.Command, args []string) error {
	logger.Trace.Println("cli: in preRun")

	// Load astro from config
	project, err := astro.NewProject(astro.WithConfig(*cli.config))
	if err != nil {
		return err
	}
	cli.project = project

	return nil
}

func (cli *AstroCLI) runApply(cmd *cobra.Command, args []string) error {
	vars := flagsToUserVariables(cli.flags.projectFlags)

	var moduleNames []string
	if cli.flags.moduleNamesString != "" {
		moduleNames = strings.Split(cli.flags.moduleNamesString, ",")
	}

	status, results, err := cli.project.Apply(
		astro.ApplyExecutionParameters{
			ExecutionParameters: astro.ExecutionParameters{
				ModuleNames:         moduleNames,
				UserVars:            vars,
				TerraformParameters: args,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("ERROR: %v", cli.processError(err))
	}

	err = cli.printExecStatus(status, results)
	if err != nil {
		return fmt.Errorf("Done; there were errors; some modules may not have been applied")
	}

	fmt.Fprintln(cli.stdout, "Done")

	return nil
}

func (cli *AstroCLI) runPlan(cmd *cobra.Command, args []string) error {
	logger.Trace.Printf("cli: plan args: %s\n", args)

	vars := flagsToUserVariables(cli.flags.projectFlags)

	var moduleNames []string
	if cli.flags.moduleNamesString != "" {
		moduleNames = strings.Split(cli.flags.moduleNamesString, ",")
	}

	status, results, err := cli.project.Plan(
		astro.PlanExecutionParameters{
			ExecutionParameters: astro.ExecutionParameters{
				ModuleNames:         moduleNames,
				UserVars:            vars,
				TerraformParameters: args,
			},
			Detach: cli.flags.detach,
		},
	)
	if err != nil {
		return fmt.Errorf("ERROR: %v", cli.processError(err))
	}

	err = cli.printExecStatus(status, results)
	if err != nil {
		return errors.New("Done; there were errors")
	}

	fmt.Fprintln(cli.stdout, "Done")

	return nil
}

// processError interprets certain astro errors and embellishes them for
// display on the CLI.
func (cli *AstroCLI) processError(err error) error {
	switch e := err.(type) {
	case astro.MissingRequiredVarsError:
		// reverse map variables to CLI flags
		return fmt.Errorf("missing required flags: %s", strings.Join(cli.varsToFlagNames(e.MissingVars()), ", "))
	default:
		return err
	}
}
