/*
Copyright 2017 The Kubernetes Authors.

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

package render

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"k8c.io/gchl/pkg/changelog"

	"github.com/go-openapi/inflect"
)

type markdown struct{}

func NewMarkdownRenderer() Renderer {
	return &markdown{}
}

var markdownTemplate = `
## {{ .Version }}

**GitHub release: [{{ .Version }}]({{ .RepositoryURL }}/releases/tag/{{ .Version }})**
{{- $breaking := .BreakingChanges }}
{{- if $breaking }}

### Breaking Changes

This release contains changes that require additional attention, please read the following items carefully.
{{ range $breaking }}
- {{ .Text }} ([#{{ .Commit.PullRequest.Number }}]({{ prlink .Commit.PullRequest.Number }}))
{{- end }}
{{- end }}
{{ range .ChangeGroups }}
### {{ typename .Type }}
{{ range .Changes }}
- {{ .Text }} ([#{{ .Commit.PullRequest.Number }}]({{ prlink .Commit.PullRequest.Number }}))
{{- end }}
{{ end }}
`

var overriddenTypeNames = map[changelog.ChangeType]string{
	changelog.ChangeTypeAPIChange:     "API Changes",
	changelog.ChangeTypeBugfix:        "Bugfixes",
	changelog.ChangeTypeCleanup:       "Cleanups",
	changelog.ChangeTypeDeprecation:   "Deprecations",
	changelog.ChangeTypeDocumentation: "Documentation",
	changelog.ChangeTypeFeature:       "New Features",
	changelog.ChangeTypeMisc:          "Miscellaneous",
	changelog.ChangeTypeChore:         "Chores",
	changelog.ChangeTypeRegression:    "Regresssions",
	changelog.ChangeTypeUpdate:        "Updates",
}

func (m *markdown) Render(log *changelog.Changelog) (string, error) {
	t := template.New("changelog").Funcs(template.FuncMap{
		"prlink": func(number int) string {
			return fmt.Sprintf("%s/pull/%d", log.RepositoryURL, number)
		},
		"typename": func(changeType changelog.ChangeType) string {
			if known, ok := overriddenTypeNames[changeType]; ok {
				return known
			}

			title := strings.ReplaceAll(string(changeType), "-", " ")
			title = inflect.Titleize(title)
			title = strings.ReplaceAll(title, "Api", "API")

			return title
		},
	})

	var err error
	t, err = t.Parse(strings.TrimSpace(markdownTemplate))
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	err = t.Execute(&b, log)

	return b.String(), err
}
