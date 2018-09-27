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

// UserVariables holds the values provided by the user via custom command line flags
type UserVariables struct {
	// Values is the set of all values specified by the user via custom command line flags
	Values map[string]string
	// Filters is the subset of values that act as module filters, as described by the README
	Filters map[string]bool
}

// NoUserVariables returns an empty UserVariables value, used by tests
func NoUserVariables() *UserVariables {
	return &UserVariables{}
}

// HasFilter is true when name is acting as a module filter
func (uv *UserVariables) HasFilter(name string) bool {
	return uv.Filters != nil && uv.Filters[name]
}

// FilterCount returns the number of module filters
func (uv *UserVariables) FilterCount() int {
	return len(uv.Filters)
}
