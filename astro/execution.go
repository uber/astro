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
	"sort"
	"strings"

	"github.com/uber/astro/astro/conf"
)

// MissingRequiredVarsError is an error type that is returned from plan or
// apply when there are variables that need to be provided at run time that are
// missing.
type MissingRequiredVarsError struct {
	missing []string
}

func (e *MissingRequiredVarsError) plural() string {
	if len(e.missing) > 0 {
		return "s"
	}
	return ""
}

// Error is the error message, so this satisfies the error interface.
func (e MissingRequiredVarsError) Error() string {
	return fmt.Sprintf("missing required variable%s: %s", e.plural(), strings.Join(e.missing, ", "))
}

// MissingVars returns a list of the missing user variables.
func (e MissingRequiredVarsError) MissingVars() []string {
	return e.missing
}

// terraformExecution is an interface that covers both bound and unbound
// executions.
type terraformExecution interface {
	ID() string
	ModuleConfig() conf.Module
	Variables() map[string]string
	TerraformParameters() []string
}

// execution represents the execution of a module with some variable
// values. This type is never used directly: instead, the types
// unboundExecution and boundExecution are used.
type execution struct {
	moduleConf *conf.Module

	// variables is a map of variables to be passed during the
	// execution.
	variables map[string]string

	// terraformParameters is a list of additional Terraform parameters for this execution
	terraformParameters []string
}

// Name is an alias for ID; so that terraform/dag trace output makes
// sense
func (e *execution) Name() string {
	return e.ID()
}

// ID returns a unique ID for this execution.
func (e *execution) ID() string {
	// For boundExecutions, the ID should be:
	// {modulename}-{variableValue1}-{variableValue2}-{and so on...}
	// Where variableValues are the values of the runtime variables.

	values := []string{}

	// Since runtime variables may have values that don't directly
	// pertain to this module/execution, we need to extract only the
	// variable names that are relevant to this module.
	keys := []string{}
	for _, v := range e.ModuleConfig().Variables {
		keys = append(keys, v.Name)
	}

	sort.Strings(keys)

	for _, key := range keys {
		values = append(values, e.variables[key])
	}

	// construct the ID
	id := e.ModuleConfig().Name
	if len(values) > 0 {
		id = fmt.Sprintf("%s-%s", id, strings.Join(values, "-"))
	}

	return id
}

// ModuleConfig returns a copy of the configuration of the module
// associated with this execution.
func (e *execution) ModuleConfig() conf.Module {
	return *e.moduleConf
}

// Variables returns a reference to the variables set for this execution
func (e *execution) Variables() map[string]string {
	return e.variables
}

// TerraformParameters returns reference to the Terraform parameters set for this execution
func (e *execution) TerraformParameters() []string {
	return e.terraformParameters
}

// unboundExecution represents a module execution before runtime
// variables have been provided by the user and template strings
// replaced in the variable values.
// An unboundExecution should never be actually executed. Instead,
// bind() should be called with user variables supplied first.
type unboundExecution struct {
	*execution
}

// bind takes a map of user-specified variables and returns a
// boundExecution with variable values replaced. An error is returned if
// not all required user values were provided.
func (e *unboundExecution) bind(userVars map[string]string) (*boundExecution, error) {
	boundVars := union(e.Variables(), userVars)

	missingVars := []string{}

	// Check that the user provided variables replace everything that
	// needs to be replaced.
	for _, val := range boundVars {
		if err := assertAllVarsReplaced(val); err != nil {
			missingVars = append(missingVars, extractMissingVarNames(val)...)
		}
	}

	if len(missingVars) > 0 {
		return nil, MissingRequiredVarsError{missing: missingVars}
	}

	// Create a copy of the config and search attributes for placeholders
	// to replace with values from the bound vars.
	boundConfig := e.ModuleConfig()

	// TODO: Loop over all module configuration using reflection

	boundBackendConfig, err := replaceAllVarsInMapValues(boundConfig.Remote.BackendConfig, boundVars)
	if err != nil {
		return nil, fmt.Errorf("unable to bind execution: %v; %v", e.ID(), err)
	}
	boundConfig.Remote.BackendConfig = boundBackendConfig

	return &boundExecution{
		&execution{
			moduleConf:          &boundConfig,
			variables:           boundVars,
			terraformParameters: e.TerraformParameters(),
		},
	}, nil
}

// boundExecution represents a module execution that is ready to be
// executed.
type boundExecution struct {
	*execution
}
