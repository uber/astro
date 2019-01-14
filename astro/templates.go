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
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

var (
	// matches "{fox}" in "the quick {fox}"
	reVarPlaceholder = regexp.MustCompile(`\{(.*)\}`)
)

// extractMissingVarNames takes an input string like "foo {bar} {baz}" and
// returns a list of the var names between {}, e.g. [bar, baz].
func extractMissingVarNames(s string) (vars []string) {
	matches := reVarPlaceholder.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		vars = append(vars, match[1])
	}
	return vars
}

// assertAllVarsReplaced asserts that all vars have been replaced in a string,
// i.e. that there are no values like "{baz}" in the string. It returns an
// error if there is.
func assertAllVarsReplaced(s string) error {
	if strings.ContainsAny(s, "{}") {
		return fmt.Errorf("not all vars replaced in string: %v", s)
	}
	return nil
}

func replaceAllVarsInMapValues(inputMap map[string]string, data interface{}) (map[string]string, error) {
	outputMap := make(map[string]string)
	for key, val := range inputMap {
		replacedValue, err := replaceAllVars(val, data)
		if err != nil {
			return nil, err
		}
		outputMap[key] = replacedValue
	}
	return outputMap, nil
}

// replaceAllVars is the same as replaceVars except returns an error if not
// all variables were replaced.
func replaceAllVars(s string, data interface{}) (string, error) {
	result, err := replaceVars(s, data)
	if err != nil {
		return "", err
	}

	return result, assertAllVarsReplaced(result)
}

func replaceVarsInMapValues(inputMap map[string]string, data interface{}) (map[string]string, error) {
	outputMap := make(map[string]string)
	for key, val := range inputMap {
		replacedValue, err := replaceVars(val, data)
		if err != nil {
			return nil, err
		}
		outputMap[key] = replacedValue
	}
	return outputMap, nil
}

// replaceVars takes a string as a template, executes the template against the
// data provided and returns the result as a string.
func replaceVars(s string, data interface{}) (string, error) {
	template := template.New("")
	if _, err := template.Parse(s); err != nil {
		return "", err
	}
	buffer := &bytes.Buffer{}
	template.Execute(buffer, data)
	return buffer.String(), nil
}
