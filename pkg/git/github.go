package git

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	gh "github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// FilterKind describes which kind of filter should be applied on the issue's release notes
type FilterKind int

const (
	FilterNone FilterKind = iota
	FilterReleaseNotes
	FilterReleaseNotesNone
)

// GithubAPI is used to request the github api
type GithubAPI struct {
	User       string
	Repository string
	Filter     FilterKind
	Client     *gh.Client
}

// NewAPIClient returns a new client for interacting with the GitHub API
func NewAPIClient(user string, repo string, token string, filter FilterKind) *GithubAPI {
	return &GithubAPI{
		User:       user,
		Repository: repo,
		Filter:     filter,
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var results []*ChangelogItem
	var errors []error

	maxWorkers := 10

	itemsChan := make(chan *ChangelogItem)
	resultsChan := make(chan *ChangelogItem, len(items))
	errorsChan := make(chan error, len(items))
	var workersWG sync.WaitGroup

	workersWG.Add(maxWorkers)
	for worker := 1; worker <= maxWorkers; worker++ {
		go func() {
			defer workersWG.Done()
			for item := range itemsChan {
			rerun:
				id, err := strconv.Atoi(item.IssueID)
				if err != nil {
					cancel()
					errorsChan <- err
					return
				}

				issue, resp, err := api.Client.Issues.Get(ctx, api.User, api.Repository, id)
				if err != nil {
					cancel()
					errorsChan <- err
					return
				}

				// skip dependabot bumps
				if issue.GetUser().GetLogin() == "dependabot-preview[bot]" {
					continue
				}

				// If the PR is a cherry-pick from prow, get the parent PR's info,
				// so that it doesn't show kubermatic-bot as the author.
				if isCherry, parentID := isCherryPick(issue.GetBody()); isCherry {
					fmt.Fprintln(os.Stderr, "Issue", id, "appears to be a cherry-pick of", parentID)
					item.IssueID = parentID
					goto rerun
				}

				if resp.StatusCode != 404 && issue != nil && issue.IsPullRequest() {
					switch api.Filter {
					case FilterReleaseNotes:
						if hasReleaseNotes(issue.GetBody()) {
							text, changeType := filter(issue.GetBody())

							item.Author = issue.GetUser().GetLogin()
							item.AuthorURL = fmt.Sprintf("https://github.com/%v", issue.GetUser().GetLogin())
							item.Text = text
							item.ChangeType = changeType
							item.IssueURL = fmt.Sprintf("https://github.com/%v/%v/issues/%v", api.User, api.Repository, id)

							resultsChan <- item
						} else {
							if !hasNotesNone(issue.GetBody()) {
								errorsChan <- fmt.Errorf("PR https://github.com/%v/%v/issues/%v is missing the release note", api.User, api.Repository, id)
							}
						}
					case FilterReleaseNotesNone:
						if hasNotesNone(issue.GetBody()) {
							item.Author = issue.GetUser().GetLogin()
							item.AuthorURL = fmt.Sprintf("https://github.com/%v", issue.GetUser().GetLogin())
							item.Text = issue.GetTitle()
							item.IssueURL = fmt.Sprintf("https://github.com/%v/%v/issues/%v", api.User, api.Repository, id)

							resultsChan <- item
						} else {
							if !hasReleaseNotes(issue.GetBody()) {
								errorsChan <- fmt.Errorf("PR https://github.com/%v/%v/issues/%v is missing the release note", api.User, api.Repository, id)
							}
						}
					case FilterNone:
						item.Author = issue.GetUser().GetLogin()
						item.AuthorURL = fmt.Sprintf("https://github.com/%v", issue.GetUser().GetLogin())
						item.Text = issue.GetTitle()
						item.IssueURL = fmt.Sprintf("https://github.com/%v/%v/issues/%v", api.User, api.Repository, id)

						resultsChan <- item
					}
				}
			}
		}()
	}

	var wg sync.WaitGroup

	// close the result/error channels after the workers are done
	wg.Add(1)
	go func() {
		defer wg.Done()
		workersWG.Wait()
		close(resultsChan)
		close(errorsChan)
	}()

	// stream the jobs to the workers
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, item := range items {
			select {
			case itemsChan <- item:
			case <-ctx.Done():
			}
		}
		close(itemsChan)
	}()

	// collect results
	wg.Add(1)
	go func() {
		defer wg.Done()
		for result := range resultsChan {
			results = append(results, result)
		}
	}()

	// collect errors
	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range errorsChan {
			if err != ctx.Err() {
				errors = append(errors, err)
			}
		}
	}()

	wg.Wait()

	if len(errors) > 0 {
		var s string
		for _, err := range errors {
			s += fmt.Sprintf("%v\n", err)
		}
		return nil, fmt.Errorf(s)
	}

	sort.Slice(results, func(i int, j int) bool {
		return sort.StringsAreSorted([]string{results[i].IssueID, results[j].IssueID})
	})

	return results, nil
}

func filter(message string) (string, string) {
	body := strings.Replace(message, "```", "___", -1)
	regex := `___release-note(.*)(.*\n[\s\S]*?\n)___`

	// get matching groups
	submatches := regexp.MustCompile(regex).FindStringSubmatch(body)

	noteType := strings.ToLower(strings.TrimSpace(submatches[1]))
	text := submatches[2]

	if noteType == "" {
		noteType = "misc"
	}

	// replace linebreaks
	parser := regexp.MustCompile(`\r?\n`)
	text = parser.ReplaceAllString(text, "")
	return text, noteType
}

func hasReleaseNotes(message string) bool {
	body := strings.Replace(message, "```", "___", -1)

	// Special case: check if pr message release notes field with content NONE, return false
	regex := `___release-note(.*\n((?i)none.*)*?\n)___`
	if matched, _ := regexp.MatchString(regex, body); matched == true {
		return false
	}

	regex = `___release-note(.*\n[\s\S]*?\n)___`
	matched, _ := regexp.MatchString(regex, body)
	return matched
}

func isCherryPick(message string) (bool, string) {
	re := regexp.MustCompile(`This is an automated cherry-pick of #([0-9]+)`)
	matches := re.FindStringSubmatch(message)
	if matches != nil && matches[1] != "" {
		return true, matches[1]
	}

	return false, ""
}

func hasNotesNone(message string) bool {
	body := strings.Replace(message, "```", "___", -1)

	// Special case: check if pr message release notes field with content NONE, return false
	regex := `___release-note(.*\n((?i)none.*)*?\n)___`
	matched, _ := regexp.MatchString(regex, body)
	return matched
}
