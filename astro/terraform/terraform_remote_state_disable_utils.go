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
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/uber/astro/astro/logger"

	version "github.com/burl/go-version"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/printer"
)

// astGet gets the node from l at key.
func astGet(l *ast.ObjectList, key string) ast.Node {
	for i := range l.Items {
		for j := range l.Items[i].Keys {
			if l.Items[i].Keys[j].Token.Text == key {
				return l.Items[i].Val
			}
		}
	}
	return nil
}

// astDelIfExists deletes the node at key from l if it exists.
// Returns true if item was deleted.
func astDelIfExists(l *ast.ObjectList, key string) bool {
	for i := range l.Items {
		for j := range l.Items[i].Keys {
			if l.Items[i].Keys[j].Token.Text == key {
				l.Items = append(l.Items[:i], l.Items[i+1:]...)
				return true
			}
		}
	}
	return false
}

func deleteTerraformBackendConfigWithHCL1(in []byte) (updatedConfig []byte, err error) {
	config, err := parseTerraformConfigWithHCL1(in)
	if err != nil {
		return nil, err
	}

	terraformConfigBlock, ok := astGet(config, "terraform").(*ast.ObjectType)
	if !ok {
		return nil, errors.New("could not parse \"terraform\" block in config")
	}

	astDelIfExists(terraformConfigBlock.List, "backend")

	buf := &bytes.Buffer{}
	printer.Fprint(buf, config)

	return buf.Bytes(), nil
}

// hcl2 (used by terraform 0.12) doesn't provide interface to walk through the AST or
// to modify block values, see https://github.com/hashicorp/hcl2/issues/23 and
// https://github.com/hashicorp/hcl2/issues/88
// As a work around we'll perform surgery directly on text, if backend config is simple.
// The method returns an error, if the config is too complicated to be parsed with the regexp.
// This method should be rewritten once hcl2 supports AST traversal and modification.
func deleteTerraformBackendConfigWithHCL2(in []byte) (updatedConfig []byte, err error) {
	// Regexp to find if any backend configuration exists
	backendDefinitionRe := regexp.MustCompile(
		// make sure `\s` matches line breaks
		`(?s)` +
			// match `{backend ` or ` backend `, but not `some_backend` or ` backend_confg`
			`[{\s+]backend\s+` +
			// backend name and opening of the configuration, e.g. `"s3" {`
			`"[^"]+"\s*{`,
	)
	// Regexp to find simple backend configuration, which doesn't contain '{}' inside
	backendBlockRe := regexp.MustCompile(
		// make sure `\s` matches line breaks
		`(?s)` +
			// match backend and it's name, e.g. `backend "s3"` or ` backend "s3"`,
			// note, that opening brace before `backend` is not included in the regex,
			// because it should not be removed.
			`(\s*backend\s+"[^"]+"\s*` +
			// match backend configuration block, that doesn't have inner braces
			`{[^{]*?})`,
	)
	if backendDefinitionRe.Match(in) {
		indexes := backendBlockRe.FindSubmatchIndex(in)
		if indexes == nil {
			return nil, fmt.Errorf("unable to delete backend config: unsupported syntax")
		}
		// Remove found backend submatch from config
		return append(in[:indexes[2]], in[indexes[3]:]...), nil
	}
	return in, nil
}

func deleteTerraformBackendConfig(in []byte, v *version.Version) (updatedConfig []byte, err error) {
	if VersionMatches(v, "<0.12") {
		return deleteTerraformBackendConfigWithHCL1(in)
	}
	return deleteTerraformBackendConfigWithHCL2(in)
}

func deleteTerraformBackendConfigFromFile(file string, v *version.Version) error {
	logger.Trace.Printf("terraform: deleting backend config from %v", file)
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	updatedConfig, err := deleteTerraformBackendConfig(b, v)
	if err != nil {
		return err
	}

	// Unlink the file before writing a new one; this is because we're working
	// with a hardlinked file and we don't want to modify the original.
	os.Remove(file)

	newFile, err := os.Create(file)
	if err != nil {
		return err
	}

	_, err = newFile.Write(updatedConfig)
	if err != nil {
		return err
	}

	newFile.Close()

	return nil
}

func parseTerraformConfigWithHCL1(in []byte) (*ast.ObjectList, error) {
	astFile, err := hcl.ParseBytes(in)
	if err != nil {
		return nil, err
	}

	rootNodes, ok := astFile.Node.(*ast.ObjectList)
	if !ok {
		return nil, errors.New("unable to parse config")
	}

	return rootNodes, nil
}
