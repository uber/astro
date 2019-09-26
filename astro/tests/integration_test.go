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
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"syscall"
	"testing"
	"time"

	"github.com/uber/astro/astro/terraform"
	"github.com/uber/astro/astro/utils"

	"github.com/burl/go-version"
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

// stringVersionMatches returns whether or not the version passed as string matches the constraint.
// See terraform.VersionMatches for more info.
func stringVersionMatches(v string, versionConstraint string) bool {
	return terraform.VersionMatches(version.Must(version.NewVersion(v)), versionConstraint)
}

// compiles the astro binary and returns the path to it.
func compileAstro(dir string) (string, error) {
	astroPath := filepath.Join(dir, "astro")
	packageName := "github.com/uber/astro/astro/cli/astro"
	out, err := exec.Command("go", "build", "-o", astroPath, packageName).CombinedOutput()
	if err != nil {
		return "", errors.New(string(out))
	}
	return astroPath, nil
}

func TestPlanInterrupted(t *testing.T) {
	fakeTerraformPath := "fixtures/terraform"
	require.True(t, utils.FileExists(fakeTerraformPath))
	fakeTerraformDir, err := filepath.Abs(filepath.Dir(fakeTerraformPath))
	require.NoError(t, err)

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", fmt.Sprintf("%s:%s", fakeTerraformDir, oldPath))
	defer os.Setenv("PATH", oldPath)

	tmpdir, err := ioutil.TempDir("", "astro-binary")
	defer os.RemoveAll(tmpdir)
	require.NoError(t, err)

	astroBinary, err := compileAstro(tmpdir)
	require.NoError(t, err)
	command := exec.Command(astroBinary, "plan")

	fixtureAbsPath, err := filepath.Abs("fixtures/plan-interrupted")
	require.NoError(t, err)
	command.Dir = fixtureAbsPath

	stdoutBytes := &bytes.Buffer{}
	stderrBytes := &bytes.Buffer{}
	command.Stdout = stdoutBytes
	command.Stderr = stderrBytes

	var cmdErr error
	processChan := make(chan struct{}, 1)
	go func() {
		defer close(processChan)
		cmdErr = command.Run()
		processChan <- struct{}{}
	}()

	// let astro start terraform processes
	time.Sleep(1000 * time.Millisecond)
	require.NoError(t, command.Process.Signal(syscall.SIGINT))

	select {
	case <-processChan:
	case <-time.After(5 * time.Second):
		// force kill the process after timeout
		require.NoError(t, command.Process.Signal(syscall.SIGKILL))
	}

	require.Error(t, cmdErr)
	require.Equal(t, 1, command.ProcessState.ExitCode())

	stdout := stdoutBytes.String()
	stderr := stderrBytes.String()
	assert.Contains(t, stdout, "\nReceived signal: interrupt, cancelling operation...\n")
	assert.Regexp(t, `foo\d{2}:.*ERROR`, stderr)
	assert.NotRegexp(t, `foo\d{2}:`, stdout)
	assert.NotRegexp(t, `bar\d{2}:`, stdout)
	assert.NotRegexp(t, `bar\d{2}:`, stderr)
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
			addedResourceRe := `\+.*null_resource.foo`
			if stringVersionMatches(version, ">=0.12") {
				addedResourceRe = `null_resource.foo.*will be created`
			}
			assert.Regexp(t, addedResourceRe, result.Stdout.String())
			assert.Equal(t, 0, result.ExitCode)
		})
	}
}

func TestProjectPlanError(t *testing.T) {
	for _, version := range terraformVersionsToTest {
		t.Run(version, func(t *testing.T) {
			result := RunTest(t, []string{"plan"}, "fixtures/plan-error", version)
			assert.Contains(t, result.Stderr.String(), "foo: [31mERROR")
			errorMessage := "Error parsing"
			if stringVersionMatches(version, ">=0.12") {
				errorMessage = "Argument or block definition required"
			}
			assert.Contains(t, result.Stderr.String(), errorMessage)
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
