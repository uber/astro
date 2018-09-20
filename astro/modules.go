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

package astro

import (
	"fmt"
	"strings"

	"github.com/uber/astro/astro/conf"
)

// module represents a Terraform module.
type module struct {
	config *conf.Module
}

// NewModule creates a new module instance.
func newModule(config conf.Module) *module {
	return &module{config: &config}
}

// Executions returns a list of all possible Executions based
// on the variable names/values.
func (m *module) executions(parameters ExecutionParameters) executionSet {
	filterCount := 0
	for _, variable := range m.config.Variables {
		if parameters.UserVars.HasFilter(variable.Name) {
			filterCount++
		}
	}
	if filterCount != parameters.UserVars.FilterCount() {
		return executionSet{}
	}

	// If a module doesn't have any variables, then there's just a
	// single execution.
	if len(m.config.Variables) < 1 {
		return executionSet{
			&unboundExecution{
				&execution{
					moduleConf:          m.config,
					terraformParameters: parameters.TerraformParameters,
				},
			},
		}
	}

	var variableValues [][]interface{}

	for _, variable := range m.config.Variables {
		v := []interface{}{}
		filtered := variable.IsFilter() && parameters.UserVars.Values[variable.Name] != ""

		if variable.Values != nil {
			for _, value := range variable.Values {
				if !filtered || value == parameters.UserVars.Values[variable.Name] {
					v = append(v, fmt.Sprintf("%s=%s", variable.Name, value))
				}
			}
		} else {
			// If there are no predefined variable values, we create a single
			// value "{var_name}" as a placeholder
			v = append(v, fmt.Sprintf("%s={%s}", variable.Name, variable.Name))
		}

		variableValues = append(variableValues, v)
	}

	executions := executionSet{}

	products := cartesian(variableValues...)

	for _, p := range products {
		e := &unboundExecution{
			&execution{
				moduleConf:          m.config,
				terraformParameters: parameters.TerraformParameters,
			},
		}

		e.variables = make(map[string]string)
		for _, value := range p {
			s := strings.Split(value.(string), "=")
			e.variables[s[0]] = s[1]
		}

		executions = append(executions, e)
	}

	return executions
}
