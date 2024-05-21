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
)

type markdown struct{}

func NewMarkdownRenderer() Renderer {
	return &markdown{}
}

var markdownTemplate = `
## {{ .Version }}

**GitHub release: [{{ .Version }}]({{ .RepositoryURL }}/releases/tag/{{ .Version }})**
{{ range .ChangeGroups }}
### {{ .Title }}
{{ range .Changes }}
- {{ .ReleaseNote }} ([#{{ .Commit.PullRequest.Number }}]({{ prlink .Commit.PullRequest.Number }}))
{{- end }}
{{ end }}
`

func (m *markdown) Render(log *changelog.Changelog) (string, error) {
	t := template.New("changelog").Funcs(template.FuncMap{
		"prlink": func(number int) string {
			return fmt.Sprintf("%s/pull/%d", log.RepositoryURL, number)
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
