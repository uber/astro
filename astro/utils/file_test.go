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

package utils_test

import (
	"testing"

	"github.com/uber/astro/astro/utils"

	"github.com/stretchr/testify/assert"
)

func TestIsWithinPath(t *testing.T) {
	tt := []struct {
		basepath string
		path     string
		result   bool
	}{
		{"/home/bob", "/home/bob/foo", true},
		{"/home/bob", "/home/bob/", true},
		{"/home/bob", "/home/bob/..", false},
		{"/home/bob", "/home/bob/foo/..", true},
		{"/home/bob", "/tmp/bob", false},
		{"/home/bob", "/home/bobcat", false},
	}

	for _, test := range tt {
		assert.Equal(t, test.result, utils.IsWithinPath(test.basepath, test.path))
	}
}
