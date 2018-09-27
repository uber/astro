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

package cmd

import (
	"testing"

	"github.com/uber/astro/astro"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoVariables(t *testing.T) {
	c, err := astro.NewConfigFromFile("fixtures/flags/no_variables.yaml")
	require.NoError(t, err)

	flags, err := commandLineFlags(c)
	require.NoError(t, err)

	assert.Equal(t, []*Flag{}, flags)
}

func TestSimpleVariables(t *testing.T) {
	c, err := astro.NewConfigFromFile("fixtures/flags/simple_variables.yaml")
	require.NoError(t, err)

	flags, err := commandLineFlags(c)
	require.NoError(t, err)

	assert.Equal(t, []*Flag{
		&Flag{
			Variable:   "no_flag",
			Flag:       "no_flag",
			IsRequired: true,
		},
		&Flag{
			Variable:   "with_flag",
			Flag:       "flag_name",
			IsRequired: true,
		},
		&Flag{
			Variable: "with_values",
			Flag:     "with_values",
			IsFilter: true,
			AllowedValues: []string{
				"dev",
				"prod",
				"staging",
			},
		},
	}, flags)
}

func TestMergeValues(t *testing.T) {
	c, err := astro.NewConfigFromFile("fixtures/flags/merge_values.yaml")
	require.NoError(t, err)

	flags, err := commandLineFlags(c)
	require.NoError(t, err)

	assert.Equal(t, []*Flag{
		&Flag{
			Variable: "environment",
			Flag:     "environment",
			IsFilter: true,
			AllowedValues: []string{
				"dev",
				"mgmt",
				"prod",
				"staging",
			},
		},
	}, flags)
}
