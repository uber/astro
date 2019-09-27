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

// Tests that backend part can be successfully removed from the config
// written in HCL 1.0 language
func TestDeleteTerraformBackendConfigWithHCL1(t *testing.T) {
	input := []byte(`
terraform {
    backend "s3" {}
    }

    provider "aws" {
      region = "${var.aws_region}"
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

	updatedConfig, err := deleteTerraformBackendConfigWithHCL1(input)
	assert.NoError(t, err)

	assert.Equal(t, `terraform {}

provider "aws" {
  region = "${var.aws_region}"
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

// Tests that backend part can be successfully removed from the config
// written in HCL 2.0 language
func TestDeleteTerraformBackendConfigWithHCL2Success(t *testing.T) {
	tests := []struct {
		config   string
		expected string
	}{
		{
			config: `
				provider "aws"{
					region = var.aws_region
				}`,
			expected: `
				provider "aws"{
					region = var.aws_region
				}`,
		},
		{
			config: `
				terraform {
					version = "v0.12.6"
					backend "local" {
						path = "path"
					}
					key = "value"
				}`,
			expected: `
				terraform {
					version = "v0.12.6"
					key = "value"
				}`,
		},
		{
			config: `
				terraform {backend "s3" {}}

				provider "aws" {
					region = "us-east-1"
				}`,
			expected: `
				terraform {}

				provider "aws" {
					region = "us-east-1"
				}`,
		},
	}
	for _, tt := range tests {
		actual, err := deleteTerraformBackendConfigWithHCL2([]byte(tt.config))
		assert.Equal(t, string(actual), tt.expected)
		assert.Nil(t, err)
	}
}

// Tests that trying to delete backend part from configs where
// backend secions contains parenthesis fails. See comment on
// deleteTerraformBackendConfigWithHCL2 for clarification.
func TestDeleteTerraformBackendConfigWithHCL2Failure(t *testing.T) {
	tests := []struct {
		config string
	}{
		{
			config: `
			terraform {
				backend "local" {
					path = "module-{{.environment}}"
				}
			}`,
		},
		{
			config: `
			terraform {
				backend "concil" {
					map = {"key": "val"}
				}
			}`,
		},
	}

	for _, tt := range tests {
		_, err := deleteTerraformBackendConfigWithHCL2([]byte(tt.config))
		assert.NotNil(t, err)
	}
}
