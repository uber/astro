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
	"os"
	"testing"

	"github.com/uber/astro/astro/tvm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectUsesDefaultTerraformVersion(t *testing.T) {
	t.Parallel()

	c, err := NewProjectFromConfigFile("fixtures/test-terraform-default-version/astro.yaml")
	require.NoError(t, err)

	// Init test Terraform version repo in fixtures
	testVersionRepo, err := tvm.NewVersionRepo(absolutePath("fixtures/tvm"), "all", "all")
	require.NoError(t, err)

	// Install test/mock
	c.terraformVersions = testVersionRepo

	executions := c.executions(NoExecutionParameters())
	require.NotEmpty(t, executions)

	// Take a single execution and bind it
	e := executions[0].(*unboundExecution)
	b, err := e.bind(map[string]string{
		"aws_region": "test1",
	})
	require.NoError(t, err)

	// Init session
	session, err := c.sessions.NewSession()
	require.NoError(t, err)

	terraform, err := session.newTerraformSession(b)
	require.NoError(t, err)

	version, err := terraform.Version()
	require.NoError(t, err)

	assert.Equal(t, "0.11.6", version.String())
}

func TestProjectUsesDefaultTerraformPath(t *testing.T) {
	t.Parallel()

	c, err := NewProjectFromConfigFile("fixtures/test-terraform-default-path/astro.yaml")
	require.NoError(t, err)

	// Init test Terraform version repo in fixtures
	testVersionRepo, err := tvm.NewVersionRepo(absolutePath("fixtures/tvm"), "all", "all")
	require.NoError(t, err)

	// Install test/mock
	c.terraformVersions = testVersionRepo

	executions := c.executions(NoExecutionParameters())
	require.NotEmpty(t, executions)

	// Take a single execution and bind it
	e := executions[0].(*unboundExecution)
	b, err := e.bind(map[string]string{
		"aws_region": "test1",
	})
	require.NoError(t, err)

	// Init session
	session, err := c.sessions.NewSession()
	require.NoError(t, err)

	terraform, err := session.newTerraformSession(b)
	require.NoError(t, err)

	version, err := terraform.Version()
	require.NoError(t, err)

	assert.Equal(t, "0.8.8", version.String())
}

func TestSharedPluginCache(t *testing.T) {
	oldVal := os.Getenv("TF_PLUGIN_CACHE_DIR")
	defer os.Setenv("TF_PLUGIN_CACHE_DIR", oldVal)

	os.Unsetenv("TF_PLUGIN_CACHE_DIR")

	// Configration points to a mock version of Terraform that verifies if
	// TF_PLUGIN_CACHE_DIR is set
	c, err := NewProjectFromConfigFile("fixtures/test-terraform-shared-plugin-cache/astro.yaml")
	require.NoError(t, err)

	// do a plan
	_, resultChan, err := c.Plan(NoPlanExecutionParameters())
	require.NoError(t, err)

	// assert no errors
	assert.Equal(t, map[string]error{
		"test": nil,
	}, testResultErrs(testReadResults(resultChan)))
}

func TestSharedPluginCachePreservesExisting(t *testing.T) {
	oldVal := os.Getenv("TF_PLUGIN_CACHE_DIR")
	defer os.Setenv("TF_PLUGIN_CACHE_DIR", oldVal)

	os.Setenv("TF_PLUGIN_CACHE_DIR", "foobar")

	// Configration points to a mock version of Terraform that verifies if
	// TF_PLUGIN_CACHE_DIR is set to "foobar"
	c, err := NewProjectFromConfigFile("fixtures/test-terraform-shared-plugin-cache-preserve-existing/astro.yaml")
	require.NoError(t, err)

	// do a plan
	_, resultChan, err := c.Plan(NoPlanExecutionParameters())
	require.NoError(t, err)

	// assert no errors
	assert.Equal(t, map[string]error{
		"test": nil,
	}, testResultErrs(testReadResults(resultChan)))
}
