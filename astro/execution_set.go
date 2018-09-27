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

	"github.com/uber/astro/astro/conf"

	"github.com/hashicorp/terraform/dag"
)

// executionSet is a set of executions that can depend on each other.
type executionSet []terraformExecution

// bindAll takes a set of unboundExecutions and returns a new set with
// all executions bound to userVars. An error is thrown if any of the
// executions in the current set are already bound.
func (s executionSet) bindAll(userVars map[string]string) ([]*boundExecution, error) {
	results := []*boundExecution{}
	for _, e := range s {
		unbound, ok := e.(*unboundExecution)
		if !ok {
			return nil, fmt.Errorf("cannot bind executions: %v not of type unboundExecution", e)
		}

		bound, err := unbound.bind(userVars)
		if err != nil {
			return nil, err
		}

		results = append(results, bound)
	}

	return results, nil
}

// filterByModule returns all the executions in this set that match
// moduleName.
func (s executionSet) filterByModule(moduleName string) (results executionSet) {
	for _, e := range s {
		if e.ModuleConfig().Name == moduleName {
			results = append(results, e)
		}
	}
	return results
}

// filterByDep returns all executions in this set that match the
// dependencies expressed in dep.
func (s executionSet) filterByDep(dep conf.Dependency) (executionSet, error) {
	var dependentExecutions executionSet

	executionsForModule := s.filterByModule(dep.Module)
	if len(executionsForModule) == 0 {
		return nil, fmt.Errorf("missing dependency: %v", dep.Module)
	}

	// If the dependency expression does not specific specific variables, then
	// assume it depends on any and all executions of this module.
	if dep.Variables == nil {
		return executionsForModule, nil
	}

	// Try to match the dependency to a specific execution.
	for _, e := range executionsForModule {
		if filterMaps(dep.Variables, e.Variables()) {
			dependentExecutions = append(dependentExecutions, e)
		}
	}

	// If there are no executions matching the dependency, it means the
	// configuration is wrong.
	if len(dependentExecutions) == 0 {
		return nil, fmt.Errorf("no execution matching dep: %v", dep)
	}

	return dependentExecutions, nil
}

// graph returns an acyclic graph of executions in this set.
func (s executionSet) graph() (*dag.AcyclicGraph, error) {
	graph := &dag.AcyclicGraph{}

	// Add all executions to the graph to start off with
	for _, e := range s {
		graph.Add(e)
	}

	// For each execution, we need to find the dependencies and connect
	// them in the graph.
	for _, e := range s {
		for _, dep := range e.ModuleConfig().Deps {
			// Fill in any placeholders in the dependency with variable
			// values from the current execution.
			vars, err := replaceVarsInMapValues(dep.Variables, e.Variables())
			if err != nil {
				return nil, fmt.Errorf("unable to resolve vars for module: %s; %v", e.ModuleConfig().Name, err)
			}
			dep.Variables = vars

			dependentExecutions, err := s.filterByDep(dep)
			if err != nil {
				return nil, fmt.Errorf("invalid dependency for %s: %v", e.ModuleConfig().Name, err)
			}
			for _, dependentExecution := range dependentExecutions {
				graph.Connect(dag.BasicEdge(e, dependentExecution))
			}
		}
	}

	addRoot(graph)

	return graph, nil
}
