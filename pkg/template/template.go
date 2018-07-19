package template

import (
	"github.com/kubermatic/gchl/pkg/git"
	"github.com/urfave/cli"
	"html/template"
	"os"
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
