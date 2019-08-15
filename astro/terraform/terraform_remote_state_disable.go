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
	"path/filepath"
	"strings"

	"github.com/uber/astro/astro/logger"
	"github.com/uber/astro/astro/utils"
)

// Detach disables any connection to the remote state for the given module. If
// the module is not initialized, it does that first, and then disconnects it.
// This is so that Terraform first downloads the remote state locally.
// The purpose of Detach is to allow safe, local testing of changes to the
// state file, without pushing anything to the remote.
func (s *Session) Detach() (Result, error) {
	logger.Trace.Printf("terraform: detaching remote state in %v", s.moduleDir)

	var res Result
	var err error

	if !s.Initialized() {
		if res, err := s.Init(); err != nil {
			return res, err
		}
	}

	terraformVersion, err := s.versionCached()
	if err != nil {
		return nil, err
	}

	if VersionMatches(terraformVersion, "<0.9") {
		res, err = s.detachLegacy()
	} else {
		res, err = s.detachModern()
	}

	if err != nil {
		return res, err
	}

	// failsafe to make sure the remote file was copied locally
	if !utils.FileExists(filepath.Join(s.moduleDir, "terraform.tfstate")) {
		return nil, errors.New("detach failed: terraform.tfstate does not exist")
	}

	return res, nil
}

func (s *Session) detachLegacy() (Result, error) {
	detachCmd, err := s.terraformCommand([]string{"remote", "config", "-disable"}, []int{0})
	if err != nil {
		return nil, err
	}

	res := &terraformResult{
		process: detachCmd,
	}

	if err := detachCmd.Run(); err != nil {
		return res, err
	}

	return res, nil
}

func (s *Session) detachModern() (Result, error) {
	if err := s.deleteBackendConfig(); err != nil {
		return nil, err
	}

	reinit, err := s.terraformCommand([]string{"init", "-force-copy"}, []int{0})
	if err != nil {
		return nil, err
	}

	res := &terraformResult{
		process: reinit,
	}

	if err := reinit.Run(); err != nil {
		return res, err
	}

	return res, nil
}

// deleteBackendConfig deletes the Terraform backend configuration from
// the .tf files in this module session.
func (s *Session) deleteBackendConfig() error {
	grep, err := s.command("grep", "grep", []string{"-rlE", "terraform\\s+{", s.moduleDir}, []int{0, 1})
	if err != nil {
		return err
	}

	if err := grep.Run(); err != nil {
		return err
	}

	candidates := strings.Split(strings.TrimSpace(grep.Stdout().String()), "\n")

	if len(candidates) < 1 {
		return errors.New("cannot find backend configuration in the Terraform files")
	}
	terraformVersion, err := s.Version()
	if err != nil {
		return err
	}

	for _, f := range candidates {
		if err := deleteTerraformBackendConfigFromFile(f, terraformVersion); err != nil {
			return err
		}
	}
	return nil
}
