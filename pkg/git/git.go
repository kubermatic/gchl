package git

import (
	"log"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type (

	// Git holds information about the local repository
	Git struct {
		repo *git.Repository
	}

	// Changelog represents the changelog
	Changelog struct {
		Version       string
		RepositoryURL string
		Items         []*ChangelogItem
	}

	// ChangelogItem represents an item in the changelog
	ChangelogItem struct {
		Author     string
		AuthorURL  string
		Hash       string
		Text       string
		IssueID    string
		IssueURL   string
		ChangeType string
	}

	searchFunc func(hash string) (*plumbing.Reference, error)
)

var errNotFound = errors.New("Unable to find reference")

// New returns a new local git client
func New(path string) *Git {
	return &Git{
		repo: open(path),
	}
}

func open(path string) *git.Repository {
	repository, err := git.PlainOpen(path)
	if err != nil {
		log.Fatalf("Unable to open git repository %s: %v", path, err)
	}

	return repository
}

func (g *Git) getHashObject(hash string) (*plumbing.Reference, error) {
	if len(hash) > 40 {
		return nil, errors.Errorf("hash prefix %s is too long", hash)
	}
	
	_, err := g.repo.CommitObject(plumbing.NewHash(hash))
	if err != nil {
		return nil, errNotFound
	}

	result := plumbing.NewReferenceFromStrings(hash, hash)
	return result, nil
}

func (g *Git) getHashObjectByPrefix(hash string) (*plumbing.Reference, error) {
	if len(hash) > 40 {
		return nil, errors.Errorf("hash prefix %s is too long", hash)
	}

	commits, err := g.repo.CommitObjects()
	if err != nil {
		return nil, err
	}

	var matches int
	var result *plumbing.Reference
	commits.ForEach(func(reference *object.Commit) error {
		commitHash := reference.Hash.String()[0:len(hash)]
		if commitHash == hash {
			result = plumbing.NewReferenceFromStrings(reference.Hash.String(), reference.Hash.String())
			matches++
		}
		return nil
	})

	if matches > 1 {
		return nil, errors.Errorf("Multiple references found for %s.\nPlease extend the hash prefix or provide the full commit hash as found in `git log`.", hash)
	}

	if result == nil {
		return nil, errNotFound
	}
	return result, nil
}

func (g *Git) getHashObjectByTagName(tagName string) (*plumbing.Reference, error) {
	tags, err := g.repo.Tags()
	if err != nil {
		return nil, err
	}

	var result *plumbing.Reference
	tags.ForEach(func(reference *plumbing.Reference) error {
		if tagName == reference.Name().Short() {
			result = reference
			return errors.New("ErrStop")
		}
		return nil
	})

	if result == nil {
		return nil, errNotFound
	}
	return result, nil
}

// GetReference returns a reference for a given name (e.g. tag name, branch name, hash value or hash prefix (first 7 digits))
func (g *Git) GetReference(name string) (*plumbing.Reference, error) {
	searchFunctions := []searchFunc{
		g.getHashObject,
		g.getHashObjectByBranchName,
		g.getHashObjectByTagName,
		g.getHashObjectByPrefix,
	}

	for _, search := range searchFunctions {
		result, err := search(name)
		if result != nil {
			return result, nil
		}
		if err != errNotFound {
			return nil, err
		}
	}

	return nil, errors.Errorf("Unable to find reference %v in commit history.\nPlease try another hash, branch or tag as a parameter.", name)
}

func (g *Git) getHashObjectByBranchName(branchName string) (*plumbing.Reference, error) {
	branches, err := g.repo.Branches()
	if err != nil {
		return nil, err
	}

	var result *plumbing.Reference
	branches.ForEach(func(reference *plumbing.Reference) error {
		if branchName == reference.Name().Short() {
			result = reference
			return errors.New("ErrStop")
		}
		return nil
	})

	if result == nil {
		return nil, errNotFound
	}
	return result, nil
}

// GetCommitsBetween return a list of ChangelogItems between two given references
func (g *Git) GetCommitsBetween(from *plumbing.Reference, to *plumbing.Reference) ([]*ChangelogItem, error) {
	var history []*ChangelogItem
	var exists bool
	knownIssues := make(map[string]bool)

	olderVersionCommits, err := g.repo.Log(&git.LogOptions{From: to.Hash()})
	if err != nil {
		return nil, err
	}
	// Get a set of all shared commits
	err = olderVersionCommits.ForEach(func(commit *object.Commit) error {
		if hasIssue(commit.Message) {
			issue := getIssueFrom(commit.Message)
			knownIssues[issue] = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	newVersionCommits, err := g.repo.Log(&git.LogOptions{From: from.Hash()})
	if err != nil {
		return history, err
	}
	err = newVersionCommits.ForEach(func(commit *object.Commit) error {
		// check whether the old version is even within the new version's history
		if commit.Hash == to.Hash() {
			exists = true
		}

		// ignore merge commits
		if len(commit.ParentHashes) > 1 {
			return nil
		}

		if hasIssue(commit.Message) {
			// check whether the issue is already present within the older version's history
			if issue := getIssueFrom(commit.Message); !knownIssues[issue] {
				history = append(history, &ChangelogItem{
					Hash:    commit.Hash.String(),
					Text:    commit.Message,
					IssueID: issue,
					Author:  commit.Author.Name,
				})
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if exists {
		return history, nil
	}

	return history, errors.Errorf("Unable to compare references, %v not found in history of %v", to.Name().Short(), from.Name().Short())
}

// GetCommitsSince return a list of ChangelogItems since a given reference
func (g *Git) GetCommitsSince(to *plumbing.Reference) ([]*ChangelogItem, error) {
	var history []*ChangelogItem
	var exists bool

	ref, err := g.repo.Head()
	if err != nil {
		return history, err
	}

	commits, err := g.repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return history, err
	}

	// Iterate over all commits
	// Break when `since` has been found
	commits.ForEach(func(commit *object.Commit) error {
		if commit.Hash == to.Hash() {
			exists = true
			return errors.New("ErrStop")
		}

		// Check if commit message contains issue in form `(#0..9)`
		// and add commit as a changelog item
		if hasIssue(commit.Message) {
			history = append(history, &ChangelogItem{
				Hash:      commit.Hash.String(),
				Text:      commit.Message,
				IssueID:   getIssueFrom(commit.Message),
				Author:    commit.Author.Name,
				AuthorURL: commit.Author.Email,
			})
		}
		return nil
	})

	if exists {
		return history, nil
	}
	return history, errors.Errorf("Unable to compare references, %v not found in history of %v", to.Name().Short(), ref.Name().Short())
}

func hasIssue(message string) bool {
	matches, _ := regexp.MatchString(`\(#(\d*?)\)`, message)
	return matches
}

func getIssueFrom(message string) string {
	regex := regexp.MustCompile(`\(#(\d*?)\)`)
	match := regex.FindStringSubmatch(message)

	// return last found match in commit message
	if len(match) != 0 {
		return match[len(match)-1]
	}
	return ""
}

// GetRemoteCredentials returns user/org name, repository and token parsed from the `--remote` flag
// When `--remote` is not set, the current repositories origin url is parsed
func (g *Git) GetRemoteCredentials(c *cli.Context) (string, string, string, error) {
	if c.GlobalString("token") == "" {
		return "", "", "", errors.New("Your `Github Personal Access Token` is missing.\nPlease provide a personal access token via flag `--token` or environment variable `GCHL_GITHUB_TOKEN`.\n\nYou may create one here: https://github.com/settings/tokens")
	}

	if remote := c.GlobalString("remote"); remote != "" {
		user, repo, err := parseRemoteString(remote)
		return user, repo, c.GlobalString("token"), err
	}

	remotes, err := g.repo.Remotes()
	if err != nil {
		return "", "", "", err
	}

	if len(remotes) == 0 {
		return "", "", "", errors.New("No `--remote` flag provided and current repository does not have a remote origin")
	}

	remote := remotes[0].Config().URLs[0]
	user, repo, err := parseRemoteString(remote)
	return user, repo, c.GlobalString("token"), err
}

func parseRemoteString(remoteURL string) (string, string, error) {
	// ssh://git@github.com/kubermatic/kubermatic
	if strings.HasPrefix(remoteURL, "ssh://git@github.com") {
		if strings.HasSuffix(remoteURL, ".git") {
			remoteURL = strings.TrimSuffix(remoteURL, ".git")
		}
		remoteURL = strings.TrimPrefix(remoteURL, "ssh://git@github.com/")
		credentials := strings.Split(remoteURL, "/")
		return credentials[0], credentials[1], nil
	}

	// git@github.com:kubermatic/kubermatic.git
	if strings.HasPrefix(remoteURL, "git@github.com") {
		if strings.HasSuffix(remoteURL, ".git") {
			remoteURL = strings.TrimSuffix(remoteURL, ".git")
		}
		remoteURL = strings.TrimPrefix(remoteURL, "git@github.com:")
		credentials := strings.Split(remoteURL, "/")
		return credentials[0], credentials[1], nil
	}

	// https://github.com/kubermatic/kubermatic.git
	if strings.HasPrefix(remoteURL, "https://github.com") {
		if strings.HasSuffix(remoteURL, ".git") {
			remoteURL = strings.TrimSuffix(remoteURL, ".git")
		}
		remoteURL = strings.TrimPrefix(remoteURL, "https://github.com/")
		credentials := strings.Split(remoteURL, "/")
		return credentials[0], credentials[1], nil
	}

	return "", "", errors.New("Unable to parse remote url string.\n Must be in format `https://github.com/x/x` or `git@github.com:x/x.git`")

}
