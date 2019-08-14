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
{{.projectFlagsHelp}}`

// projectFlag is a CLI flag that represents a variable from the user's astro
// project config file.
type projectFlag struct {
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
func (flag *projectFlag) AddToFlagSet(flags *pflag.FlagSet) {
	if len(flag.AllowedValues) > 0 {
		flags.Var(&stringEnum{flag: flag}, flag.Name, flag.Description)
	} else {
		flags.StringVar(&flag.Value, flag.Name, "", flag.Description)
	}
}

// stringEnum implements pflag.Value interface, to check that the passed-in
// value is one of the strings in AllowedValues.
type stringEnum struct {
	flag *projectFlag
}

// String returns the current value
func (s *stringEnum) String() string {
	return s.flag.Value
}

// Set checks that the passed-in value is only of the allowd values, and
// returns an error if it is not
func (s *stringEnum) Set(value string) error {
	for _, allowedValue := range s.flag.AllowedValues {
		if allowedValue == value {
			s.flag.Value = value
			return nil
		}
	}
	return fmt.Errorf("allowed values: %s", strings.Join(s.flag.AllowedValues, ", "))
}

// Type is the type of Value. For more info, see:
// https://godoc.org/github.com/spf13/pflag#Values
func (s *stringEnum) Type() string {
	return "string"
}

// addProjectFlagsToCommands adds the user flags to the specified Cobra commands.
func addProjectFlagsToCommands(flags []*projectFlag, cmds ...*cobra.Command) {
	if len(flags) == 0 {
		return
	}

	projectFlagSet := flagsToFlagSet(flags)

	for _, cmd := range cmds {
		for _, flag := range flags {
			flag.AddToFlagSet(cmd.Flags())
		}

		// Update help text for the command to include the user flags
		helpTmpl := cmd.HelpTemplate()
		helpTmpl += "\nUser flags:\n"
		helpTmpl += projectFlagSet.FlagUsages()

		cmd.SetHelpTemplate(helpTmpl)

		// Mark flag hidden so it doesn't appear in the normal help. We have to
		// do this *after* calling flagUsages above, otherwise the flags don't
		// appear in the output.
		for _, flag := range flags {
			cmd.Flags().MarkHidden(flag.Name)
		}
	}
}

// flagsFromConfig reads the astro config and returns a list of projectFlags that
// can be used to fill in astro variable values at runtime.
func flagsFromConfig(config *conf.Project) (flags []*projectFlag) {
	if config == nil {
		return
	}

	flagMap := map[string]*projectFlag{}

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
				flag := &projectFlag{
					Name:        flagName,
					Description: flagConf.Description,
					Variable:    variableConf.Name,
				}
				flag.AllowedValues = make([]string, len(variableConf.Values))
				copy(flag.AllowedValues, variableConf.Values)

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
func flagsToUserVariables(projectFlags []*projectFlag) *astro.UserVariables {
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

// Converts a list of projectFlags to a pflag.flagSet.
func flagsToFlagSet(flags []*projectFlag) *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("projectFlags", pflag.ContinueOnError)
	for _, flag := range flags {
		flag.AddToFlagSet(flagSet)
	}
	return flagSet
}

// flagName returns the flag name, given a variable name.
func (cli *AstroCLI) flagName(variableName string) string {
	if flag, ok := cli.config.Flags[variableName]; ok {
		return flag.Name
	}
	return variableName
}

// varsToFlagNames converts a list of variable names to CLI flags.
func (cli *AstroCLI) varsToFlagNames(variableNames []string) (flagNames []string) {
	for _, v := range variableNames {
		flagNames = append(flagNames, fmt.Sprintf("--%s", cli.flagName(v)))
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
