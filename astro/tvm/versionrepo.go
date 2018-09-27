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

// Package tvm stands for Terraform version manager. It will
// automatically download and manage multiple Terraform binaries.
package tvm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/uber/astro/astro/utils"

	homedir "github.com/mitchellh/go-homedir"
)

// terraformBinaryFile is the name of the Terraform binary.
const terraformBinaryFile = "terraform"

// terraformZipFileDownloadURL is the path to download Terraform zip
// files from the Hashicorp website.
var terraformZipFileDownloadURL = "https://releases.hashicorp.com/terraform/%s/terraform_%s_%s_%s.zip"

// VersionRepo is a directory on the filesystem that keeps
// Terraform binaries.
type VersionRepo struct {
	repoPath string
	arch     string
	platform string

	downloadLock *sync.Map
}

// NewVersionRepo creates a new VersionRepo. The arch will
// be appended to the provided path for all downloaded binaries.
func NewVersionRepo(repoPath string, arch string, platform string) (*VersionRepo, error) {
	if repoPath == "" {
		home, err := homedir.Dir()
		if err != nil {
			return nil, err
		}

		repoPath = filepath.Join(home, ".tvm")
	}

	if !utils.FileExists(repoPath) {
		if err := os.Mkdir(repoPath, 0755); err != nil {
			return nil, err
		}
	}
	return &VersionRepo{
		downloadLock: &sync.Map{},
		repoPath:     repoPath,
		arch:         arch,
		platform:     platform,
	}, nil
}

// NewVersionRepoForCurrentSystem returns a new VersionRepo instance
// with platform and architecture information retrieve from the current
// system.
func NewVersionRepoForCurrentSystem(repoPath string) (*VersionRepo, error) {
	return NewVersionRepo(repoPath, runtime.GOARCH, runtime.GOOS)
}

// dir returns the directory in the repository that contains the
// specified version.
func (r *VersionRepo) dir(version string) string {
	return filepath.Join(r.repoPath, r.platform, r.arch, version)
}

// download gets the Terraform binary from the Terraform website. It
// returns the path to the downloaded file or an error if there was a
// problem.
func (r *VersionRepo) download(version string) (string, error) {
	downloadLock := r.getLock(version)

	downloadLock.Lock()
	defer downloadLock.Unlock()

	url := fmt.Sprintf(terraformZipFileDownloadURL, version, version, r.platform, r.arch)

	// Temporary file for zip file
	tmpfile, err := ioutil.TempFile("", "terraform")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpfile.Name())

	// Download zip file
	if err := downloadFile(url, tmpfile.Name()); err != nil {
		return "", err
	}

	targetDir := r.dir(version)

	// Extract contents of zip file to repo path
	if err := unzip(tmpfile.Name(), targetDir); err != nil {
		return "", err
	}

	// Check the binary is there
	if !r.exists(version) {
		return "", errors.New("Terraform binary missing from zip file")
	}

	return r.terraformPath(version), nil
}

// exists returns whether or not the binary for the specified version
// exists.
func (r *VersionRepo) exists(version string) bool {
	return utils.FileExists(r.terraformPath(version))
}

func (r *VersionRepo) getLock(version string) *sync.Mutex {
	v, _ := r.downloadLock.LoadOrStore(version, &sync.Mutex{})
	return v.(*sync.Mutex)
}

// Get takes a version and returns the path to the Terraform binary for
// that version. If the binary doesn't exist, it will be downloaded from
// the Terraform website automatically.
func (r *VersionRepo) Get(version string) (string, error) {
	path := r.terraformPath(version)
	if !utils.FileExists(path) {
		return r.download(version)
	}
	return path, nil
}

// Link symlinks the version binary into the targetPath. It will
// download the binary if the version does not exist in the repository.
func (r *VersionRepo) Link(version string, targetPath string, overwrite bool) error {
	terraformPath, err := r.Get(version)
	if err != nil {
		return err
	}

	if overwrite {
		_, err := os.Lstat(targetPath)
		if !os.IsNotExist(err) {
			os.Remove(targetPath)
		}
	}

	return os.Symlink(terraformPath, targetPath)
}

// terraformPath returns the path to the Terraform binary file with the
// specified version.
func (r *VersionRepo) terraformPath(version string) string {
	return filepath.Join(r.dir(version), terraformBinaryFile)
}
