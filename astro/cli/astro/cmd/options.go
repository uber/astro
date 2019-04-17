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

package cmd

import (
	"io"

	multierror "github.com/hashicorp/go-multierror"
)

// Option is an option for the cli that allows for changing of options or
// dependency injection for testing.
type Option func(*AstroCLI) error

func (cli *AstroCLI) applyOptions(opts ...Option) (errs error) {
	for _, opt := range opts {
		if err := opt(cli); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

// WithStdout allows you to pass custom input.
func WithStdout(stdout io.Writer) Option {
	return func(cli *AstroCLI) error {
		cli.stdout = stdout
		return nil
	}
}

// WithStderr allows you to pass a custom stderr.
func WithStderr(stderr io.Writer) Option {
	return func(cli *AstroCLI) error {
		cli.stderr = stderr
		return nil
	}
}

// WithStdin allows you to pass a custom stdin.
func WithStdin(stdin io.Reader) Option {
	return func(cli *AstroCLI) error {
		cli.stdin = stdin
		return nil
	}
}
