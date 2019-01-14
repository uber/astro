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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelpUserFlags(t *testing.T) {
	result := runTest(t, []string{
		"--config=simple_variables.yaml",
		"plan",
		"--help",
	}, "fixtures/flags", VERSION_LATEST)
	assert.Contains(t, result.Stdout.String(), "User flags:")
	assert.Contains(t, result.Stdout.String(), "--foo")
	assert.Contains(t, result.Stdout.String(), "--baz")
	assert.Contains(t, result.Stdout.String(), "Baz Description")
	assert.Contains(t, result.Stdout.String(), "--qux")
}

func TestHelpNoUserFlags(t *testing.T) {
	result := runTest(t, []string{
		"--config=no_variables.yaml",
		"plan",
		"--help",
	}, "fixtures/flags", VERSION_LATEST)
	assert.NotContains(t, result.Stdout.String(), "User flags:")
}

func TestHelpShowsConfigLoadError(t *testing.T) {
	result := runTest(t, []string{
		"--config=/nonexistent/path/to/config",
		"plan",
		"--help",
	}, "fixtures/flags", VERSION_LATEST)
	assert.Contains(t, result.Stderr.String(), "There was an error loading astro config")
}

func TestHelpDoesntAlwaysShowLoadingError(t *testing.T) {
	result := runTest(t, []string{
		"--help",
	}, "fixtures/flags", VERSION_LATEST)
	assert.NotContains(t, result.Stderr.String(), "There was an error loading astro config")
}

func TestPlanErrorOnMissingValues(t *testing.T) {
	result := runTest(t, []string{
		"--config=simple_variables.yaml",
		"plan",
	}, "fixtures/flags", VERSION_LATEST)
	assert.Error(t, result.Err)
	assert.Contains(t, result.Stderr.String(), "missing required flags")
	assert.Contains(t, result.Stderr.String(), "--foo")
	assert.Contains(t, result.Stderr.String(), "--baz")
}

func TestPlanAllowedValues(t *testing.T) {
	tt := []string{
		"mgmt",
		"dev",
		"staging",
		"prod",
	}
	for _, env := range tt {
		t.Run(env, func(t *testing.T) {
			result := runTest(t, []string{
				"--config=merge_values.yaml",
				"plan",
				"--environment",
				env,
			}, "fixtures/flags", VERSION_LATEST)
			assert.NoError(t, result.Err)
		})
	}
}

func TestPlanFailOnNotAllowedValue(t *testing.T) {
	result := runTest(t, []string{
		"--config=merge_values.yaml",
		"plan",
		"--environment",
		"foo",
	}, "fixtures/flags", VERSION_LATEST)
	assert.Error(t, result.Err)
	assert.Contains(t, result.Stderr.String(), "invalid argument")
	assert.Contains(t, result.Stderr.String(), "allowed values")
}
