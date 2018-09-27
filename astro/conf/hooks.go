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

import (
	"errors"
)

// Hook holds configuration for user commands that can be executed at various
// stages of the CLI lifecycle.
// Each hook is a shell-like string that will be executed.
// An error in a hook will cause an error in that execution and it will be
// aborted immediately.
// Hooks may optionally output key/value pairs in the form "KEY=VAL" and these
// will be parsed by Colonist and set as environment variables.
type Hook struct {
	// Command is the shell command to be executed
	Command string

	// If set, hook output will be parsed for "KEY=VAL" pairs, which will
	// be set as environment variables
	SetEnv bool `json:"set_env"`
}

// Hooks holds information for shared hooks
type Hooks struct {
	// Startup hooks are executed at CLI startup, after configuration has been
	// validated but before an operation like plan or apply is run.
	Startup []Hook

	// PreModuleRun sets the default for the prehook for a module execution.
	// See the docs on ModuleHooks below.
	PreModuleRun []Hook `json:"pre_module_run"`
}

// ModuleHooks contains configuration for user hooks that should run for a
// given module execution.
type ModuleHooks struct {
	// PreModuleRun hooks are run before a module executes.
	PreModuleRun []Hook `json:"pre_module_run"`
}

// ApplyDefaultsFrom copies the default values from the Hook configuration to
// a ModuleHooks configuration.
func (conf *ModuleHooks) ApplyDefaultsFrom(defaultHooks Hooks) {
	if conf.PreModuleRun == nil {
		conf.PreModuleRun = defaultHooks.PreModuleRun
	}
}

// Validate checks the hook configuration is good
func (hook *Hook) Validate() error {
	if hook.Command == "" {
		return errors.New("Missing hook command")
	}
	return nil
}
