package template

import (
	"html/template"
	"os"
	"sort"
	"strings"

	"github.com/kubermatic/gchl/pkg/git"
	"github.com/urfave/cli"
)

// Render renders the changelog as markdown to stdout
func Render(ctx *cli.Context, changelog *git.Changelog) {
	t := template.New("Changelog")
	t.Parse(
		`
### [{{.Version}}]({{.RepositoryURL}})

**Merged pull requests:**

{{range $.Items -}}
- {{.Text}} [#{{.IssueID}}]({{.IssueURL}}) ([{{.Author}}]({{.AuthorURL}}))
{{end}}
`)
	t.Execute(os.Stdout, changelog)
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

// RenderReleaseNotes renders the changelog as markdown to stdout
func RenderReleaseNotes(ctx *cli.Context, changelog *git.Changelog) {
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
	t.Parse(
		`
### [{{.Version}}]({{.RepositoryURL}})

{{range .ChangesGroupedByType}}
**{{ .Type }}:**

{{range .Items -}}
 - {{.Text}} [#{{.IssueID}}]({{.IssueURL}}) ([{{.Author}}]({{.AuthorURL}}))
{{end}}
{{end}}

`)

	t.Execute(os.Stdout, data)
}
