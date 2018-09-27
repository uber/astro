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
	"errors"
	"fmt"
	"path/filepath"

	"github.com/uber/astro/astro/utils"

	multierror "github.com/hashicorp/go-multierror"
)

// Module is the static configuration of a Terraform module.
type Module struct {
	// Deps is a list of Terraform modules that need to be run before this one
	// can run.
	Deps []Dependency
	// Hooks contains the module-specific hooks that can run.
	Hooks ModuleHooks
	// Name is a unique name for this Terraform module.
	Name string
	// Path is the path to the module, relative to the code root.
	Path string
	// Remote is the Terraform remote for this module.
	Remote Remote
	// TerraformCodeRoot is the base path to the Terraform code. Users cannot
	// set this; instead they should set it on the project configuration.
	TerraformCodeRoot string `json:"-"`
	// Terraform stores Terraform configuration that should be used when
	// running this module.
	Terraform Terraform
	// Variables is a list of Terraform variables and possible values that this
	// module accepts.
	Variables []Variable
}

// Validate validates whether the configuration is good. Returns any validation
// errors.
func (m *Module) Validate() (errs error) {
	if m.Path == "" {
		errs = multierror.Append(errs, errors.New("path cannot be empty"))
	} else {
		fullModulePath := filepath.Join(m.TerraformCodeRoot, m.Path)

		if !utils.IsWithinPath(m.TerraformCodeRoot, fullModulePath) {
			errs = multierror.Append(errs, fmt.Errorf("module path cannot be outside code root: module path: %v; code root: %v", fullModulePath, m.TerraformCodeRoot))
		}

		if !utils.IsDirectory(fullModulePath) {
			errs = multierror.Append(errs, fmt.Errorf("module directory does not exist: %v", fullModulePath))
		}
	}
	if err := m.Terraform.Validate(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("Terraform: %v", err))
	}
	for _, hook := range m.Hooks.PreModuleRun {
		if err := hook.Validate(); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("PreModuleRun Hook: %v", err))
		}
	}

	return errs
}
