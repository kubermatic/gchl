/*
Copyright 2022 The Kubermatic Kubernetes Platform contributors.

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

package github

import (
	"context"
	"fmt"

	"k8c.io/gchl/pkg/types"

	"github.com/shurcooL/githubv4"
)

type commitSchema struct {
	OID                    string
	MessageHeadline        string
	AssociatedPullRequests struct {
		Nodes []graphqlPullRequest
	} `graphql:"associatedPullRequests(first: 5)"`
}

type historyQuery struct {
	Repository struct {
		Object struct {
			Commit struct {
				History struct {
					Nodes    []commitSchema
					PageInfo struct {
						EndCursor   githubv4.String
						HasNextPage bool
					}
				} `graphql:"history(first: 50, after: $cursor)"`
			} `graphql:"... on Commit"`
		} `graphql:"object(expression: $head)"`
	} `graphql:"repository(name: $name, owner: $owner)"`
}

type Stopper func(types.Commit) bool

// History will return all commits, beginning with the head hash, until the stop
// function returns false.
func (c *Client) History(ctx context.Context, owner string, name string, headHash string, stop Stopper) ([]types.Commit, error) {
	commits := []types.Commit{}
	cursor := ""

	for {
		var (
			err  error
			page []types.Commit
		)

		page, cursor, err = c.fetchHistoryPage(ctx, owner, name, headHash, stop, cursor)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch commits: %w", err)
		}

		commits = append(commits, page...)

		if cursor == "" {
			break
		}
	}

	return commits, nil
}

func (c *Client) fetchHistoryPage(ctx context.Context, owner string, name string, headHash string, stop Stopper, cursor string) ([]types.Commit, string, error) {
	variables := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(name),
		"head":  githubv4.String(headHash),
	}

	if cursor == "" {
		variables["cursor"] = (*githubv4.String)(nil)
	} else {
		variables["cursor"] = githubv4.String(cursor)
	}

	c.log.WithField("cursor", cursor).Debug("fetchHistory()")

	var q historyQuery

	err := c.client.Query(ctx, &q, variables)
	if err != nil {
		return nil, "", err
	}

	cursor = ""
	if info := q.Repository.Object.Commit.History.PageInfo; info.HasNextPage {
		cursor = string(info.EndCursor)
	}

	commits := []types.Commit{}
	for _, node := range q.Repository.Object.Commit.History.Nodes {
		if len(node.AssociatedPullRequests.Nodes) == 0 {
			c.log.WithField("commit", node.OID).Warn("Commit has no associated pull request.")
			continue
		}

		commit := convertCommit(node)
		if stop(commit) {
			cursor = ""
			break
		}

		commits = append(commits, commit)
	}

	return commits, cursor, nil
}

func convertCommit(api commitSchema) types.Commit {
	pr := api.AssociatedPullRequests.Nodes[0]

	commit := types.Commit{
		Hash:        api.OID,
		Title:       api.MessageHeadline,
		Author:      pr.Author.Login,
		PullRequest: convertPullRequest(pr),
	}

	return commit
}

func (c *Client) Log(ctx context.Context, owner string, name string, headHash string, maxCommits int) ([]types.Commit, error) {
	commits := []types.Commit{}
	cursor := ""

	for {
		var (
			err  error
			page []types.Commit
		)

		page, cursor, err = c.fetchLogPage(ctx, owner, name, headHash, cursor)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch commits: %w", err)
		}

		if len(page) > maxCommits {
			page = page[:maxCommits]
		}

		commits = append(commits, page...)
		maxCommits -= len(page)

		if cursor == "" || maxCommits <= 0 {
			break
		}
	}

	return commits, nil
}

func (c *Client) fetchLogPage(ctx context.Context, owner string, name string, headHash string, cursor string) ([]types.Commit, string, error) {
	variables := map[string]interface{}{
		"owner": githubv4.String(owner),
		"name":  githubv4.String(name),
		"head":  githubv4.String(headHash),
	}

	if cursor == "" {
		variables["cursor"] = (*githubv4.String)(nil)
	} else {
		variables["cursor"] = githubv4.String(cursor)
	}

	c.log.WithField("cursor", cursor).Debug("fetchLog()")

	var q historyQuery

	err := c.client.Query(ctx, &q, variables)
	if err != nil {
		return nil, "", err
	}

	cursor = ""
	if info := q.Repository.Object.Commit.History.PageInfo; info.HasNextPage {
		cursor = string(info.EndCursor)
	}

	commits := []types.Commit{}
	for _, commit := range q.Repository.Object.Commit.History.Nodes {
		if len(commit.AssociatedPullRequests.Nodes) == 0 {
			c.log.WithField("commit", commit.OID).Warn("Commit has no associated pull request.")
			continue
		}

		commits = append(commits, convertCommit(commit))
	}

	return commits, cursor, nil
}
