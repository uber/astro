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
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/astro/astro/tests"
)

func TestErrorDisplay(t *testing.T) {
	result := tests.RunTest(t, []string{"plan"}, "fixtures/plan-error", tests.VERSION_LATEST)

	re := regexp.MustCompile("There are some problems with the configuration")
	matches := re.FindAllString(result.Stderr.String(), -1)

	// Test that the error is only printed once
	assert.Exactly(t, 1, len(matches))
}
