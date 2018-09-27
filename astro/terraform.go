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
	"github.com/uber/astro/astro/terraform"
)

// newTerraformSession returns a new Terraform session.
func (session *Session) newTerraformSession(execution *boundExecution) (*terraform.Session, error) {
	terraformSessionDir := filepath.Join(session.path, execution.ID())

	moduleConfig := execution.ModuleConfig()

	config := terraform.Config{
		Name:       moduleConfig.Name,
		BasePath:   moduleConfig.TerraformCodeRoot,
		ModulePath: moduleConfig.Path,
		Remote:     moduleConfig.Remote,
		Variables:  execution.Variables(),
	}

	// Fetch the right Terraform version
	terraformVersion := moduleConfig.Terraform.Version

	if terraformVersion != nil {
		terraformPath, err := session.repo.project.terraformVersions.Get(terraformVersion.String())
		if err != nil {
			return nil, fmt.Errorf("unable to activate Terraform %v: %v", terraformVersion.String(), err)
		}

		config.TerraformPath = terraformPath
	}

	// In Terraform 0.9.x and later, the backend configuration must be
	// in the Terraform code itself.
	if terraform.VersionMatches(terraformVersion, ">= 0.9") {
		config.Remote.Backend = ""
	}

	// Create a shared plugin directory
	if terraform.VersionMatches(terraformVersion, ">= 0.10") {
		if _, exists := os.LookupEnv("TF_PLUGIN_CACHE_DIR"); !exists {
			pluginDir := filepath.Join(session.repo.path, "plugins")
			logger.Trace.Printf("astro: creating shared plugin directory: %v", pluginDir)

			if err := os.MkdirAll(pluginDir, 0755); err != nil {
				return nil, err
			}
			config.SharedPluginDir = pluginDir
		}
	}

	// If an override path has been specified, use that instead
	if moduleConfig.Terraform.Path != "" {
		config.TerraformPath = moduleConfig.Terraform.Path
	}

	return terraform.NewTerraformSession(execution.ID(), terraformSessionDir, config)
}
