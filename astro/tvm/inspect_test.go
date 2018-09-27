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
	"testing"

	"github.com/uber/astro/astro/tvm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInspectOK(t *testing.T) {
	version, err := tvm.InspectVersion("test/terraform-version-ok")
	require.NoError(t, err)
	assert.Equal(t, "0.7.13", version.String())
}
func TestInspectFail(t *testing.T) {
	version, err := tvm.InspectVersion("test/terraform-fail")
	assert.Nil(t, version)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exit status 127")
}

func TestInspectBadVersion(t *testing.T) {
	version, err := tvm.InspectVersion("test/terraform-version-bad")
	assert.Nil(t, version)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to parse version from data")
}

func TestInspectEmptyVersion(t *testing.T) {
	version, err := tvm.InspectVersion("test/terraform-version-empty")
	assert.Nil(t, version)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to read lines from data")
}
