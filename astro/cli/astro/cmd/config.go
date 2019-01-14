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

package cmd

import (
	"errors"
	"fmt"

	"github.com/uber/astro/astro"
	"github.com/uber/astro/astro/conf"
	"github.com/uber/astro/astro/utils"
)

// configFileSearchPaths is the default list of paths the astro CLI
// will attempt to find a config file at.
var configFileSearchPaths = []string{
	"astro.yaml",
	"astro.yml",
	"terraform/astro.yaml",
	"terraform/astro.yml",
}

var errCannotFindConfig = errors.New("unable to find config file")

// Global cache
var (
	_conf    *conf.Project
	_project *astro.Project
)

// firstExistingFilePath takes a list of paths and returns the first one
// where a file exists (or symlink to a file).
func firstExistingFilePath(paths ...string) string {
	for _, f := range paths {
		if utils.FileExists(f) {
			return f
		}
	}
	return ""
}

// configFile returns the path of the project config file.
func configFile() (string, error) {
	// User provided config file path takes precedence
	if userCfgFile != "" {
		return userCfgFile, nil
	}

	// Try to find the config file
	if path := firstExistingFilePath(configFileSearchPaths...); path != "" {
		return path, nil
	}

	return "", errCannotFindConfig
}

// currentConfig loads configuration or returns the previously loaded config.
func currentConfig() (*conf.Project, error) {
	if _conf != nil {
		return _conf, nil
	}

	file, err := configFile()
	if err != nil {
		return nil, err
	}
	_conf, err = astro.NewConfigFromFile(file)

	return _conf, err
}

// currentProject creates a new astro project or returns the previously created
// astro project.
func currentProject() (*astro.Project, error) {
	if _project != nil {
		return _project, nil
	}

	config, err := currentConfig()
	if err != nil {
		return nil, err
	}
	c, err := astro.NewProject(*config)
	if err != nil {
		return nil, fmt.Errorf("unable to load module configuration: %v", err)
	}

	_project = c

	return _project, nil
}
