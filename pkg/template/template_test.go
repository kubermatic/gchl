/*
Copyright 2020 The Kubermatic Kubernetes Platform contributors.

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

package template

import (
	"testing"

	"k8c.io/gchl/pkg/git"
)

var testDataSet = &git.Changelog{
	Version:       "v1.2.3",
	RepositoryURL: "https://github.com/fooman/repo.git",
	Items: []*git.ChangelogItem{
		{
			Author:     "fooman",
			AuthorURL:  "github.com/fooman",
			Hash:       "bd8647296b6f01cc2e950eeef50e362643f54aa3",
			Text:       "Interesting stuff is now added",
			IssueID:    "66",
			IssueURL:   "https://github.com/fooman/repo/issue/66",
			ChangeType: "Feature",
		},
		{
			Author:     "fooman",
			AuthorURL:  "github.com/fooman",
			Hash:       "e5df5dbef654f91efde06f5afd38fceeeddca47e",
			Text:       "Broken stuff works again",
			IssueID:    "42",
			IssueURL:   "https://github.com/fooman/repo/issue/42",
			ChangeType: "Bugfix",
		},
	},
}

func TestRender(t *testing.T) {
	expectedOutput := `
### [v1.2.3](https://github.com/fooman/repo.git)

**Merged pull requests:**

- Interesting stuff is now added [#66](https://github.com/fooman/repo/issue/66) ([fooman](github.com/fooman))
- Broken stuff works again [#42](https://github.com/fooman/repo/issue/42) ([fooman](github.com/fooman))

`
	result, err := Render(nil, testDataSet)
	if err != nil {
		t.Fatal(err)
	}
	if expectedOutput != result {
		t.Fatalf("expected %q, got %q", expectedOutput, result)
	}
}

func TestRenderReleaseNotes(t *testing.T) {
	expectedOutput := `
### [v1.2.3](https://github.com/fooman/repo.git)


**Bugfix:**

- Broken stuff works again [#42](https://github.com/fooman/repo/issue/42) ([fooman](github.com/fooman))


**Feature:**

- Interesting stuff is now added [#66](https://github.com/fooman/repo/issue/66) ([fooman](github.com/fooman))



`
	result, err := RenderReleaseNotes(nil, testDataSet)
	if err != nil {
		t.Fatal(err)
	}
	if expectedOutput != result {
		t.Fatalf("expected %q, got %q", expectedOutput, result)
	}
}
