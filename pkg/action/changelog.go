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

package action

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"k8c.io/gchl/pkg/git"
	"k8c.io/gchl/pkg/template"
)

// GenerateChangelogBetween returns a changelog on stdout
func (a *Action) GenerateChangelogBetween(c *cli.Context) error {
	local := git.New(c.GlobalString("repository"))
	user, repository, token, err := local.GetRemoteCredentials(c)
	if err != nil {
		return err
	}

	if len(c.Args()) > 2 || len(c.Args()) == 0 {
		return errors.Errorf("Usage: gchl [global options] between [reference] [reference]")
	}

	from, err := local.GetReference(c.Args().Get(0))
	if err != nil {
		return err
	}
	to, err := local.GetReference(c.Args().Get(1))
	if err != nil {
		return err
	}

	commits, err := local.GetCommitsBetween(from, to)
	if err != nil {
		return err
	}

	realeaseNotes := c.Bool("release-notes")
	realeaseNotesNone := c.Bool("release-notes-none")
	if realeaseNotes && realeaseNotesNone {
		return fmt.Errorf("--release-notes and --release-notes-none cannot be used at the same time")
	}

	filter := git.FilterNone
	if realeaseNotes {
		filter = git.FilterReleaseNotes
	}
	if realeaseNotesNone {
		filter = git.FilterReleaseNotesNone
	}

	commits, err = queryGithubAPI(user, repository, token, filter, commits)
	if err != nil {
		return err
	}

	if len(commits) == 0 {
		return errors.Errorf("No Pull Requests relevant for the changelog found between %v to %v. Exit. ", from.Name().Short(), to.Name().Short())
	}

	changelog := git.Changelog{
		Version:       c.GlobalString("for-version"),
		RepositoryURL: c.GlobalString("remote"),
		Items:         commits,
	}

	var output string
	if realeaseNotes {
		output, err = template.RenderReleaseNotes(c, &changelog)
	} else {
		output, err = template.Render(c, &changelog)
	}
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}

// GenerateChangelogSince returns a changelog on stdout
func (a *Action) GenerateChangelogSince(c *cli.Context) error {
	local := git.New(c.GlobalString("repository"))
	user, repository, token, err := local.GetRemoteCredentials(c)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	if len(c.Args()) > 1 || len(c.Args()) == 0 {
		return errors.Errorf("Usage: gchl [global options] since [reference]")
	}

	since, err := local.GetReference(c.Args().Get(0))
	if err != nil {
		return err
	}

	commits, err := local.GetCommitsSince(since)
	if err != nil {
		return err
	}

	realeaseNotes := c.Bool("release-notes")
	realeaseNotesNone := c.Bool("release-notes-none")
	if realeaseNotes && realeaseNotesNone {
		return fmt.Errorf("--release-notes and --release-notes-none cannot be used at the same time")
	}

	filter := git.FilterNone
	if realeaseNotes {
		filter = git.FilterReleaseNotes
	}
	if realeaseNotesNone {
		filter = git.FilterReleaseNotesNone
	}

	commits, err = queryGithubAPI(user, repository, token, filter, commits)
	if err != nil {
		return err
	}

	if len(commits) == 0 {
		return errors.Errorf("No Pull Requests relevant for the changelog found since %v. Exit. ", since.Name().Short())
	}

	changelog := git.Changelog{
		Version:       c.GlobalString("for-version"),
		RepositoryURL: c.GlobalString("remote"),
		Items:         commits,
	}

	var output string
	if realeaseNotes {
		output, err = template.RenderReleaseNotes(c, &changelog)
	} else {
		output, err = template.Render(c, &changelog)
	}
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}

func queryGithubAPI(user string, repository string, token string, filter git.FilterKind, commits []*git.ChangelogItem) ([]*git.ChangelogItem, error) {
	api := git.NewAPIClient(user, repository, token, filter)
	commits, err := api.CompareRemote(commits)
	if err != nil {
		return nil, err
	}
	return commits, err
}
