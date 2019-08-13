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

package tests

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getSessionDirs returns a list of the sessions inside a session repository.
// This excludes other directories that might have been created in there, e.g.
// the shared plugin cache directory.
func getSessionDirs(sessionBaseDir string) ([]string, error) {
	sessionRegexp, err := regexp.Compile("[0-9A-Z]{26}")
	if err != nil {
		return nil, err
	}

	dirs, err := ioutil.ReadDir(sessionBaseDir)
	if err != nil {
		return nil, err
	}

	sessionDirs := []string{}

	for _, dir := range dirs {
		if sessionRegexp.MatchString(dir.Name()) {
			sessionDirs = append(sessionDirs, dir.Name())
		}
	}

	return sessionDirs, nil
}

func TestProjectApplyChangesSuccess(t *testing.T) {
	for _, version := range terraformVersionsToTest {
		t.Run(version, func(t *testing.T) {
			err := os.RemoveAll("/tmp/terraform-tests/apply-changes-success")
			require.NoError(t, err)

			err = os.MkdirAll("/tmp/terraform-tests/apply-changes-success", 0775)
			require.NoError(t, err)

			result := RunTest(t, []string{"apply"}, "fixtures/apply-changes-success", version)
			assert.Contains(t, result.Stdout.String(), "foo: [32mOK")
			assert.Empty(t, result.Stderr.String())
			assert.Equal(t, 0, result.ExitCode)
		})
	}
}

func TestProjectPlanSuccessNoChanges(t *testing.T) {
	for _, version := range terraformVersionsToTest {
		t.Run(version, func(t *testing.T) {
			result := RunTest(t, []string{"plan", "--trace"}, "fixtures/plan-success-nochanges", version)
			assert.Equal(t, "foo: \x1b[32mOK\x1b[0m\x1b[37m No changes\x1b[0m\x1b[37m (0s)\x1b[0m\nDone\n", result.Stdout.String())
			assert.Equal(t, 0, result.ExitCode)
		})
	}
}

func TestProjectPlanSuccessChanges(t *testing.T) {
	for _, version := range terraformVersionsToTest {
		t.Run(version, func(t *testing.T) {
			result := RunTest(t, []string{"plan"}, "fixtures/plan-success-changes", version)
			assert.Contains(t, result.Stdout.String(), "foo: [32mOK[0m[33m Changes[0m[37m")
			assert.Regexp(t, `\+.*null_resource.foo`, result.Stdout.String())
			assert.Equal(t, 0, result.ExitCode)
		})
	}
}

func TestProjectPlanError(t *testing.T) {
	for _, version := range terraformVersionsToTest {
		t.Run(version, func(t *testing.T) {
			result := RunTest(t, []string{"plan"}, "fixtures/plan-error", version)
			assert.Contains(t, result.Stderr.String(), "foo: [31mERROR")
			assert.Contains(t, result.Stderr.String(), "Error parsing")
			assert.Equal(t, 1, result.ExitCode)
		})
	}
}

func TestProjectPlanDetachSuccess(t *testing.T) {
	for _, version := range terraformVersionsToTest {
		t.Run(version, func(t *testing.T) {
			err := os.RemoveAll("/tmp/terraform-tests/plan-detach")
			require.NoError(t, err)

			err = os.MkdirAll("/tmp/terraform-tests/plan-detach", 0775)
			require.NoError(t, err)

			result := RunTest(t, []string{"plan", "--detach"}, "fixtures/plan-detach", version)
			require.Empty(t, result.Stderr.String())
			require.Equal(t, 0, result.ExitCode)
			require.Equal(t, "foo: \x1b[32mOK\x1b[0m\x1b[37m No changes\x1b[0m\x1b[37m (0s)\x1b[0m\nDone\n", result.Stdout.String())

			sessionDirs, err := getSessionDirs("/tmp/terraform-tests/plan-detach/.astro")
			require.NoError(t, err)
			require.Equal(t, 1, len(sessionDirs), "unable to find session: expect only a single session to have been written, found multiple")

			_, err = os.Stat(filepath.Join("/tmp/terraform-tests/plan-detach/.astro", sessionDirs[0], "foo/sandbox/terraform.tfstate"))
			assert.NoError(t, err)
		})
	}
}
