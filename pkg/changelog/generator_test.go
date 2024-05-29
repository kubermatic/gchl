/*
Copyright 2022 The Kubermatic Kubernetes Platform contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package changelog

import (
	"os"
	"path/filepath"
	"testing"

	"k8c.io/gchl/pkg/types"

	"gopkg.in/yaml.v3"
)

// func TestHumanReadableChangeType(t *testing.T) {
// 	testcases := []struct {
// 		identifier string
// 		expected   string
// 	}{
// 		{
// 			identifier: "breaking change",
// 			expected:   "Breaking Change",
// 		},
// 		{
// 			identifier: "api change",
// 			expected:   "API Change",
// 		},
// 		{
// 			identifier: "api-change",
// 			expected:   "API Changes",
// 		},
// 	}

// 	for _, testcase := range testcases {
// 		t.Run(testcase.identifier, func(t *testing.T) {
// 			result := humanReadableChangeType(testcase.identifier)

// 			if result != testcase.expected {
// 				t.Fatalf("Expected %q, got %q.", testcase.expected, result)
// 			}
// 		})
// 	}
// }

type generateChangesTestcase struct {
	PR      types.PullRequest `yaml:"pr"`
	Changes []Change          `yaml:"changes"`
}

func TestGenerateChanges(t *testing.T) {
	files, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatalf("Failed to load testcases: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		t.Run(filepath.Base(file.Name()), func(t *testing.T) {
			content, err := os.ReadFile("testdata/" + file.Name())
			if err != nil {
				t.Fatalf("Failed to load testcase: %v", err)
			}

			testcase := generateChangesTestcase{}
			if err := yaml.Unmarshal(content, &testcase); err != nil {
				t.Fatalf("Failed to load testcase: %v", err)
			}

			changes, err := processCommit(types.Commit{
				PullRequest: testcase.PR,
			})
			if err != nil {
				t.Fatalf("Failed to generate changes: %v", err)
			}

			if len(changes) != len(testcase.Changes) {
				t.Fatalf("Expected %d changes, got %d.", len(testcase.Changes), len(changes))
			}

			for i, change := range changes {
				expectedChange := testcase.Changes[i]

				if expectedChange.Type != change.Type {
					t.Errorf("change #%d: expected type %q, got %q", i, expectedChange.Type, change.Type)
				}

				if expectedChange.Text != change.Text {
					t.Errorf("change #%d: expected release note %q, got %q", i, expectedChange.Text, change.Text)
				}
			}
		})
	}
}
