/*
Copyright 2024 The Kubermatic Kubernetes Platform contributors.

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
	"testing"
)

func TestIsUpdate(t *testing.T) {
	testcases := []struct {
		text     string
		expected bool
	}{
		{
			text:     "Update OSM to v1.5.2; fixing cloud-init bootstrapping issues on Ubuntu 22.04 on Azure",
			expected: true,
		},
		{
			text:     "updates operating-system-manager to v1.5.1.",
			expected: true,
		},
		{
			text:     "update Metering to v1.2.1.",
			expected: true,
		},
		{
			text:     "Updated to Go 1.22.2",
			expected: true,
		},
		{
			text:     "Update Cilium to 1.14.9 and 1.13.14",
			expected: true,
		},
		{
			text:     "updating to Kubernetes 1.29",
			expected: true,
		},
		{
			text:     "Update Vertical Pod Autoscaler to 1.0",
			expected: true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.text, func(t *testing.T) {
			result := isUpdate(testcase.text)

			if result != testcase.expected {
				t.Fatalf("Expected %v, got %v.", testcase.expected, result)
			}
		})
	}
}

func TestRemoveActionRequired(t *testing.T) {
	testcases := []struct {
		input    string
		expected string
	}{
		{
			input:    "Action required: if you use `velero.restic.deploy: true`...",
			expected: "if you use `velero.restic.deploy: true`...",
		},
		{
			input:    "**ACTION REQUIRED**: For velero helm chart upgrade. If running...",
			expected: "For velero helm chart upgrade. If running...",
		},
		{
			input:    "Action required: [User-mla] If you had copied `values.yaml...",
			expected: "[User-mla] If you had copied `values.yaml...",
		},
		{
			input:    "[ACTION REQUIRED] KubeLB: The prefix for the tenant namespaces created...",
			expected: "KubeLB: The prefix for the tenant namespaces created...",
		},
		{
			input:    "[Action Required] The field `ovdcNetwork` in `cluster` and `preset...",
			expected: "The field `ovdcNetwork` in `cluster` and `preset...",
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.input, func(t *testing.T) {
			result := removeActionRequired(testcase.input)

			if result != testcase.expected {
				t.Fatalf("Expected %q, got %q.", testcase.expected, result)
			}
		})
	}
}
