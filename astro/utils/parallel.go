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

package utils

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Parallel runs at most maxConcurrent functions in parallel.
func Parallel(maxConcurrent int, fns ...func()) {
	wg := sync.WaitGroup{}
	guard := make(chan struct{}, maxConcurrent)

	isShutDown := false
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, os.Interrupt)
	go func() {
		signal := <-signalCh
		fmt.Printf("\nReceived signal: %s, gracefully shutting down...\n", signal)
		isShutDown = true
	}()

	for _, fn := range fns {
		// https://stackoverflow.com/a/25306241
		guard <- struct{}{} // would block if guard channel is already filled
		if isShutDown {
			fmt.Println("Shutting down. No new terraform processes will be launched.")
			break
		} else {
			wg.Add(1)
			go func(fn func()) {
				defer wg.Done()
				fn()
				<-guard
			}(fn)
		}
	}

	wg.Wait()
}
