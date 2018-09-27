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

	"github.com/stretchr/testify/require"
)

func TestGraph(t *testing.T) {
	t.Parallel()

	c, err := NewProjectFromConfigFile("fixtures/test-graph/astro.yaml")
	require.NoError(t, err)

	graph, err := c.executions(nil, NoUserVariables()).graph()
	require.NoError(t, err)
	require.NoError(t, graph.Validate())
	graph.TransitiveReduction()

	_, err = graph.Root()
	require.NoError(t, err)
}
