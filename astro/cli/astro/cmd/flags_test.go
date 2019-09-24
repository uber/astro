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

	"github.com/uber/astro/astro/tests"
)

func TestHelpWorks(t *testing.T) {
	result := tests.RunTest(t, []string{"--help"}, "fixtures/no-config", tests.VERSION_LATEST)
	assert.Contains(t, result.Stderr.String(), "A tool for managing multiple Terraform modules")
	assert.Equal(t, 0, result.ExitCode)
}
func TestHelpUserFlags(t *testing.T) {
	result := tests.RunTest(t, []string{
		"plan",
		"--help",
	}, "fixtures/config-simple", tests.VERSION_LATEST)
	assert.Contains(t, result.Stderr.String(), "User flags:")
	assert.Contains(t, result.Stderr.String(), "--foo")
	assert.Contains(t, result.Stderr.String(), "--baz")
	assert.Contains(t, result.Stderr.String(), "Baz Description")
	assert.Contains(t, result.Stderr.String(), "--qux")
}

func TestHelpNoUserFlags(t *testing.T) {
	result := tests.RunTest(t, []string{
		"--config=no_variables.yaml",
		"plan",
		"--help",
	}, "fixtures/flags", tests.VERSION_LATEST)
	assert.NotContains(t, result.Stderr.String(), "User flags:")
}

func TestConfigLoadErrorWhenSpecified(t *testing.T) {
	result := tests.RunTest(t, []string{
		"--config=/nonexistent/path/to/config",
		"plan",
		"--help",
	}, "fixtures/config-simple", tests.VERSION_LATEST)
	assert.Contains(t, result.Stderr.String(), "file does not exist")
	assert.Equal(t, 1, result.ExitCode)
}

func TestUnknownFlag(t *testing.T) {
	result := tests.RunTest(t, []string{
		"plan",
		"--foo",
		"bar",
	}, "fixtures/flags", tests.VERSION_LATEST)
	assert.Contains(t, result.Stderr.String(), "No astro config was loaded")
	assert.Equal(t, 1, result.ExitCode)
}

func TestPlanErrorOnMissingValues(t *testing.T) {
	result := tests.RunTest(t, []string{
		"plan",
	}, "fixtures/config-simple", tests.VERSION_LATEST)
	assert.Equal(t, 1, result.ExitCode)
	assert.Contains(t, result.Stderr.String(), "missing required flags")
	assert.Contains(t, result.Stderr.String(), "--foo")
	assert.Contains(t, result.Stderr.String(), "--baz")
}

func TestPlanAllowedValues(t *testing.T) {
	testcases := []struct {
		env             string
		expectedModules []string
	}{
		{
			"mgmt",
			[]string{"foo_mgmt-mgmt"},
		},
		{
			"dev",
			[]string{"misc-dev", "test_env-dev"},
		},
		{
			"staging",
			[]string{"misc-staging", "test_env-staging"},
		},
		{
			"prod",
			[]string{"misc-prod"},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.env, func(t *testing.T) {
			result := tests.RunTest(t, []string{
				"--config=merge_values.yaml",
				"plan",
				"--environment",
				tt.env,
			}, "fixtures/flags", tests.VERSION_LATEST)
			// Check that all expected modules were planned for environment
			for _, module := range tt.expectedModules {
				assert.Contains(t, result.Stdout.String(), module)
			}

			assert.Equal(t, 0, result.ExitCode)
		})
	}
}

func TestPlanFailOnNotAllowedValue(t *testing.T) {
	result := tests.RunTest(t, []string{
		"--config=merge_values.yaml",
		"plan",
		"--environment",
		"foo",
	}, "fixtures/flags", tests.VERSION_LATEST)
	assert.Equal(t, 1, result.ExitCode)
	assert.Contains(t, result.Stderr.String(), "invalid argument")
	assert.Contains(t, result.Stderr.String(), "allowed values")
}
