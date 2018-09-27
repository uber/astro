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
	"strings"
	"text/template"
)

func assertAllVarsReplaced(s string) error {
	if strings.ContainsAny(s, "{}") || strings.Contains(s, "<no value>") {
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
