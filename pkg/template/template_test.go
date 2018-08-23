package template

import (
	"testing"

	"github.com/kubermatic/gchl/pkg/git"
)

var testDataSet = &git.Changelog{
	Version:       "v1.2.3",
	RepositoryURL: "https://github.com/fooman/repo.git",
	Items: []*git.ChangelogItem{
		&git.ChangelogItem{
			Author:     "fooman",
			AuthorURL:  "github.com/fooman",
			Hash:       "bd8647296b6f01cc2e950eeef50e362643f54aa3",
			Text:       "Interesting stuff is now added",
			IssueID:    "66",
			IssueURL:   "https://github.com/fooman/repo/issue/66",
			ChangeType: "Feature",
		},
		&git.ChangelogItem{
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
