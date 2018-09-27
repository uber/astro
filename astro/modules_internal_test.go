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

package astro

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/astro/astro/conf"
)

func TestModuleExecution(t *testing.T) {
	t.Parallel()

	conf := conf.Module{
		Name: "TestModule",
		Path: "test",
		Variables: []conf.Variable{
			conf.Variable{
				Name: "aws_region",
			},
			conf.Variable{
				Name:   "environment",
				Values: []string{"dev", "staging", "prod"},
			},
		},
	}

	expected := executionSet{
		&unboundExecution{
			&execution{
				moduleConf: &conf,
				variables: map[string]string{
					"aws_region":  "{aws_region}",
					"environment": "dev",
				},
			},
		},
		&unboundExecution{
			&execution{
				moduleConf: &conf,
				variables: map[string]string{
					"aws_region":  "{aws_region}",
					"environment": "staging",
				},
			},
		},
		&unboundExecution{
			&execution{
				moduleConf: &conf,
				variables: map[string]string{
					"aws_region":  "{aws_region}",
					"environment": "prod",
				},
			},
		},
	}

	assert.EqualValues(t, expected, newModule(conf).executions(NoUserVariables()))
}
