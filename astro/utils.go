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

// cartesian returns the cartesian product of two (or more) lists of
// interfaces.
func cartesian(params ...[]interface{}) (finalResults [][]interface{}) {
	var iterate func([]interface{}, ...[]interface{})

	iterate = func(singleResult []interface{}, params ...[]interface{}) {
		if len(params) == 0 {
			finalResults = append(finalResults, singleResult)
			return
		}

		p, params := params[0], params[1:]
		for i := 0; i < len(p); i++ {
			iterate(append(singleResult, p[i]), params...)
		}
	}

	iterate([]interface{}{}, params...)

	return finalResults
}

// filterMaps checks that the values of matching keys in a and b are the same.
// NOTE: Keys that don't match are IGNORED.
func filterMaps(a, b map[string]string) bool {
	for key := range a {
		// Key doesn't exist; move on to the next key
		if _, ok := b[key]; !ok {
			continue
		}
		if a[key] != b[key] {
			return false
		}
	}
	return true
}
