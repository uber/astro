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

package exec2

// Cmd is the configuration struct for a process.
type Cmd struct {
	// Args is a list of arguments to provide to the process.
	Args []string
	// CombinedOutputLogFile is the path to a file where the process's
	// stdout and stderr should be logged.
	CombinedOutputLogFile string
	// Command is the path to the process that you want to run
	Command string
	// Environment variables to use. If empty, set to current process's env.
	Env []string
	// ExpectedSuccessCodes is a list of exit codes the process will return if
	// it completes successfully.
	ExpectedSuccessCodes []int
	// WorkingDir is the working directory of the process.
	WorkingDir string
}
