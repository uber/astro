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
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/uber/astro/astro/conf"
	"github.com/uber/astro/astro/logger"

	"github.com/kballard/go-shellquote"
)

// runCommandkAndSetEnvironment runs the specified hook/command.
//
// If parseEnvironment is true, output in the format "KEY=VAL" for
// hooks is insert into the current process's environment. An error is returned
// if the hook fails to execute.
func runCommandkAndSetEnvironment(workingDir string, hook conf.Hook) error {
	logger.Trace.Printf("astro: running hook: %v", hook.Command)

	args, err := shellquote.Split(hook.Command)
	if err != nil {
		return err
	}

	prog, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}

	output := &bytes.Buffer{}

	cmd := exec.Command(prog, args[1:]...)
	cmd.Dir = workingDir

	// Have to pipe through stderr and stdin so that scripts that prompt, e.g.
	// for MFA will work.
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		return err
	}

	if hook.SetEnv {
		if err := parseOutputIntoEnv(output); err != nil {
			return fmt.Errorf("unable to set env var from hook output: %v", err)
		}
	}

	return nil
}

// parseOutputIntoEnv takes stdout of a hook and reads for lines in the format
// "KEY=VAL". If then sets those as environment variables. It stops processing
// on the first line that doesn't match this format.
func parseOutputIntoEnv(buf *bytes.Buffer) error {
	scanner := bufio.NewScanner(buf)

	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "=", 2)
		if len(parts) != 2 {
			// abort processing output on first non-conforming line
			return nil
		}

		if err := os.Setenv(parts[0], parts[1]); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error parsing hook output: %v", err)
	}

	return nil
}
