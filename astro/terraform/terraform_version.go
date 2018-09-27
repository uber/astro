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
	"fmt"

	"github.com/uber/astro/astro/tvm"

	version "github.com/burl/go-version"
)

// Version returns the version of Terraform that the binary identifies
// itself as.
func (s *Session) Version() (*version.Version, error) {
	return tvm.InspectVersion(s.config.TerraformPath)
}

func (s *Session) versionCached() (*version.Version, error) {
	if s.versionCachedValue == nil {
		v, err := s.Version()
		if err != nil {
			return nil, fmt.Errorf("unable to detect Terraform version: %v", err)
		}

		s.versionCachedValue = v
	}
	return s.versionCachedValue, nil
}
