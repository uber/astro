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

package astro_test

import (
	"testing"

	"github.com/uber/astro/astro"

	"github.com/stretchr/testify/assert"
)

// TestMissingDependencyExecution tests that we fail when a particular
// execution of a dependency does not exist in the definition of that
// module.
func TestMissingDependencyExecution(t *testing.T) {
	t.Parallel()

	c, err := astro.NewProjectFromConfigFile("fixtures/test-missing-dependency-execution/astro.yaml")

	assert.Nil(t, c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid dependency for app: no execution matching dep")
}

// TestMissingDependencyModule tests that an error is thrown if the
// module is not defined at all in the config.
func TestMissingDependencyModule(t *testing.T) {
	t.Parallel()

	c, err := astro.NewProjectFromConfigFile("fixtures/test-missing-dependency-module/astro.yaml")

	assert.Nil(t, c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid dependency for app: missing dependency: foo")
}

// TODO: Test multiple modules with the same name
