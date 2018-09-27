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

/* Taken from the Terraform source code. */

package astro

import "github.com/hashicorp/terraform/dag"

const rootNodeName = "root"

type graphNodeRoot struct{}

func (n graphNodeRoot) Name() string {
	return rootNodeName
}

func addRoot(g *dag.AcyclicGraph) error {
	// If we already have a good root, we're done
	if _, err := g.Root(); err == nil {
		return nil
	}

	// Add a root
	var root graphNodeRoot
	g.Add(root)

	// Connect the root to all the edges that need it
	for _, v := range g.Vertices() {
		if v == root {
			continue
		}

		if g.UpEdges(v).Len() == 0 {
			g.Connect(dag.BasicEdge(root, v))
		}
	}

	return nil
}
