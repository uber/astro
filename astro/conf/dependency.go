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

package conf

// Dependency is static config representing the dependency of a module
type Dependency struct {
	// Module is the name of the module we're depending on.
	Module string
	// Variables is an optional map of specific parameters to narrow down the
	// dependency to a specific execution. If this is nil, and the module has
	// many different possible executions, we'll depend on all of them.
	Variables map[string]string
}
