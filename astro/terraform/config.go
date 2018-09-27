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

	"github.com/uber/astro/astro/conf"
	multierror "github.com/hashicorp/go-multierror"
)

// Config is the Terraform configuration required to initialize and run
// Terraform commands for a module.
type Config struct {
	// Name is a unique name for this Terraform module.
	Name string
	// BasePath is the base path to the Terraform code.
	BasePath string
	// ModulePath is the path to the module, relative to the basepath.
	ModulePath string
	// Remote is the Terraform remote configuration for this module.
	Remote conf.Remote
	// Variables is a map of the variable values for execution.
	Variables map[string]string

	// TerraformPath is the path to the Terraform binary
	TerraformPath string

	// SharedPluginDir is the path to a directory that should contain shared
	// plugins.
	SharedPluginDir string
}

// Validate validates the Terraform configuration is valid.
func (config Config) Validate() (errs error) {
	if config.BasePath == "" {
		errs = multierror.Append(errs, errors.New("base path cannot be empty"))
	}
	if config.ModulePath == "" {
		errs = multierror.Append(errs, errors.New("module path cannot be empty"))
	}
	if config.TerraformPath == "" {
		errs = multierror.Append(errs, errors.New("terraform path cannot be empty"))
	}
	return errs
}
