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
	"regexp"
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
		args = append(args, "-var", fmt.Sprintf("%s=%s", key, val))
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
		terraformVersion, err := s.versionCached()
		if err != nil {
			return nil, err
		}
		if VersionMatches(terraformVersion, "<0.12") {
			result, err := s.Show(fmt.Sprintf("%s.plan", s.id))
			if err != nil {
				return result, err
			}
			changes = result.Stdout()
		} else {
			rawPlanOutput := process.Stdout().String()
			var re = regexp.MustCompile(`(?s)Terraform will perform the following actions:(.*)-{72}`)
			if match := re.FindStringSubmatch(rawPlanOutput); len(match) == 2 {
				changes = match[1]
			} else {
				return &terraformResult{
					process: process,
				}, fmt.Errorf("unable to parse terraform plan output")
			}
		}
	}

	return &PlanResult{
		terraformResult: &terraformResult{
			process: process,
		},
		changes: changes,
	}, nil
}
