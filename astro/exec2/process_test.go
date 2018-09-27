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

package exec2_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/uber/astro/astro/exec2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newHelloWorld() *exec2.Process {
	return exec2.NewProcess(exec2.Cmd{
		Command: "/bin/sh",
		Args:    []string{"-c", "echo Hello, world!"},
	})
}

func TestProcess(t *testing.T) {
	process := newHelloWorld()

	err := process.Run()
	require.NoError(t, err)

	assert.True(t, process.Success())
	assert.Equal(t, 0, process.ExitCode())
	assert.Equal(t, "Hello, world!\n", process.Stdout().String())
}

func TestProcessError(t *testing.T) {
	process := exec2.NewProcess(exec2.Cmd{
		Command: "/bin/sh",
		Args:    []string{"-c", "echo Houston, we have a problem >&2; exit 23"},
	})

	err := process.Run()
	assert.Error(t, err)

	assert.False(t, process.Success())
	assert.Equal(t, 23, process.ExitCode())
	assert.Equal(t, "", process.Stdout().String())
	assert.Equal(t, "Houston, we have a problem\n", process.Stderr().String())
}

func TestCombinedOutputLog(t *testing.T) {
	tmpLogFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)

	defer os.Remove(tmpLogFile.Name())

	// This process writes something to stdout and stderr
	process := exec2.NewProcess(exec2.Cmd{
		Command: "/bin/sh",
		Args:    []string{"-c", "echo Hello, world!; echo uhoh! >&2"},
		CombinedOutputLogFile: tmpLogFile.Name(),
	})

	err = process.Run()
	require.NoError(t, err)

	// Read contents back from log file
	logFileContents, err := ioutil.ReadAll(tmpLogFile)
	require.NoError(t, err)

	// Log file should be stdout/stderr combined; but we can't be sure
	// of the order.
	assert.Contains(t, string(logFileContents), "+ /bin/sh [-c echo Hello, world!; echo uhoh! >&2]")
	assert.Contains(t, string(logFileContents), "Hello, world")
	assert.Contains(t, string(logFileContents), "uhoh!")

	// Check that stdout/stderr are still captured correctly
	assert.Equal(t, "Hello, world!\n", process.Stdout().String())
	assert.Equal(t, "uhoh!\n", process.Stderr().String())
}

func TestExited(t *testing.T) {
	process := newHelloWorld()
	assert.False(t, process.Exited())
	process.Run()
	assert.True(t, process.Exited())
}
