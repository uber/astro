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
	"path/filepath"

	"github.com/uber/astro/astro/conf"
	"github.com/uber/astro/astro/logger"
	"github.com/uber/astro/astro/tvm"
	"github.com/uber/astro/astro/utils"
)

// Project is a collection of Terraform modules, based on configuration.
//
// Modules may be invoked with various parameters, which are either
// provided by the user at runtime, or predefined in configuration.
//
// The combination of a module, along with a map of variable values, is
// called an "execution".
//
// Executions can have dependencies between each other (again, defined
// in the configuration). Based on dependencies, all modules can be
// planned or applied concurrently.
//
type Project struct {
	config            *conf.Project
	sessions          *SessionRepo
	terraformVersions *tvm.VersionRepo
}

// NewProject returns a new instance of Project.
func NewProject(config conf.Project) (*Project, error) {
	logger.Trace.Println("astro: initializing")

	project := &Project{}

	versionRepo, err := tvm.NewVersionRepoForCurrentSystem("")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tvm: %v", err)
	}

	sessionRepoPath := filepath.Join(config.SessionRepoDir, ".astro")
	sessions, err := NewSessionRepo(project, sessionRepoPath, utils.ULIDString)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize session repository: %v", err)
	}

	project.config = &config
	project.sessions = sessions
	project.terraformVersions = versionRepo

	// validate config
	if errs := project.config.Validate(); errs != nil {
		return nil, errs
	}

	// check dependency graph is all good
	if _, err := project.executions(nil, NoUserVariables()).graph(); err != nil {
		return nil, err
	}

	if config.Hooks.Startup == nil {
		return project, nil
	}

	session, err := project.sessions.Current()
	if err != nil {
		return nil, err
	}
	for _, hook := range config.Hooks.Startup {
		if err := runCommandkAndSetEnvironment(session.path, hook); err != nil {
			return nil, fmt.Errorf("error running Startup hook: %v", err)
		}
	}

	return project, nil
}

// executions returns a set of executions for modules registered in this
// project.
func (c *Project) executions(moduleNames []string, userVars *UserVariables) executionSet {
	results := executionSet{}
	for _, m := range c.modules(moduleNames) {
		results = append(results, m.executions(userVars)...)
	}
	return results
}

// modules creates a list of modules based on the config.
func (c *Project) modules(moduleNames []string) []*module {
	results := []*module{}
	for _, moduleConfig := range c.config.Modules {
		// skip, if we're filtering and this module doesn't match the filter
		if moduleNames != nil && !utils.StringSliceContains(moduleNames, moduleConfig.Name) {
			logger.Trace.Printf("astro: ignoring module %v as it does not match filter", moduleConfig.Name)
			continue
		}
		results = append(results, newModule(moduleConfig))
	}
	return results
}

// Plan does a Terraform plan for every possible execution, in
// parallel, ignoring dependencies.
func (c *Project) Plan(moduleNames []string, userVars *UserVariables, detach bool) (<-chan string, <-chan *Result, error) {
	logger.Trace.Println("astro: running Plan")

	// Binds user vars
	boundExecutions, err := c.executions(moduleNames, userVars).bindAll(userVars.Values)
	if err != nil {
		return nil, nil, err
	}

	// Get session
	session, err := c.sessions.Current()
	if err != nil {
		return nil, nil, err
	}

	return session.plan(boundExecutions, detach)
}

// Apply does a Terraform apply for every possible execution,
// in parallel, taking into consideration dependencies. It returns an
// error if it is unable to start, e.g. due to a missing required
// variable.
func (c *Project) Apply(moduleNames []string, userVars *UserVariables) (<-chan string, <-chan *Result, error) {
	logger.Trace.Println("astro: running Apply")

	// Bind user vars
	boundExecutions, err := c.executions(moduleNames, userVars).bindAll(userVars.Values)
	if err != nil {
		return nil, nil, err
	}

	// Get session
	session, err := c.sessions.Current()
	if err != nil {
		return nil, nil, err
	}

	var applyFn func([]*boundExecution) (<-chan string, <-chan *Result, error)
	if moduleNames != nil {
		applyFn = session.apply
	} else {
		applyFn = session.applyWithGraph
	}

	return applyFn(boundExecutions)
}
