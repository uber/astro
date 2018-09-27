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
	"testing"

	"github.com/uber/astro/astro/utils"

	version "github.com/burl/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func absolutePath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(fmt.Sprintf("unable to get absolute path for: %v", path))
	}
	return absPath
}

func TestTerraformCodeRootPaths(t *testing.T) {
	t.Parallel()

	c, err := NewProjectFromConfigFile("fixtures/test-rewrite-code-root-paths/astro.yaml")

	// There should be no validation errors
	assert.NoError(t, err)

	// Check TerraformCodeRoot is rewritten to the absolute path
	assert.Equal(t, absolutePath("fixtures/test-rewrite-code-root-paths/terraform"), c.config.TerraformCodeRoot)

	// Check that TerraformCodeRoot has propagated to all modules
	for _, module := range c.config.Modules {
		assert.Equal(t, absolutePath("fixtures/test-rewrite-code-root-paths/terraform"), module.TerraformCodeRoot)
	}
}

func TestModulePathCannotEscapeCodeRoot(t *testing.T) {
	t.Parallel()

	c, err := NewProjectFromConfigFile("fixtures/test-module-path-cannot-escape-code-root/astro.yaml")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "module path cannot be outside code root")
	assert.Nil(t, c)
}

func TestRewritePathsInternal(t *testing.T) {
	t.Parallel()

	tests := []string{
		"foo",
		"../bar",
		"/tmp/baz",
	}

	for i := range tests {
		expected := absolutePath(tests[i])
		rewriteRelPaths(absolutePath(""), false, &tests[i])
		assert.Equal(t, expected, tests[i])
	}
}

func TestSessionRepoDir(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	require.NoError(t, err)

	defer os.RemoveAll(tmpdir)

	testConfigFilePath := filepath.Join(tmpdir, "test-session-repo-dir.yaml")

	// copy test configuration into tmpdir
	err = os.Link("fixtures/test-session-repo-dir/astro.yaml", testConfigFilePath)
	require.NoError(t, err)

	// load astro
	c, err := NewProjectFromConfigFile(testConfigFilePath)
	require.NoError(t, err)

	assert.Equal(t, tmpdir, c.config.SessionRepoDir)

	// Check that a session repo dir has been created in the tmpdir
	if !utils.FileExists(filepath.Join(tmpdir, ".astro")) {
		assert.Fail(t, "missing .astro directory")
	}
}

func TestUnmarshalTerraformVersion(t *testing.T) {
	c, err := NewProjectFromConfigFile("fixtures/foosite.yaml")
	require.NoError(t, err)

	expectedObj, err := version.NewVersion("0.8.8")
	require.NoError(t, err)

	assert.Equal(t, expectedObj, c.config.TerraformDefaults.Version)
}
