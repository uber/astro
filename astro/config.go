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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/uber/astro/astro/conf"
	"github.com/uber/astro/astro/logger"

	"github.com/ghodss/yaml"
)

// NewConfigFromFile parses the configuration in the specified config file
func NewConfigFromFile(configFilePath string) (*conf.Project, error) {
	yamlBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	config, err := configFromYAML(yamlBytes, filepath.Dir(configFilePath))
	if err != nil {
		return nil, fmt.Errorf("failed to load YAML from file: %s; %v", configFilePath, err)
	}
	return config, nil
}

// NewProjectFromConfigFile creates a new Project based on the specified
// config file.
func NewProjectFromConfigFile(configFilePath string) (*Project, error) {
	logger.Trace.Printf("config: reading config from file: \"%v\"", configFilePath)

	config, err := NewConfigFromFile(configFilePath)
	if err != nil {
		return nil, err
	}
	return NewProject(*config)
}

// NewProjectFromYAML creates a new Project based on the specified YAML
// config.
func NewProjectFromYAML(yamlBytes []byte) (*Project, error) {
	config, err := configFromYAML(yamlBytes, "")
	if err != nil {
		return nil, err
	}

	return NewProject(*config)
}

// configFromYAML takes YAML bytes and returns a Project configuration
// struct.
func configFromYAML(yamlBytes []byte, rootPath string) (*conf.Project, error) {
	var config conf.Project

	err := yaml.Unmarshal(yamlBytes, &config)
	if err != nil {
		return nil, err
	}

	// Convert rootPath to absolute
	rootPath, err = filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	// Rewrite paths to absolute
	if err := rewriteConfigPaths(rootPath, &config); err != nil {
		return nil, fmt.Errorf("failed to resolve relative paths in config file: %s; %v", rootPath, err)
	}

	// Set configuration defaults
	if err := setDefaults(&config, rootPath); err != nil {
		return nil, err
	}

	// Fill in Terraform versions. This has to be done after paths are
	// rewritten.
	if err := setTerraformVersionFields(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setDefaults fills in a bunch of default values for the config.
func setDefaults(config *conf.Project, rootPath string) error {
	logger.Trace.Printf("config: setting defaults, rootPath: \"%v\"", rootPath)

	// For cases where we're creating a new project that is not from a
	// configuration file (e.g. in tests), we'll use the current working
	// directory as the root path.
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if config.TerraformDefaults.Path == "" && config.TerraformDefaults.Version == nil {
		if err := config.TerraformDefaults.SetDefaultPath(); err != nil {
			return err
		}
	}

	// Terraform code root is the root path of the config file (if it was
	// loaded from a file) otherwise is set to the current working dir.
	if config.TerraformCodeRoot == "" {
		if rootPath != "" {
			absRootPath, err := filepath.Abs(rootPath)
			if err != nil {
				return err
			}
			config.TerraformCodeRoot = absRootPath
		} else {
			config.TerraformCodeRoot = cwd
		}
	}

	if config.SessionRepoDir == "" {
		if rootPath != "" {
			config.SessionRepoDir = rootPath
		} else {
			config.SessionRepoDir = cwd
		}
	}

	// Fill in module defaults
	for i := range config.Modules {
		logger.Trace.Printf("config: applying default TerraformCodeRoot: \"%v\"", config.TerraformCodeRoot)
		config.Modules[i].Hooks.ApplyDefaultsFrom(config.Hooks)
		config.Modules[i].TerraformCodeRoot = config.TerraformCodeRoot
		config.Modules[i].Terraform.ApplyDefaultsFrom(config.TerraformDefaults)
	}

	return nil
}

// setTerraformVersionFields detects the Terraform version for any version
// fields that are unset and fills it in.
func setTerraformVersionFields(config *conf.Project) error {
	if config.TerraformDefaults.Version == nil {
		if err := config.TerraformDefaults.SetVersionFromBinary(); err != nil {
			return err
		}
	}
	for i := range config.Modules {
		if config.Modules[i].Terraform.Version == nil {
			if err := config.Modules[i].Terraform.SetVersionFromBinary(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Rewrite relative paths in the config file to be absolute paths.
func rewriteConfigPaths(rootPath string, config *conf.Project) error {
	if err := rewriteRelPaths(rootPath, false,
		&config.SessionRepoDir,
		&config.TerraformCodeRoot,
		&config.TerraformDefaults.Path); err != nil {
		return err
	}

	if err := rewriteRelPathsInSlices(rootPath, config.Hooks.Startup, config.Hooks.PreModuleRun); err != nil {
		return err
	}

	for _, moduleConfig := range config.Modules {
		if err := rewriteRelPathsInSlices(rootPath, moduleConfig.Hooks.PreModuleRun); err != nil {
			return err
		}
	}

	return nil
}

// rewriteRelPaths rewrites all relative paths to be absolute - relative to
// the specified root dir. If the path is already absolute, it is left
// untouched. If a path is empty, it is left empty.
func rewriteRelPaths(root string, isCommand bool, relpaths ...*string) error {
	for _, path := range relpaths {
		if *path == "" {
			continue
		}

		if filepath.IsAbs(*path) {
			continue
		}

		// do not resolve commands that are not explicit relative paths (rewrites ./foo but not foo)
		if isCommand && shouldSearchExecutableInOSPath(*path) {
			continue
		}

		rootAbsPath, err := filepath.Abs(root)
		if err != nil {
			return err
		}

		newPath := filepath.Join(rootAbsPath, *path)
		logger.Trace.Printf("config: rewriting path \"%v\" to \"%v\"", *path, newPath)
		*path = newPath
	}

	return nil
}

func rewriteRelPathsInSlices(root string, relpaths ...[]conf.Hook) error {
	for i := range relpaths {
		for j := range relpaths[i] {
			if err := rewriteRelPaths(root, true, &relpaths[i][j].Command); err != nil {
				return err
			}
		}

	}
	return nil
}

// shouldSearchExecutableInOSPath checks whether the executable should be searched in the operating system path or resolved relative to current direvtory
func shouldSearchExecutableInOSPath(path string) bool {
	return !strings.ContainsRune(path, filepath.Separator)
}
