package git

import (
	"context"
	"fmt"
	gh "github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"regexp"
	"strconv"
	"strings"
)

// GithubAPI is used to request the github api
type GithubAPI struct {
	User       string
	Repository string
	UseFilter  bool
	Client     *gh.Client
}

// NewAPIClient returns a new client for interacting with the GitHub API
func NewAPIClient(user string, repo string, token string, filter bool) *GithubAPI {
	return &GithubAPI{
		User:       user,
		Repository: repo,
		UseFilter:  filter,
		Client:     newClient(token),
	}
}

func newClient(token string) *gh.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return gh.NewClient(tc)
}

// CompareRemote matches a list of ChangelogItems with the GitHub API.
// It will use the ChangelogItems IssueID to query the pull request.
// If the issue is a pull request, the values from the github response will be used to create a new ChangelogItem.
func (api *GithubAPI) CompareRemote(items []*ChangelogItem) ([]*ChangelogItem, error) {
	ctx := context.Background()
	var commits []*ChangelogItem

	for _, item := range items {
		id, err := strconv.Atoi(item.IssueID)
		if err != nil {
			return commits, err
		}

		issue, resp, err := api.Client.Issues.Get(ctx, api.User, api.Repository, id)
		if err != nil {
			return commits, err
		}

		if resp.StatusCode != 404 && issue != nil && issue.IsPullRequest() {
			if api.UseFilter {
				if hasReleaseNotes(issue.GetBody()) {
					item.Author = issue.GetUser().GetLogin()
					item.AuthorURL = fmt.Sprintf("https://github.com/%v", issue.GetUser().GetLogin())
					item.Text = filter(issue.GetBody())
					item.IssueURL = fmt.Sprintf("https://github.com/%v/%v/issues/%v", api.User, api.Repository, id)

					commits = append(commits, item)
				}
			} else {
				item.Author = issue.GetUser().GetLogin()
				item.AuthorURL = fmt.Sprintf("https://github.com/%v", issue.GetUser().GetLogin())
				item.Text = issue.GetTitle()
				item.IssueURL = fmt.Sprintf("https://github.com/%v/%v/issues/%v", api.User, api.Repository, id)

				commits = append(commits, item)
			}
		}
	}
	return commits, nil
}

func filter(message string) string {
	body := strings.Replace(message, "```", "___", -1)
	regex := `___release-note(.*\n[\s\S]*?\n)___`
	parser := regexp.MustCompile(regex)

	// get matching group
	text := parser.FindStringSubmatch(body)[1]

	// replace linebreaks
	parser = regexp.MustCompile(`\r?\n`)
	text = parser.ReplaceAllString(text, "")
	return text
}

func hasReleaseNotes(message string) bool {
	body := strings.Replace(message, "```", "___", -1)

	// Special case: check if pr message release notes field with content NONE, return false
	regex := `___release-note(.*\n[NnOoNnEe.*]*?\n)___`
	if matched, _ := regexp.MatchString(regex, body); matched == true {
		return false
	}

	regex = `___release-note(.*\n[\s\S]*?\n)___`
	matched, _ := regexp.MatchString(regex, body)
	return matched
}
