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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	// Full path to differ for tests will be stored here on init
	testDifferPath string
)

func init() {
	testDifferPath, _ = which([]string{"diff"})
}

func TestRewriteOutputChange(t *testing.T) {
	if testDifferPath == "" {
		t.Skip("skipping test since there is no diff program")
	}

	inputText := `
module.policies.data.aws_iam_policy_document.billing: Refreshing state...

Your plan was also saved to the path below. Call the "apply" subcommand
with this plan file and Terraform will exactly execute this execution
plan.

Path: mgmt.plan

~ module.policies.aws_iam_policy.billing
policy: "{\n  \"Version\": \"2012-10-17\",\n  \"Statement\": [\n    {\n      \"Effect\": \"Allow\",\n      \"Action\": [\n        \"budgets:*\",\n        \"aws-portal:View*\"\n      ],\n      \"Resource\": [\n        \"*\"\n      ]\n    }\n  ]\n}" => "{\n  \"Version\": \"2012-10-17\",\n  \"Statement\": [\n    {\n      \"Sid\": \"\",\n      \"Effect\": \"Allow\",\n      \"Action\": [\n        \"budgets:*\",\n        \"aws-portal:View*\"\n      ],\n      \"Resource\": \"*\"\n    }\n  ]\n}"

Plan: 0 to add, 1 to change, 0 to destroy.
`
	expectedOutput := `
module.policies.data.aws_iam_policy_document.billing: Refreshing state...

Your plan was also saved to the path below. Call the "apply" subcommand
with this plan file and Terraform will exactly execute this execution
plan.

Path: mgmt.plan

~ module.policies.aws_iam_policy.billing

@@ -6,9 +6,8 @@
         "aws-portal:View*"
       ],
       "Effect": "Allow",
-      "Resource": [
-        "*"
-      ]
+      "Resource": "*",
+      "Sid": ""
     }
   ],
   "Version": "2012-10-17"


Plan: 0 to add, 1 to change, 0 to destroy.
`

	diffedPolicy, err := readableTerraformPolicyChangesWithDiffer(testDifferPath, inputText)

	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(expectedOutput), strings.TrimSpace(diffedPolicy))
}

func TestRewriteOutputAdd(t *testing.T) {
	if testDifferPath == "" {
		t.Skip("skipping test since there is no diff program")
	}

	inputText := `
module.policies.data.aws_iam_policy_document.billing: Refreshing state...

Your plan was also saved to the path below. Call the "apply" subcommand
with this plan file and Terraform will exactly execute this execution
plan.

Path: mgmt.plan

~ module.policies.aws_iam_policy.billing
policy: "{\n  \"Version\": \"2012-10-17\",\n  \"Statement\": [\n    {\n      \"Sid\": \"\",\n      \"Effect\": \"Allow\",\n      \"Action\": [\n        \"budgets:*\",\n        \"aws-portal:View*\"\n      ],\n      \"Resource\": \"*\"\n    }\n  ]\n}"

Plan: 0 to add, 1 to change, 0 to destroy.
`
	expectedOutput := `
module.policies.data.aws_iam_policy_document.billing: Refreshing state...

Your plan was also saved to the path below. Call the "apply" subcommand
with this plan file and Terraform will exactly execute this execution
plan.

Path: mgmt.plan

~ module.policies.aws_iam_policy.billing

@@ -0,0 +1,14 @@
+{
+  "Statement": [
+    {
+      "Action": [
+        "budgets:*",
+        "aws-portal:View*"
+      ],
+      "Effect": "Allow",
+      "Resource": "*",
+      "Sid": ""
+    }
+  ],
+  "Version": "2012-10-17"
+}


Plan: 0 to add, 1 to change, 0 to destroy.
`

	diffedPolicy, err := readableTerraformPolicyChangesWithDiffer(testDifferPath, inputText)

	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(expectedOutput), strings.TrimSpace(diffedPolicy))
}

func TestJSONNormalization(t *testing.T) {
	if testDifferPath == "" {
		t.Skip("skipping test since there is no diff program")
	}

	inputText := `
~ module.policies.aws_iam_policy.billing
policy: "{\n  \"Version\": \"2012-10-17\",\n  \"Statement\": [\n    {\n      \"Effect\": \"Allow\",\n      \"Action\": [\n        \"budgets:*\",\n        \"aws-portal:View*\"\n     ],\n \"Sid\": \"a\",\n     \"Resource\": \"*\"\n    }\n  ]\n}" => "{\n  \"Version\": \"2012-10-17\",\n  \"Statement\": [\n    {\n      \"Sid\": \"b\",\n      \"Resource\": \"*\"\n    ,\n      \"Effect\": \"Allow\",\n      \"Action\": [\n        \"budgets:*\",\n        \"aws-portal:View*\"\n      ]}\n  ]\n}"
`
	expectedOutput := `
~ module.policies.aws_iam_policy.billing

@@ -7,7 +7,7 @@
       ],
       "Effect": "Allow",
       "Resource": "*",
-      "Sid": "a"
+      "Sid": "b"
     }
   ],
   "Version": "2012-10-17"
`

	diffedPolicy, err := readableTerraformPolicyChangesWithDiffer(testDifferPath, inputText)

	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(expectedOutput), strings.TrimSpace(diffedPolicy))
}
