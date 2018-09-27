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

import "github.com/uber/astro/astro/utils"

// configFileSearchPaths is the default list of paths the astro CLI
// will attempt to find a config file at.
var configFileSearchPaths = []string{
	"astro.yaml",
	"astro.yml",
	"terraform/astro.yaml",
	"terraform/astro.yml",
}

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
