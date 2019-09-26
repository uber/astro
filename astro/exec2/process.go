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

	"github.com/hashicorp/go-multierror"
)

// a flag to indicate that we caught an interrupt signal
// so no new processes will be launched
var isInterrupted = false

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

// Process returns the Process field of underlying exec command
// This allows us to interact with it, i.e. for sending signals
func (p *Process) Process() *os.Process {
	return p.execCmd.Process
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
	command := p.config.Command
	args := p.config.Args

	if isInterrupted {
		return fmt.Errorf("astro was interrupted, command won't be run: %s, args: %v", command, args)
	}

	logger.Trace.Printf("exec2: running command: %v; args: %v\n", command, args)
	p.execCmd = exec.Command(command, args...)

	// Apply options
	p.execCmd.Dir = p.config.WorkingDir
	p.execCmd.Env = p.config.Env
	p.configureOutputs()

	// If no success codes were given, default to 0
	if p.config.ExpectedSuccessCodes == nil {
		p.config.ExpectedSuccessCodes = []int{0}
	}

	// Run the process
	started := time.Now()
	if err := p.execCmd.Start(); err != nil {
		p.time = time.Since(started)
		return err
	} else {
		// wait for the command to finish
		waitCh := make(chan error, 1)
		go func() {
			waitCh <- p.execCmd.Wait()
			close(waitCh)
		}()
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

		var errors error
		for {
			select {
			case sig := <-sigChan:
				isInterrupted = true
				errors = multierror.Append(fmt.Errorf("signal received: %s", sig))
				process := p.execCmd.Process
				logger.Trace.Printf("Signal: %s, process: %d\n", sig, process.Pid)
				if err := process.Signal(sig); err != nil {
					errors = multierror.Append(errors, err)
				}
			case err := <-waitCh:
				// Record run time
				p.time = time.Since(started)
				logger.Trace.Printf("exec2: command exit code: %v\n", p.ExitCode())
				// Return an error, if the command didn't exit with a success code
				if !p.Success() {
					errors = multierror.Append(errors, err)
					return fmt.Errorf("%s%v", p.Stderr().String(), errors)
				}
				return errors
			}
		}
	}
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
