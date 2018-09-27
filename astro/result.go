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

import "github.com/uber/astro/astro/terraform"

// Result is what is returned from astro execution.
type Result struct {
	id              string
	terraformResult terraform.Result
	err             error
}

// ID is a unique name that identifies the execution that run.
func (r *Result) ID() string {
	return r.id
}

// TerraformResult is the result of the Terraform command, or nil if
// there wasn't one.
func (r *Result) TerraformResult() terraform.Result {
	return r.terraformResult
}

// Err returns the error of the execution, if there was one.
func (r *Result) Err() error {
	return r.err
}
