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
	"os/exec"

	"github.com/uber/astro/astro/logger"
	"github.com/uber/astro/astro/tvm"

	version "github.com/burl/go-version"
	multierror "github.com/hashicorp/go-multierror"
)

// Terraform configuration that affect the running of Terraform itself.
type Terraform struct {
	// Path is the path to the Terraform binary. Colonist will use this
	// if set, otherwise it will automatically download the version in
	// Version below.
	Path string
	// Terraform version to use. If Path is empty, Colonist will
	// download this version automatically.
	Version *version.Version
}

// ApplyDefaultsFrom takes a Terraform struct representation the default
// configuration and fills in any fields that were not set.
func (conf *Terraform) ApplyDefaultsFrom(defaultConf Terraform) {
	if conf.Path == "" {
		conf.Path = defaultConf.Path
	}
	if conf.Version == nil {
		conf.Version = defaultConf.Version
	}
}

// SetDefaultPath sets the path the Terraform binary from the environment, if
// it hasn't already been provided in configuration.
func (conf *Terraform) SetDefaultPath() error {
	// If the existing project config doesn't specify a Terraform path,
	// search for it in the current environment.
	terraformPath, err := exec.LookPath("terraform")
	if err != nil {
		return err
	}

	logger.Trace.Printf("conf/terraform: setting Terraform path to: %v", terraformPath)
	conf.Path = terraformPath

	return nil
}

// SetVersionFromBinary sets the value of the Version field from the binary.
func (conf *Terraform) SetVersionFromBinary() error {
	version, err := tvm.InspectVersion(conf.Path)
	if err != nil {
		return fmt.Errorf("unable to detect Terraform version: %v", err)
	}

	logger.Trace.Printf("conf/terraform: set Terraform version to: %v", version)
	conf.Version = version
	return nil
}

// Validate checks the Terraform configuration is good.
func (conf *Terraform) Validate() (errs error) {
	// Version must be set by the time astro runs; however, in the config it
	// can be left blank and astro will detect and autofill the version from
	// the Terraform in the user's environment.
	if conf.Version == nil {
		errs = multierror.Append(errs, errors.New("Version is not set"))
	}
	return errs
}
