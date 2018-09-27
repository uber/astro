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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/uber/astro/astro/exec2"
	"github.com/uber/astro/astro/logger"
	"github.com/uber/astro/astro/utils"
	version "github.com/burl/go-version"
)

// Session is a wrapper around Terraform commands. It ensures that all
// commands are run within the same working directory.
//
// It also adds "auto init", meaning it will automatically run
// `terraform init` and `terraform get` as necessary before planning or
// applying.
type Session struct {
	id     string
	config *Config

	baseDir    string
	logDir     string
	moduleDir  string
	sandboxDir string

	versionCachedValue *version.Version
}

// NewTerraformSession creates a new Terraform session in the specified
// directory. It will return an error if a previous Terraform session
// was already created here.
func NewTerraformSession(id, baseDir string, config Config) (*Session, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if utils.FileExists(baseDir) {
		return nil, fmt.Errorf("cannot create new session: session already exist at %v", baseDir)
	}

	logDir, err := filepath.Abs(filepath.Join(baseDir, "logs"))
	if err != nil {
		return nil, err
	}

	sandboxDir, err := filepath.Abs(filepath.Join(baseDir, "sandbox"))
	if err != nil {
		return nil, err
	}

	for _, dir := range []string{baseDir, logDir, sandboxDir} {
		logger.Trace.Printf("terraform: mkdir: %v\n", dir)
		if err := os.Mkdir(dir, 0755); err != nil {
			return nil, err
		}
	}

	// Copy the Terraform code tree into the sandbox
	logger.Trace.Printf("terraform: copying tree from %v to %v", config.BasePath, sandboxDir)
	if err := cloneTree(config.BasePath, sandboxDir); err != nil {
		return nil, fmt.Errorf("unable to clone tree from %v to %v: %v", config.BasePath, sandboxDir, err)
	}

	moduleDir, err := filepath.Abs(filepath.Join(sandboxDir, config.ModulePath))
	if err != nil {
		return nil, err
	}

	return &Session{
		id:         id,
		config:     &config,
		baseDir:    baseDir,
		sandboxDir: sandboxDir,
		moduleDir:  moduleDir,
		logDir:     logDir,
	}, nil
}

// command returns an exec2.Process ready to be executed.
func (s *Session) command(logfileName string, cmd string, args []string, expectedSuccessCodes []int) (*exec2.Process, error) {
	env := os.Environ()

	if s.config.SharedPluginDir != "" {
		env = append(env, fmt.Sprintf("TF_PLUGIN_CACHE_DIR=%s", s.config.SharedPluginDir))
	}

	return exec2.NewProcess(exec2.Cmd{
		Command: cmd,
		Args:    args,
		Env:     env,
		CombinedOutputLogFile: filepath.Join(s.logDir, fmt.Sprintf("%s.log", logfileName)),
		ExpectedSuccessCodes:  expectedSuccessCodes,
		WorkingDir:            s.moduleDir,
	}), nil
}

func (s *Session) terraformCommand(args []string, expectedSuccessCodes []int) (*exec2.Process, error) {
	if len(args) < 1 {
		return nil, errors.New("missing args")
	}
	return s.command(args[0], s.config.TerraformPath, args, expectedSuccessCodes)
}

// SetTerraformPath sets the path to Terraform.
func (s *Session) SetTerraformPath(path string) {
	s.config.TerraformPath = path
}

// cloneTree copies the files in existingPath to newPath recursively,
// using hard links.
func cloneTree(existingPath string, newPath string) error {
	existingPathDeref, err := filepath.EvalSymlinks(existingPath)
	if err != nil {
		return err
	}

	newPathDeref, err := filepath.EvalSymlinks(newPath)
	if err != nil {
		return err
	}

	find := exec.Command("find", ".",
		"!", "-path", "*/.terraform/*",
		"!", "-name", ".terraform",
		"!", "-path", "*/.astro/*",
		"!", "-name", ".astro",
		"!", "-name", "terraform.tfstate*",
	)
	find.Dir = existingPathDeref
	cpio := exec.Command("cpio", "-pl", newPathDeref)
	cpio.Dir = existingPathDeref

	cpio.Stdin, err = find.StdoutPipe()
	if err != nil {
		return err
	}

	if err := find.Start(); err != nil {
		return err
	}
	if err := cpio.Start(); err != nil {
		return err
	}
	if err := find.Wait(); err != nil {
		return err
	}
	return cpio.Wait()
}
