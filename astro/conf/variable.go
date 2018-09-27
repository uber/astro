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

// Variable represents a variable that can be passed into a
// Terraform module.
type Variable struct {
	// Name is the name/key of the variable.
	Name string
	// Flag is the command-line flag of the variable
	Flag string
	// Values is a list of possible values for the variable. A value of nil
	// means the possible values are unbound.
	Values []string
}

// CommandFlag is name of the command line flag that can be used to set this variable
func (v *Variable) CommandFlag() string {
	if v.Flag != "" {
		return v.Flag
	}
	return v.Name
}

// IsRequired returns true if the command-line parameter is mandatory
//
// The full behaviour is described in the README
func (v *Variable) IsRequired() bool {
	return len(v.Values) == 0
}

// IsFilter returns true if the command-line parameter acts as a filter
//
// The full behaviour is described in the README
func (v *Variable) IsFilter() bool {
	return len(v.Values) > 0
}
