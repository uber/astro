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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteTerraformBackendConfig(t *testing.T) {
	input := []byte(`
terraform {
    backend "s3" {}
    }

    provider "aws" {
    region = "us-east-1"
    }

    module "codecommit" {
    source = "../../modules/codecommit"

    rw_roles = [
        "sre",
    ]
    ro_roles = [
        "dev",
        "engsec",
    ]
}`)

	updatedConfig, err := deleteTerraformBackendConfig(input)
	assert.NoError(t, err)

	assert.Equal(t, `terraform {}

provider "aws" {
  region = "us-east-1"
}

module "codecommit" {
  source = "../../modules/codecommit"

  rw_roles = [
    "sre",
  ]

  ro_roles = [
    "dev",
    "engsec",
  ]
}`, string(updatedConfig))
}
