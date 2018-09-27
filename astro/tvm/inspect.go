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

package tvm

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/burl/go-version"
)

// InspectVersion will find out what version the Terraform binary at the
// given location is.
func InspectVersion(binaryPath string) (*version.Version, error) {
	stdout, err := exec.Command(binaryPath, "version").Output()
	if err != nil {
		return nil, err
	}

	s := bytes.SplitN(stdout, []byte("\n"), 2)
	if len(s) < 2 {
		return nil, fmt.Errorf("unable to read lines from data: %s", s)
	}

	// e.g. "Terraform v0.7.13"
	versionLine := s[0]

	v := bytes.Split(versionLine, []byte("v"))
	if len(v) != 2 {
		return nil, fmt.Errorf("unable to parse version from data: %s", v)
	}

	return version.NewVersion(string(v[1]))
}
