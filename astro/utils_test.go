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
)

func TestCartesian(t *testing.T) {
	t.Parallel()

	params := [][]interface{}{
		[]interface{}{"a", "b", "c"},
		[]interface{}{1, 2, 3},
		[]interface{}{"x", "y"},
	}
	res := cartesian(params...)

	assert.Equal(t, [][]interface{}{
		[]interface{}{"a", 1, "x"},
		[]interface{}{"a", 1, "y"},
		[]interface{}{"a", 2, "x"},
		[]interface{}{"a", 2, "y"},
		[]interface{}{"a", 3, "x"},
		[]interface{}{"a", 3, "y"},
		[]interface{}{"b", 1, "x"},
		[]interface{}{"b", 1, "y"},
		[]interface{}{"b", 2, "x"},
		[]interface{}{"b", 2, "y"},
		[]interface{}{"b", 3, "x"},
		[]interface{}{"b", 3, "y"},
		[]interface{}{"c", 1, "x"},
		[]interface{}{"c", 1, "y"},
		[]interface{}{"c", 2, "x"},
		[]interface{}{"c", 2, "y"},
		[]interface{}{"c", 3, "x"},
		[]interface{}{"c", 3, "y"},
	}, res)
}
