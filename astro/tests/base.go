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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uber/astro/astro/cli/astro/cmd"
	"github.com/uber/astro/astro/tvm"
	"github.com/uber/astro/astro/utils"
)

var (
	// During setup, this is initialized with a Terraform version
	// repository so multiple versions of Terraform can be tested.
	terraformVersionRepo *tvm.VersionRepo
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

const VERSION_LATEST = ""

type TestResult struct {
	Stdout   *bytes.Buffer
	Stderr   *bytes.Buffer
	ExitCode int
	Version  string
}

func init() {
	var err error

	// Initialize Terraform version repository
	terraformVersionRepo, err = tvm.NewVersionRepoForCurrentSystem("")
	if err != nil {
		panic(err)
	}

	// Download Terraform versions first so that multiple tests don't
	// try to do it in parallel.
	for _, version := range terraformVersionsToTest {
		if _, err := terraformVersionRepo.Get(version); err != nil {
			panic(err)
		}
	}
}

func RunTest(t *testing.T, args []string, fixtureBasePath string, version string) *TestResult {
	fixturePath := fixtureBasePath

	// If requested version is empty, assume the latest
	if version == VERSION_LATEST {
		version = terraformVersionsToTest[len(terraformVersionsToTest)-1]
	}

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

	// This also blocks us from running in parallel (the need to chdir)
	oldDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	os.Chdir(fixturePath)
	defer os.Chdir(oldDir)

	stdoutBytes := &bytes.Buffer{}
	stderrBytes := &bytes.Buffer{}

	cli, err := cmd.NewAstroCLI(
		cmd.WithStdout(stdoutBytes),
		cmd.WithStderr(stderrBytes),
	)
	require.NoError(t, err)

	exitCode := cli.Run(args)

	return &TestResult{
		ExitCode: exitCode,
		Stdout:   stdoutBytes,
		Stderr:   stderrBytes,
		Version:  version,
	}
}
