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
	"bytes"
	"sort"
	"strings"
	"text/template"

	"github.com/urfave/cli"

	"k8c.io/gchl/pkg/git"
)

// Render renders the changelog as markdown to stdout
func Render(ctx *cli.Context, changelog *git.Changelog) (string, error) {
	t := template.New("Changelog")

	_, err := t.Parse(
		`
## [{{ .Version }}]({{ .RepositoryURL }}/releases/tag/{{ .Version }})

{{ range $.Items -}}
- {{ .Text }} ([#{{ .IssueID }}]({{ .IssueURL }}))
{{ end }}
`)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	err = t.Execute(&b, changelog)
	return b.String(), err
}

// RenderReleaseNotes renders the changelog as markdown to stdout
func RenderReleaseNotes(ctx *cli.Context, changelog *git.Changelog) (string, error) {
	changesGroupedByType := groupChangesByType(changelog.Items)

	data := struct {
		ChangesGroupedByType []changesByType
		Version              string
		RepositoryURL        string
	}{
		ChangesGroupedByType: changesGroupedByType,
		Version:              changelog.Version,
		RepositoryURL:        changelog.RepositoryURL,
	}

	t := template.New("Changelog")
	_, err := t.Parse(strings.TrimSpace(`
## [{{ .Version }}]({{ .RepositoryURL }}/releases/tag/{{ .Version }})
{{ range .ChangesGroupedByType }}
### {{ .Type }}
{{ range .Items }}
- {{ .Text }} ([#{{ .IssueID }}]({{ .IssueURL }}))
{{- end }}
{{ end }}
`))
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	err = t.Execute(&b, data)
	return b.String(), err
}

type changesByType struct {
	Type  string
	Items []*git.ChangelogItem
}

func groupChangesByType(items []*git.ChangelogItem) []changesByType {
	// create a map of change types to change slices
	m := make(map[string][]*git.ChangelogItem)
	for _, item := range items {
		if _, ok := m[item.ChangeType]; !ok {
			m[item.ChangeType] = []*git.ChangelogItem{}
		}

		m[item.ChangeType] = append(m[item.ChangeType], item)
	}

	// Get a sorted list of change types. Sorting alphabetically might not be
	// the optimal way, but it prevents the order of entries from constantly changing.
	var sortedTypes []string
	for key := range m {
		// capitalize the type
		sortedTypes = append(sortedTypes, key)
	}
	sort.Strings(sortedTypes)

	// We cannot use the list of groups and the map separately in the template
	// (because `range` mutates `.`), so we need to have both the group names
	// and items in one sorted array.
	var groupedChanges []changesByType
	for _, t := range sortedTypes {
		groupedChanges = append(groupedChanges,
			changesByType{
				Type:  strings.Title(t),
				Items: m[t],
			})
	}

	return groupedChanges
}
