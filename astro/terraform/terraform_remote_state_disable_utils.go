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
	"io/ioutil"
	"os"

	"github.com/uber/astro/astro/logger"

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

// astDel deletes the node at key from l. Returns an error if the key does not
// exist.
func astDel(l *ast.ObjectList, key string) error {
	for i := range l.Items {
		for j := range l.Items[i].Keys {
			if l.Items[i].Keys[j].Token.Text == key {
				l.Items = append(l.Items[:i], l.Items[i+1:]...)
				return nil
			}
		}
	}
	return errors.New("cannot delete key %v: does not exist")
}

func deleteTerraformBackendConfig(in []byte) (updatedConfig []byte, err error) {
	config, err := parseTerraformConfig(in)
	if err != nil {
		return nil, err
	}

	terraformConfigBlock, ok := astGet(config, "terraform").(*ast.ObjectType)
	if !ok {
		return nil, errors.New("could not parse \"terraform\" block in config")
	}

	if err := astDel(terraformConfigBlock.List, "backend"); err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	printer.Fprint(buf, config)

	return buf.Bytes(), nil
}

func deleteTerraformBackendConfigFromFile(file string) error {
	logger.Trace.Printf("terraform: deleting backend config from %v", file)
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	updatedConfig, err := deleteTerraformBackendConfig(b)
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

func parseTerraformConfig(in []byte) (*ast.ObjectList, error) {
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
