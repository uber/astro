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

package terraform

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/uber/astro/astro/logger"
	"github.com/uber/astro/astro/utils"
)

func (s *Session) terraformInitArgsLegacy() ([]string, error) {
	args := []string{"remote", "config"}

	if s.config.Remote.Backend != "" {
		args = append(args, "-backend", s.config.Remote.Backend)
	}

	for key, val := range s.config.Remote.BackendConfig {
		args = append(args, fmt.Sprintf("-backend-config=%s=%s", key, val))
	}

	return args, nil
}

func (s *Session) terraformInitArgsModern() ([]string, error) {
	args := []string{"init"}

	if s.config.Remote.Backend != "" {
		return nil, errors.New("backend configuration was specified but is not compatible with Terraform 0.9.x and later")
	}

	// Backend config parameters are permitted, however
	for key, val := range s.config.Remote.BackendConfig {
		args = append(args, fmt.Sprintf("-backend-config=%s=%s", key, val))
	}

	// Input is a new option that means Terraform will return an
	// error in cases where it will normally ask for input (and
	// hang).
	args = append(args, "-input=false")

	return args, nil
}

// Init initializes a Terraform module. This needs to happen before other
// commands like "plan" and "apply" can be called. See:
// https://www.terraform.io/docs/commands/init.html
func (s *Session) Init() (Result, error) {
	logger.Trace.Printf("terraform: initializing module in directory: %v\n", s.moduleDir)

	terraformVersion, err := s.versionCached()
	if err != nil {
		return nil, err
	}

	// If we're on 0.8.x and lower and there is no backend config, we
	// can skip straight to the `terraform get`. No init required.
	if VersionMatches(terraformVersion, "< 0.9") && s.config.Remote.Backend == "" {
		return s.Get()
	}

	var args []string

	if VersionMatches(terraformVersion, "< 0.9") {
		args, err = s.terraformInitArgsLegacy()
		if err != nil {
			return nil, err
		}
	} else {
		args, err = s.terraformInitArgsModern()
		if err != nil {
			return nil, err
		}
	}

	process, err := s.terraformCommand(args, []int{0})
	if err != nil {
		return nil, err
	}

	if err := process.Run(); err != nil {
		logger.Trace.Printf("terraform: init failed: %v\n", err)
		return &terraformResult{
			process: process,
		}, err
	}

	return s.Get()
}

// Initialized returns whether or not `terraform init` has been run.
func (s *Session) Initialized() bool {
	terraformSpecialDir := filepath.Join(s.moduleDir, ".terraform")
	return utils.IsDirectory(terraformSpecialDir)
}
