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

package tvm_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/uber/astro/astro/tvm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDownloadSameRepo does multiple downloads in parallel in the same repo
// which should test the syncMap / download lock.
func TestDownloadSameRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping download test in short mode.")
	}

	tmpdir, err := ioutil.TempDir("", "terraform-tests")
	require.NoError(t, err)

	defer os.RemoveAll(tmpdir)

	for i := 1; i <= 3; i++ {
		t.Run(string(i), func(t *testing.T) {
			t.Parallel()

			versions, err := tvm.NewVersionRepoForCurrentSystem(tmpdir)
			require.NoError(t, err)

			terraformBinary, err := versions.Get("0.7.13")
			require.NoError(t, err)

			version, err := tvm.InspectVersion(terraformBinary)
			require.NoError(t, err)

			assert.Equal(t, "0.7.13", version.String())
		})
	}
}

// TestDownloadDifferentRepo does multiple downloads in parallel
// reinitializing the repo every time. This will cause multiple downloads to
// occur but it will test for races that can occur when executing a binary that
// is currently being written to.
func TestDownloadDifferentRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping download test in short mode.")
	}

	tmpdir, err := ioutil.TempDir("", "terraform-tests")
	require.NoError(t, err)

	defer os.RemoveAll(tmpdir)

	versions, err := tvm.NewVersionRepoForCurrentSystem(tmpdir)
	require.NoError(t, err)

	for i := 1; i <= 3; i++ {
		t.Run(string(i), func(t *testing.T) {
			t.Parallel()

			terraformBinary, err := versions.Get("0.7.13")
			require.NoError(t, err)

			version, err := tvm.InspectVersion(terraformBinary)
			require.NoError(t, err)

			assert.Equal(t, "0.7.13", version.String())
		})
	}
}
