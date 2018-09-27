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
	"strings"
	"time"

	"github.com/uber/astro/astro/exec2"
)

// Result is a generic interface that satisfies types returned
// by Terraform methods.
type Result interface {
	Runtime() string
	Stdout() string
	Stderr() string
}

// terraformResult is returned by the Plan/Apply commands.
type terraformResult struct {
	process *exec2.Process
}

// Runtime returns a human readable string with how long it took to run
// the command.
func (r *terraformResult) Runtime() string {
	return r.process.Runtime().Truncate(time.Second).String()
}

// Stdout returns the stdout for this execution.
func (r *terraformResult) Stdout() string {
	return r.process.Stdout().String()
}

// Stderr returns the stderr for this execution.
func (r *terraformResult) Stderr() string {
	return r.process.Stderr().String()
}

// PlanResult is the terraformResult of a Terraform plan.
type PlanResult struct {
	*terraformResult

	changes string
}

// Changes returns the changes for this plan.
func (r *PlanResult) Changes() string {
	return strings.TrimSpace(r.changes)
}

// HasChanges returns whether this plan had changes or not.
func (r *PlanResult) HasChanges() bool {
	return r.process.ExitCode() == 2
}
