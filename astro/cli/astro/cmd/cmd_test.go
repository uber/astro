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

package cmd_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/uber/astro/astro/tvm"
	"github.com/uber/astro/astro/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// Add Terraform versions here to run the tests against those
	// versions.
	terraformVersionsToTest = []string{
		"0.7.13",
		"0.8.8",
		"0.9.11",
		"0.10.8",
		"0.11.5",
	}
)

var (
	// During setup, this is set to the path of the compiled astro
	// binary so tests can execute it.
	astroBinary string

	// During setup, this is initialized with a Terraform version
	// repository so multiple versions of Terraform can be tested.
	terraformVersionRepo *tvm.VersionRepo
)

// compiles the astro binary and returns the path to it.
func compileAstro() (string, error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}

	out, err := exec.Command("go", "build", "-o", f.Name(), "..").CombinedOutput()
	if err != nil {
		return "", errors.New(string(out))
	}

	return f.Name(), nil
}

func TestMain(m *testing.M) {
	// compile the astro binary so we can execute it during tests
	binary, err := compileAstro()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	astroBinary = binary

	// Initialize Terraform version repository
	terraformVersionRepo, err = tvm.NewVersionRepoForCurrentSystem("")
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	// Download Terraform versions first so that multiple tests don't
	// try to do it in parallel.
	for _, version := range terraformVersionsToTest {
		if _, err := terraformVersionRepo.Get(version); err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
	}

	os.Exit(m.Run())
}

type testResult struct {
	Stdout  *bytes.Buffer
	Stderr  *bytes.Buffer
	Err     error
	Version string
}

func runTest(t *testing.T, args []string, fixtureBasePath string, version string) *testResult {
	fixturePath := fixtureBasePath

	// Determine if this version has a version-specific fixture.
	versionSpecificFixturePath := fmt.Sprintf("%s-%s", fixtureBasePath, version)
	if utils.FileExists(versionSpecificFixturePath) {
		fixturePath = versionSpecificFixturePath
	}

	// If there is a Makefile in the fixture, run make to set up any prereqs
	// for the test.
	if utils.FileExists(filepath.Join(fixturePath, "Makefile")) {
		make := exec.Command("make")
		make.Dir = fixturePath
		out, err := make.CombinedOutput()
		if err != nil {
			fmt.Fprint(os.Stderr, string(out))
		}
		require.NoError(t, err)
	}

	// Get Terraform path
	terraformBinaryPath, err := terraformVersionRepo.Get(version)
	require.NoError(t, err)

	// Override Terraform path
	terraformBinaryDir := filepath.Dir(terraformBinaryPath)

	// TODO: this blocks us from running multiple tests in parallel.
	// Need a better way to override the version externally.
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", fmt.Sprintf("%s:%s", terraformBinaryDir, oldPath))
	defer os.Setenv("PATH", oldPath)

	cmd := exec.Command(astroBinary, args...)
	cmd.Dir = fixturePath

	stdoutBytes := &bytes.Buffer{}
	stderrBytes := &bytes.Buffer{}

	cmd.Stdout = stdoutBytes
	cmd.Stderr = stderrBytes

	cmdErr := cmd.Run()

	return &testResult{
		Err:     cmdErr,
		Stdout:  stdoutBytes,
		Stderr:  stderrBytes,
		Version: version,
	}
}

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

func TestHelpWorks(t *testing.T) {
	result := runTest(t, []string{"--help"}, "", terraformVersionsToTest[len(terraformVersionsToTest)-1])
	assert.Contains(t, "A tool for managing multiple Terraform modules", result.Stdout.String())
	assert.NoError(t, result.Err)
}

func TestProjectApplyChangesSuccess(t *testing.T) {
	for _, version := range terraformVersionsToTest {
		t.Run(version, func(t *testing.T) {
			err := os.RemoveAll("/tmp/terraform-tests/apply-changes-success")
			require.NoError(t, err)

			err = os.MkdirAll("/tmp/terraform-tests/apply-changes-success", 0775)
			require.NoError(t, err)

			result := runTest(t, []string{"apply"}, "fixtures/apply-changes-success", version)
			assert.Contains(t, result.Stdout.String(), "foo: [32mOK")
			assert.Empty(t, result.Stderr.String())
			assert.NoError(t, result.Err)
		})
	}
}

func TestProjectPlanSuccessNoChanges(t *testing.T) {
	for _, version := range terraformVersionsToTest {
		t.Run(version, func(t *testing.T) {
			result := runTest(t, []string{"plan", "--trace"}, "fixtures/plan-success-nochanges", version)
			assert.Equal(t, "foo: \x1b[32mOK\x1b[0m\x1b[37m No changes\x1b[0m\x1b[37m (0s)\x1b[0m\nDone\n", result.Stdout.String())
			assert.NoError(t, result.Err)
		})
	}
}

func TestProjectPlanSuccessChanges(t *testing.T) {
	for _, version := range terraformVersionsToTest {
		t.Run(version, func(t *testing.T) {
			result := runTest(t, []string{"plan"}, "fixtures/plan-success-changes", version)
			assert.Contains(t, result.Stdout.String(), "foo: [32mOK[0m[33m Changes[0m[37m")
			assert.Regexp(t, `\+.*null_resource.foo`, result.Stdout.String())
			assert.NoError(t, result.Err)
		})
	}
}

func TestProjectPlanError(t *testing.T) {
	for _, version := range terraformVersionsToTest {
		t.Run(version, func(t *testing.T) {
			result := runTest(t, []string{"plan"}, "fixtures/plan-error", version)
			assert.Contains(t, result.Stderr.String(), "foo: [31mERROR")
			assert.Contains(t, result.Stderr.String(), "Error parsing")
			assert.Error(t, result.Err)
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

			result := runTest(t, []string{"plan", "--detach"}, "fixtures/plan-detach", version)
			require.Empty(t, result.Stderr.String())
			require.NoError(t, result.Err)
			require.Equal(t, "foo: \x1b[32mOK\x1b[0m\x1b[37m No changes\x1b[0m\x1b[37m (0s)\x1b[0m\nDone\n", result.Stdout.String())

			sessionDirs, err := getSessionDirs("/tmp/terraform-tests/plan-detach/.astro")
			require.NoError(t, err)
			require.Equal(t, 1, len(sessionDirs), "unable to find session: expect only a single session to have been written, found multiple")

			_, err = os.Stat(filepath.Join("/tmp/terraform-tests/plan-detach/.astro", sessionDirs[0], "foo/sandbox/terraform.tfstate"))
			assert.NoError(t, err)
		})
	}
}
