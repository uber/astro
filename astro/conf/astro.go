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

package conf

import (
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
)

// Project represents the structure of the YAML configuration for astro.
type Project struct {
	// Flags is a mapping of module variable names to user flags, e.g. for on
	// the CLI.
	Flags map[string]Flag

	// Hooks contains configuration of hooks that can be invoked at various
	// stages of the CLI lifecycle.
	Hooks Hooks

	// Modules is a list of Terraform modules.
	Modules []Module

	// SessionRepoDir is the path to the directory where astro
	// will create the .astro session repo that stores log files and
	// plans during a session. Defaults to the same directory as the config
	// file.
	SessionRepoDir string `json:"session_repo_dir"`

	// TerraformCodeRoot is the path to the root of the Terraform code for this
	// Project. Defaults to the same directory as the config file.
	TerraformCodeRoot string `json:"terraform_code_root"`

	// Default Terraform configuration for this project. This
	// configuration is used when executing Terraform. Modules can
	// override this configuration with their own.
	TerraformDefaults Terraform `json:"terraform"`
}

// Validate checks the project configuration is good.
func (conf *Project) Validate() (errs error) {
	if err := conf.TerraformDefaults.Validate(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("TerraformDefaults: %v", err))
	}
	for _, moduleConf := range conf.Modules {
		if err := moduleConf.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("Module[%v]: %v", moduleConf.Name, err))
		}
	}
	for _, hook := range conf.Hooks.Startup {
		if err := hook.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("Startup Hook: %v", err))
		}
	}
	for _, hook := range conf.Hooks.PreModuleRun {
		if err := hook.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("PreModuleRun Hook: %v", err))
		}
	}
	return errs
}
