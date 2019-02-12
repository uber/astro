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
	"fmt"

	"github.com/uber/astro/astro/logger"
)

// Plan runs a `terraform plan`
func (s *Session) Plan() (Result, error) {
	if !s.Initialized() {
		if result, err := s.Init(); err != nil {
			return result, err
		}
	}

	args := []string{"plan", "-detailed-exitcode", fmt.Sprintf("-out=%s.plan", s.id)}

	for key, val := range s.config.Variables {
		if key != "workspace" {
			args = append(args, "-var", fmt.Sprintf("%s=%s", key, val))
		} else if key == "workspace" {
			logger.Trace.Println("checking out workspace: %s", val)
			process, err := s.terraformCommand([]string{"workspace", "select", val}, []int{0})

			if err != nil {
				return nil, err
			}

			if err := process.Run(); err != nil {
				return &terraformResult{
					process: process,
				}, err
			}
		}
	}

	args = append(args, s.config.TerraformParameters...)

	process, err := s.terraformCommand(args, []int{0, 2})
	if err != nil {
		return nil, err
	}

	if err := process.Run(); err != nil {
		return &terraformResult{
			process: process,
		}, err
	}

	var changes string

	// With -detailed-exitcode, plans that return exit code 2 mean there
	// are changes (so there's no error).
	if process.ExitCode() == 2 {
		// Fetch changes
		result, err := s.Show(fmt.Sprintf("%s.plan", s.id))
		if err != nil {
			return result, err
		}

		changes = result.Stdout()
	}

	return &PlanResult{
		terraformResult: &terraformResult{
			process: process,
		},
		changes: changes,
	}, nil
}
