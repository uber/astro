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

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/uber/astro/astro/logger"
)

// NewProcess creates a new process, given the configuration. It does
// not start the process.
func NewProcess(config Cmd) *Process {
	return &Process{config: &config}
}

// Process is a process that has either run or is going to be run.
type Process struct {
	config       *Cmd
	execCmd      *exec.Cmd
	stdoutBuffer *bytes.Buffer
	stderrBuffer *bytes.Buffer
	time         time.Duration
}

func (p *Process) configureOutputs() error {
	p.stdoutBuffer = &bytes.Buffer{}
	p.stderrBuffer = &bytes.Buffer{}

	stdoutWriters := []io.Writer{p.stdoutBuffer}
	stderrWriters := []io.Writer{p.stderrBuffer}

	if p.config.CombinedOutputLogFile != "" {
		combinedOutputLog, err := os.Create(p.config.CombinedOutputLogFile)
		if err != nil {
			return err
		}

		stdoutWriters = append(stdoutWriters, combinedOutputLog)
		stderrWriters = append(stderrWriters, combinedOutputLog)

		fmt.Fprintf(combinedOutputLog, "+ %s %s\n", p.config.Command, p.config.Args)
	}

	p.execCmd.Stdout = io.MultiWriter(stdoutWriters...)
	p.execCmd.Stderr = io.MultiWriter(stderrWriters...)

	return nil
}

// ExitCode returns the exit code for the process. If the process has
// not yet run or exited, the result will be 0.
func (p *Process) ExitCode() int {
	if !p.Exited() {
		return 0
	}

	status, ok := p.execCmd.ProcessState.Sys().(syscall.WaitStatus)
	if !ok {
		return 127
	}

	return status.ExitStatus()
}

// Exited returns whether or not the process has run and exited.
func (p *Process) Exited() bool {
	if p.execCmd == nil || p.execCmd.ProcessState == nil {
		return false
	}
	return p.execCmd.ProcessState.Exited()
}

// Run runs the process.
func (p *Process) Run() error {
	logger.Trace.Printf("exec2: running command: %v; args: %v\n", p.config.Command, p.config.Args)
	p.execCmd = exec.Command(p.config.Command, p.config.Args...)

	// Apply options
	p.execCmd.Dir = p.config.WorkingDir
	p.execCmd.Env = p.config.Env
	p.configureOutputs()

	// If no success codes were given, default to 0
	if p.config.ExpectedSuccessCodes == nil {
		p.config.ExpectedSuccessCodes = []int{0}
	}

	var resultError error
	var resultTime time.Duration

	// Run the process
	started := time.Now()
	if err := p.execCmd.Start(); err != nil {
		logger.Error.Print(err) // Command not found on PATH, not executable, etc.
		resultError = err
		resultTime = time.Since(started)
	} else {
		// wait for the command to finish
		waitCh := make(chan error, 1)
		go func() {
			waitCh <- p.execCmd.Wait()
			close(waitCh)
		}()
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, os.Interrupt)

	Loop:
		for {
			select {
			case sig := <-sigChan:
				process := p.execCmd.Process
				logger.Trace.Printf("Signal: %s, process: %d", sig, process.Pid)
				if err := process.Signal(sig); err != nil {
					// Not clear how we can hit this, but probably not
					// worth terminating the child.
					logger.Error.Print("error sending signal", sig, err)
				}
			case chErr := <-waitCh:
				// Record run time
				resultTime = time.Since(started)

				// Subprocess exited. Get the return code, if we can
				var waitStatus syscall.WaitStatus
				logger.Trace.Printf("exec2: command exit code from Process: %v\n", p.ExitCode())
				if exitError, ok := chErr.(*exec.ExitError); ok {
					waitStatus = exitError.Sys().(syscall.WaitStatus)
					logger.Trace.Printf("exec2: command exit code from chErr: %v\n", waitStatus.ExitStatus())
					// Return an error, if the command didn't exit with a success code
					if !p.Success() {
						resultError = fmt.Errorf("%s%v", p.Stderr().String(), chErr)
					}
				}
				if chErr != nil {
					logger.Error.Print("Channel error", chErr)
				}
				break Loop
			}
		}
	}

	// Record run time
	p.time = resultTime

	// Return an error, if the command didn't exit with a success code
	if !p.Success() || resultError != nil {
		return fmt.Errorf("%s%v", p.Stderr().String(), resultError)
	}

	return nil
}

// Runtime returns the time.Duration the process took to run.
func (p *Process) Runtime() time.Duration {
	return p.time
}

// Stdout returns the contents of the process's stdout.
func (p *Process) Stdout() *bytes.Buffer {
	return p.stdoutBuffer
}

// Stderr returns the contents of the process's stderr.
func (p *Process) Stderr() *bytes.Buffer {
	return p.stderrBuffer
}

// Success returns whether or not the process has exited and if it
// exited with a success code.
func (p *Process) Success() bool {
	if !p.Exited() {
		return false
	}

	exitCode := p.ExitCode()

	for _, c := range p.config.ExpectedSuccessCodes {
		if exitCode == c {
			return true
		}
	}

	return false
}
