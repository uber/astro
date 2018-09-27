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
	"os"
	"path/filepath"

	"github.com/uber/astro/astro/logger"
	"github.com/uber/astro/astro/utils"

	"github.com/hashicorp/terraform/dag"
)

// SessionRepo is a parent directory that contains inidividual project
// sessions.
type SessionRepo struct {
	project *Project

	path       string
	generateID func() string

	current *Session
}

// NewSessionRepo creates or opens a project session repo.
func NewSessionRepo(project *Project, repoPath string, idGenFunc func() string) (*SessionRepo, error) {
	// Create session directory if it doesn't exist
	if !utils.IsDirectory(repoPath) {
		if err := os.Mkdir(repoPath, 0755); err != nil {
			return nil, err
		}
	}

	return &SessionRepo{
		project:    project,
		path:       repoPath,
		generateID: idGenFunc,
	}, nil
}

// Session is a directory containing log output, Terraform state files
// and plans.
type Session struct {
	repo *SessionRepo

	id   string
	path string
}

// NewSession creates a new session in the repository.
func (r *SessionRepo) NewSession() (*Session, error) {
	id := r.generateID()

	sessionPath := filepath.Join(r.path, id)
	if err := os.Mkdir(sessionPath, 0755); err != nil {
		return nil, err
	}

	return &Session{
		id:   id,
		path: sessionPath,
		repo: r,
	}, nil
}

// Current returns the last session created, or creates one if it's the
// first time it's called.
func (r *SessionRepo) Current() (*Session, error) {
	if r.current != nil {
		return r.current, nil
	}

	session, err := r.NewSession()
	if err != nil {
		return nil, err
	}

	r.current = session

	return session, nil
}

func (s *Session) apply(boundExecutions []*boundExecution) (<-chan string, <-chan *Result, error) {
	logger.Trace.Println("astro session: running apply without graph")

	numberOfExecutions := len(boundExecutions)
	// Needs to be big enough to buffer log lines from below for tests that
	// don't consume from the channel.
	status := make(chan string, numberOfExecutions*10)
	results := make(chan *Result, numberOfExecutions)

	logger.Trace.Printf("astro: %d executions to apply\n", numberOfExecutions)

	fns := []func(){}
	for _, e := range boundExecutions {
		b := e // save for use inside the loop
		fns = append(fns, func() {
			terraform, err := s.newTerraformSession(b)
			if err != nil {
				results <- &Result{
					id:  b.ID(),
					err: err,
				}
				return
			}

			status <- fmt.Sprintf("[%s] Initializing...", b.ID())
			if result, err := terraform.Init(); err != nil {
				results <- &Result{
					id:              b.ID(),
					terraformResult: result,
					err:             err,
				}
				return
			}

			status <- fmt.Sprintf("[%s] Applying...", b.ID())
			result, err := terraform.Apply()
			results <- &Result{
				id:              b.ID(),
				terraformResult: result,
				err:             err,
			}
		})
	}

	go func() {
		defer close(results) // signals the end of all executions
		utils.Parallel(10, fns...)
	}()

	return status, results, nil
}

func (s *Session) applyWithGraph(boundExecutions []*boundExecution) (<-chan string, <-chan *Result, error) {
	logger.Trace.Println("astro session: running apply with graph")

	// Convert unboundExecutions to executionSet
	executions := make(executionSet, len(boundExecutions))
	for i, e := range boundExecutions {
		executions[i] = e
	}

	// Generate dep graph
	graph, err := executions.graph()
	if err != nil {
		return nil, nil, err
	}

	numberOfExecutions := len(executions)
	// Needs to be big enough to buffer log lines from below for tests that
	// don't consume from the channel.
	status := make(chan string, numberOfExecutions*10)
	results := make(chan *Result, numberOfExecutions)

	// Walk the graph and execute
	go func() {
		defer close(results)

		graph.Walk(func(vertex dag.Vertex) error {
			// skip if we've reached the root
			if _, ok := vertex.(graphNodeRoot); ok {
				return nil
			}

			b := vertex.(*boundExecution)

			terraform, err := s.newTerraformSession(b)
			if err != nil {
				results <- &Result{
					id:  b.ID(),
					err: err,
				}
				return err
			}

			for _, hook := range b.ModuleConfig().Hooks.PreModuleRun {
				status <- fmt.Sprintf("[%s] Running PreModuleRun hook...", b.ID())
				if err := runCommandkAndSetEnvironment(s.path, hook); err != nil {
					results <- &Result{
						id:  b.ID(),
						err: fmt.Errorf("error running PreModuleRun hook: %v", err),
					}
					return err
				}
			}

			status <- fmt.Sprintf("[%s] Initializing...", b.ID())
			if result, err := terraform.Init(); err != nil {
				results <- &Result{
					id:              b.ID(),
					terraformResult: result,
					err:             err,
				}
				return err
			}

			status <- fmt.Sprintf("[%s] Applying...", b.ID())

			result, err := terraform.Apply()
			results <- &Result{
				id:              b.ID(),
				terraformResult: result,
				err:             err,
			}

			// This will cause any executions that depend on this one
			// to be skipped.
			return err
		})
	}()

	return status, results, nil
}

func (s *Session) plan(boundExecutions []*boundExecution, detach bool) (<-chan string, <-chan *Result, error) {
	logger.Trace.Println("astro session: running plan")

	numberOfExecutions := len(boundExecutions)
	// Needs to be big enough to buffer log lines from below for tests that
	// don't consume from the channel.
	status := make(chan string, numberOfExecutions*10)
	results := make(chan *Result, numberOfExecutions)

	logger.Trace.Printf("astro: %d executions to plan\n", numberOfExecutions)

	// Create plan functions
	fns := []func(){}
	for _, e := range boundExecutions {
		b := e // save for use inside the loop
		fns = append(fns, func() {
			terraform, err := s.newTerraformSession(b)
			if err != nil {
				results <- &Result{
					id:  b.ID(),
					err: err,
				}
				return
			}

			for _, hook := range e.ModuleConfig().Hooks.PreModuleRun {
				status <- fmt.Sprintf("[%s] Running PreModuleRun hook...", b.ID())
				if err := runCommandkAndSetEnvironment(s.path, hook); err != nil {
					results <- &Result{
						id:  b.ID(),
						err: fmt.Errorf("error running PreModuleRun hook: %v", err),
					}
					return
				}
			}

			status <- fmt.Sprintf("[%s] Initializing...", b.ID())
			if result, err := terraform.Init(); err != nil {
				results <- &Result{
					id:              b.ID(),
					terraformResult: result,
					err:             err,
				}
				return
			}

			if detach {
				status <- fmt.Sprintf("[%s] Disconnecting remote state...", b.ID())
				if result, err := terraform.Detach(); err != nil {
					results <- &Result{
						id:              b.ID(),
						terraformResult: result,
						err:             err,
					}
					return
				}
			}

			status <- fmt.Sprintf("[%s] Planning...", b.ID())
			result, err := terraform.Plan()
			results <- &Result{
				id:              b.ID(),
				terraformResult: result,
				err:             err,
			}
		})
	}

	// Run plans in parallel
	go func() {
		defer close(results) // signals the end of all executions
		utils.Parallel(10, fns...)
	}()

	return status, results, nil
}
