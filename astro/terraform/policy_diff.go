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

package terraform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	multierror "github.com/hashicorp/go-multierror"
)

var (
	// Full path to differ will be stored here on init
	differPath string
	// $PATH will be searched for these tools on init
	differTools = []string{
		"colordiff",
		"diff",
	}
	newline = []byte("\n")
	// regular expressions that matches a policy add/change in a Terraform diff.
	terraformPolicyAddLine    = regexp.MustCompile(`^(.*\b(?:assume_role_policy|policy):\s+)"(.*)"`)
	terraformPolicyChangeLine = regexp.MustCompile(`^(.*\b(?:assume_role_policy|policy):\s+)"(.*)" => "(.*)"`)
)

func init() {
	differPath, _ = which(differTools)
}

// terraformPolicyChangeToDiff takes a Terraform policy change output line
// (i.e. from a Terraform plan) parses the JSON and outputs a unified diff.
func terraformPolicyChangeToDiff(differ, policyBefore, policyAfter string) ([]byte, error) {
	jsonBefore, err := jsonPretty(unescape(policyBefore))
	if err != nil {
		return nil, err
	}
	before, err := writeToTempFile(jsonBefore)
	if err != nil {
		return nil, err
	}
	defer os.Remove(before)

	jsonAfter, err := jsonPretty(unescape(policyAfter))
	if err != nil {
		return nil, err
	}
	after, err := writeToTempFile(jsonAfter)
	if err != nil {
		return nil, err
	}
	defer os.Remove(after)

	return diff(differ, before, after)
}

// diff invokes diff to output a diff of two files.
func diff(differ, file1, file2 string) ([]byte, error) {
	cmd := exec.Command(differ, "-u", file1, file2)
	out, err := cmd.Output()

	// We only want to throw an error here if the exit status was 2 or
	// higher. From the diff man page: "Exit status is 0 if inputs are the
	// same, 1 if different, 2 if trouble."
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			return nil, err
		}

		status, ok := exitErr.Sys().(syscall.WaitStatus)
		if !ok || status.ExitStatus() > 1 {
			return nil, err
		}
	}

	return out, nil
}

// jsonPretty takes unformatted JSON and indents it so it is human readable. If
// the JSON cannot be indented, the original JSON is returned.
func jsonPretty(in []byte) ([]byte, error) {
	if len(in) == 0 {
		return in, nil
	}
	var unmarshalled interface{}
	err := json.Unmarshal(in, &unmarshalled)
	if err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(unmarshalled, "", "  ")
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CanDisplayReadableTerraformPolicyChanges is true when the prerequisites for
// ReadableTerraformPolicyChanges are fulfilled
func CanDisplayReadableTerraformPolicyChanges() bool {
	return differPath != ""
}

func readableTerraformPolicyChangesWithDiffer(differ, terraformChanges string) (string, error) {
	result := ""
	var errs error
	for _, line := range strings.Split(terraformChanges, "\n") {
		// Check if the line matches a Terraform policy diff
		changeGroups := terraformPolicyChangeLine.FindStringSubmatch(line)
		addGroups := terraformPolicyAddLine.FindStringSubmatch(line)
		if changeGroups == nil && addGroups == nil {
			// If it doesn't match, just print the line verbatim and move on
			result += line
			result += "\n"
			continue
		}

		// Get a readable diff from the policy change
		var difftext []byte
		var err error
		var fieldName string
		if changeGroups != nil {
			fieldName = changeGroups[1]
			difftext, err = terraformPolicyChangeToDiff(differ, changeGroups[2], changeGroups[3])
		} else {
			fieldName = addGroups[1]
			difftext, err = terraformPolicyChangeToDiff(differ, "", addGroups[2])
		}
		if err != nil {
			errs = multierror.Append(errs, err)
			result += line
			result += "\n"
			continue
		}

		// Output a readable diff
		if len(difftext) > 0 {
			result += fieldName
			result += "\n"
			result += string(tail(difftext, 2, true))
			result += "\n"
		} else {
			result += fieldName
			result += "<no changes after normalization>\n"
		}
	}

	return result, errs
}

// ReadableTerraformPolicyChanges takes the output of `terraform plan` and
// rewrites policy diff to be in unified diff format
func ReadableTerraformPolicyChanges(terraformChanges string) (string, error) {
	return readableTerraformPolicyChangesWithDiffer(differPath, terraformChanges)
}

// tail is an implementation of the unix tail command. If fromN is true, it is
// equivalent to `tail -n +K`. See `main tail` for more info.
func tail(input []byte, n int, fromN bool) []byte {
	// split lines
	sub := bytes.Split(input, newline)
	if fromN {
		return bytes.Join(sub[n:], newline)
	}
	return bytes.Join(sub[len(sub)-n:], newline)
}

// unescape takes an escaped JSON string output by Terraform on the console
// and converts it to valid JSON.
func unescape(in string) []byte {
	out := []byte(in)
	out = bytes.Replace(out, []byte(`\n`), []byte("\n"), -1)
	out = bytes.Replace(out, []byte(`\"`), []byte(`"`), -1)
	out = bytes.Replace(out, []byte(`\\`), []byte(`\`), -1)
	return out
}

// writeToTempFile creates a temporary file and writes the specified data to
// it.
func writeToTempFile(data []byte) (filePath string, err error) {
	tmpfile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}

	if len(data) > 0 {
		tmpfile.Write(data)
		tmpfile.Write(newline)
	}

	return tmpfile.Name(), nil
}

// which searches the $PATH for each of the candidates and returns the full
// path to the first program that exists.
func which(candidates []string) (string, error) {
	for _, candidate := range candidates {
		path, err := exec.LookPath(candidate)
		if err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("cannot find any of: %v in $PATH", candidates)
}
