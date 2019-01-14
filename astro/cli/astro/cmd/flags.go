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
	"os"
	"sort"
	"strings"

	"github.com/uber/astro/astro"
	"github.com/uber/astro/astro/conf"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Text that will be appended to the --help output for plan/apply, showing
// user flags from the astro project config.
const userHelpTemplate = `
User flags:
{{.ProjectFlagsHelp}}`

// ProjectFlag is a CLI flag that represents a variable from the user's astro
// project config file.
type ProjectFlag struct {
	// Name of the user flag at the command line
	Name string
	// Value is the string var the flag value will be put into
	Value string
	// Description is what shows up next to the flag in --help
	Description string
	// Variable is the name of the user variable from the user's astro project
	// config that this flag maps to.
	Variable string
	// AllowedValues is the list of valid values for this flag
	AllowedValues []string
}

// AddToFlagSet adds the flag to the specified flag set.
func (flag *ProjectFlag) AddToFlagSet(flags *pflag.FlagSet) {
	if len(flag.AllowedValues) > 0 {
		flags.Var(&StringEnum{Flag: flag}, flag.Name, flag.Description)
	} else {
		flags.StringVar(&flag.Value, flag.Name, "", flag.Description)
	}
}

// StringEnum implements pflag.Value interface, to check that the passed-in
// value is one of the strings in AllowedValues.
type StringEnum struct {
	Flag *ProjectFlag
}

// String returns the current value
func (s *StringEnum) String() string {
	return s.Flag.Value
}

// Set checks that the passed-in value is only of the allowd values, and
// returns an error if it is not
func (s *StringEnum) Set(value string) error {
	for _, allowedValue := range s.Flag.AllowedValues {
		if allowedValue == value {
			s.Flag.Value = value
			return nil
		}
	}
	return fmt.Errorf("allowed values: %s", strings.Join(s.Flag.AllowedValues, ", "))
}

func (s *StringEnum) Type() string {
	return "string"
}

// addProjectFlagsToCommands adds the user flags to the specified Cobra commands.
func addProjectFlagsToCommands(flags []*ProjectFlag, cmds ...*cobra.Command) {
	ProjectFlagSet := flagsToFlagSet(flags)

	for _, cmd := range cmds {
		for _, flag := range flags {
			flag.AddToFlagSet(cmd.Flags())
		}

		// Update help text for the command to include the user flags
		helpTmpl := cmd.HelpTemplate()
		helpTmpl += "\nUser flags:\n"
		helpTmpl += ProjectFlagSet.FlagUsages()

		cmd.SetHelpTemplate(helpTmpl)

		// Mark flag hidden so it doesn't appear in the normal help. We have to
		// do this *after* calling FlagUsages above, otherwise the flags don't
		// appear in the output.
		for _, flag := range flags {
			cmd.Flags().MarkHidden(flag.Name)
		}
	}
}

// Load the astro configuration file and read flags from the project config.
func loadProjectFlagsFromConfig() ([]*ProjectFlag, error) {
	findConfig := &cobra.Command{
		SilenceUsage:  true,
		SilenceErrors: true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}

	// Strip the help options from os.Args so that the pre-loading of the
	// config doesn't fail with pflag.ErrHelp
	args := []string{}
	for _, arg := range os.Args {
		if arg == "-h" || arg == "--help" || arg == "-help" {
			continue
		}
		args = append(args, arg)
	}

	// Do an early first parse of the config flag before the main command,
	findConfig.PersistentFlags().StringVar(&userCfgFile, "config", "", "config file")
	if err := findConfig.ParseFlags(args); err != nil {
		return nil, err
	}

	config, err := currentConfig()
	if err != nil {
		return nil, err
	}

	return flagsFromConfig(config), nil
}

// flagsFromConfig reads the astro config and returns a list of ProjectFlags that
// can be used to fill in astro variable values at runtime.
func flagsFromConfig(config *conf.Project) (flags []*ProjectFlag) {
	flagMap := map[string]*ProjectFlag{}

	for _, moduleConf := range config.Modules {
		for _, variableConf := range moduleConf.Variables {
			var flagName string
			var flagConf conf.Flag

			// Check for flag mapping in project configuration. This is a block
			// in the configuration that allows users to remap variable names
			// on the CLI and set a description for the --help message.
			flagConf, flagConfExists := config.Flags[variableConf.Name]
			if flagConfExists {
				flagName = flagConf.Name
			} else {
				flagName = variableConf.Name
			}

			if flag, ok := flagMap[flagName]; ok {
				// aggregate values from all variables in the config
				flag.AllowedValues = uniqueStrings(append(flag.AllowedValues, variableConf.Values...))
			} else {
				flag := &ProjectFlag{
					Name:          flagName,
					Description:   flagConf.Description,
					Variable:      variableConf.Name,
					AllowedValues: variableConf.Values,
				}

				flagMap[variableConf.Name] = flag
			}
		}
	}

	// return as list
	for _, flag := range flagMap {
		flags = append(flags, flag)
	}

	return flags
}

// Create an astro.UserVariables suitable for passing into ExecutionParameters
// from the user flags.
func flagsToUserVariables() *astro.UserVariables {
	values := make(map[string]string)
	filters := make(map[string]bool)

	for _, flag := range projectFlags {
		if flag.Value != "" {
			values[flag.Variable] = flag.Value
			if len(flag.AllowedValues) > 0 {
				filters[flag.Variable] = true
			}
		}
	}

	return &astro.UserVariables{
		Values:  values,
		Filters: filters,
	}
}

// Converts a list of ProjectFlags to a pflag.FlagSet.
func flagsToFlagSet(flags []*ProjectFlag) *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("ProjectFlags", pflag.ContinueOnError)
	for _, flag := range flags {
		flag.AddToFlagSet(flagSet)
	}
	return flagSet
}

// flagName returns the flag name, given a variable name.
func flagName(variableName string) string {
	if flag, ok := _conf.Flags[variableName]; ok {
		return flag.Name
	}
	return variableName
}

// varsToFlagNames converts a list of variable names to CLI flags.
func varsToFlagNames(variableNames []string) (flagNames []string) {
	for _, v := range variableNames {
		flagNames = append(flagNames, fmt.Sprintf("--%s", flagName(v)))
	}
	return flagNames
}

func uniqueStrings(strings []string) []string {
	sort.Strings(strings)
	pos := 0
	prev := ""
	for i, s := range strings {
		if i == 0 || prev != s {
			strings[pos] = s
			prev = s
			pos++
		}
	}

	return strings[0:pos]
}
