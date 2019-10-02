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
	"errors"
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
		root    *cobra.Command
		plan    *cobra.Command
		apply   *cobra.Command
		version *cobra.Command
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
	cli.createRootCommand()
	cli.createPlanCmd()
	cli.createApplyCmd()
	cli.createVersionCmd()

	cli.commands.root.AddCommand(
		cli.commands.plan,
		cli.commands.apply,
		cli.commands.version,
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

	userProvidedConfigPath, err := configPathFromArgs(args)
	if err != nil {
		fmt.Fprintln(cli.stderr, err.Error())
		return 1
	}

	configFilePath := firstExistingFilePath(
		append([]string{userProvidedConfigPath}, configFileSearchPaths...)...,
	)

	if configFilePath != "" {
		config, err := astro.NewConfigFromFile(configFilePath)
		if err != nil {
			fmt.Fprintln(cli.stderr, err.Error())
			return 1
		}

		cli.config = config
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

func (cli *AstroCLI) createRootCommand() {
	rootCmd := &cobra.Command{
		Use:           "astro",
		Short:         "A tool for managing multiple Terraform modules.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().BoolVarP(&cli.flags.verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&cli.flags.trace, "trace", "", false, "trace output")
	rootCmd.PersistentFlags().StringVar(&cli.flags.userCfgFile, "config", "", "config file")

	cli.commands.root = rootCmd
}

func (cli *AstroCLI) createApplyCmd() {
	applyCmd := &cobra.Command{
		Use:                   "apply [flags] [-- [Terraform argument]...]",
		DisableFlagsInUseLine: true,
		Short:                 "Run Terraform apply on all modules",
		PersistentPreRunE:     cli.preRun,
		RunE:                  cli.runApply,
	}

	applyCmd.PersistentFlags().StringVar(&cli.flags.moduleNamesString, "modules", "", "list of modules to apply")

	cli.commands.apply = applyCmd
}

func (cli *AstroCLI) createPlanCmd() {
	planCmd := &cobra.Command{
		Use:                   "plan [flags] [-- [Terraform argument]...]",
		DisableFlagsInUseLine: true,
		Short:                 "Generate execution plans for modules",
		PersistentPreRunE:     cli.preRun,
		RunE:                  cli.runPlan,
	}

	planCmd.PersistentFlags().BoolVar(&cli.flags.detach, "detach", false, "disconnect remote state before planning")
	planCmd.PersistentFlags().StringVar(&cli.flags.moduleNamesString, "modules", "", "list of modules to plan")

	cli.commands.plan = planCmd
}

func (cli *AstroCLI) preRun(cmd *cobra.Command, args []string) error {
	logger.Trace.Println("cli: in preRun")

	if cli.config == nil {
		return fmt.Errorf("unable to find config file")
	}
	// Load astro from config
	project, err := astro.NewProject(astro.WithConfig(*cli.config))
	if err != nil {
		return err
	}
	cli.project = project

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
