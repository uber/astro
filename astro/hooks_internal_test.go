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
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHookStartupSuccess(t *testing.T) {
	t.Parallel()

	c, err := NewProjectFromConfigFile("fixtures/test-hook-startup-success/astro.yaml")
	require.NoError(t, err)
	require.NotNil(t, c)

	session, err := c.sessions.Current()
	require.NoError(t, err)

	b, err := ioutil.ReadFile(filepath.Join(session.path, "mock-hook.log"))
	require.NoError(t, err)

	assert.Equal(t, "SUCCESS\n", string(b))
}

func TestHookStartupFail(t *testing.T) {
	t.Parallel()

	c, err := NewProjectFromConfigFile("fixtures/test-hook-startup-fail/astro.yaml")
	require.Nil(t, c)
	assert.Contains(t, err.Error(), "error running Startup hook")
}

func TestHookInjectEnvVars(t *testing.T) {
	t.Parallel()

	c, err := NewProjectFromConfigFile("fixtures/test-hook-inject-env-vars/astro.yaml")
	require.NoError(t, err)
	require.NotNil(t, c)

	_, resultChan, err := c.Plan(nil, NoUserVariables(), false)
	assert.NoError(t, err)

	// there should be no errors
	assert.Equal(t, map[string]error{
		"test": nil,
	}, testResultErrs(testReadResults(resultChan)))
}
