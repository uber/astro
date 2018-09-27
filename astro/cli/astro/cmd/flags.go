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

	"github.com/uber/astro/astro/conf"
)

// Flag is populated with the custom variables read from command line flags
type Flag struct {
	// Variable is the name of the Terraform variable set by the flag
	Variable string
	// Value is the value read from the command line, if present
	Value string
	// Flag is the name of the command line flag used to set the variable
	Flag string
	// IsRequired is true when thie flag is mandatory
	IsRequired bool
	// IsFilter is true when thie flag acts as a filter for the list of modules
	IsFilter bool
	// AllowedValues is the list of valid values for this flag
	AllowedValues []string
}

// StringEnum implements pflag.Value interface, to check that the passed-in value is one of the strings in AllowedValues
type StringEnum struct {
	Flag *Flag
}

// String returns the current value
func (s *StringEnum) String() string {
	return s.Flag.Value
}

// Set checks that the passed-in value is only of the allowd values, and returns an error if it is not
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

// commandLineFlags returns a list of variables that can be set via command line
func commandLineFlags(conf *conf.Project) ([]*Flag, error) {
	var err error
	flags := make([]*Flag, 0)

	for _, moduleConf := range conf.Modules {
		for _, variableConf := range moduleConf.Variables {
			if flags, err = addOrUpdateVariable(flags, variableConf); err != nil {
				return nil, err
			}
		}
	}

	for i := range flags {
		flags[i].AllowedValues = uniqueStrings(flags[i].AllowedValues)
	}

	return flags, nil
}

func addOrUpdateVariable(flags []*Flag, variable conf.Variable) ([]*Flag, error) {
	commandFlag := variable.CommandFlag()
	found := false
	for i := range flags {
		if flags[i].Variable == variable.Name {
			if err := checkVariableConsistency(flags[i], variable); err != nil {
				return nil, err
			}
			flags[i].AllowedValues = append(flags[i].AllowedValues, variable.Values...)
			found = true
		}
		if flags[i].Flag == commandFlag {
			if flags[i].Variable != variable.Name {
				return nil, fmt.Errorf("Flag '%s' is mapped to conflicting flags '%s' and '%s'", commandFlag, variable.Name, flags[i].Variable)
			}
		}
	}
	if !found {
		flags = append(flags, &Flag{
			Variable:      variable.Name,
			Flag:          commandFlag,
			IsRequired:    variable.IsRequired(),
			IsFilter:      variable.IsFilter(),
			AllowedValues: variable.Values,
		})
	}
	return flags, nil
}

func checkVariableConsistency(flag *Flag, variable conf.Variable) error {
	commandFlag := variable.CommandFlag()
	if flag.Flag != commandFlag {
		return fmt.Errorf("variable '%s' is mapped to conflicting flags '%s' and '%s'", variable.Name, commandFlag, flag.Flag)
	}
	if flag.IsRequired != variable.IsRequired() {
		return fmt.Errorf("variable '%s' is inconsistently marked as required", variable.Name)
	}
	if flag.IsFilter != variable.IsFilter() {
		return fmt.Errorf("variable '%s' is inconsistently used as filter", variable.Name)
	}
	return nil
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
