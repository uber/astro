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

package tvm

import (
	"archive/zip"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// downloadFile will download the specified file to the specified path.
func downloadFile(url string, path string) error {
	// Create file
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func unzip(zipfilePath string, destDir string) error {
	r, err := zip.OpenReader(zipfilePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Get file data
		fh, err := f.Open()
		if err != nil {
			return err
		}
		defer fh.Close()

		path := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			// Directory
			os.MkdirAll(path, os.ModePerm)
		} else {
			// File
			if err = os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
				return err
			}

			out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}

			_, err = io.Copy(out, fh)

			out.Close()

			if err != nil {
				return err
			}
		}
	}
	return nil
}
