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
	"strings"

	"k8c.io/gchl/pkg/types"

	"github.com/shurcooL/githubv4"
)

type ref struct {
	Name   string
	Target struct {
		OID string

		// for tags, the target (ref.target.OID) is the OID of the tag itself,
		// not the commit that the tag points to (in contrast to branches, where
		// ref.target.OID is directly the commit, because branches are not objects
		// like tags). To get the actual commit hash we need to dig deeper.
		Tag struct {
			Target struct {
				OID string
			}
		} `graphql:"... on Tag"`
	}
}

type refsQuery struct {
	Repository struct {
		DefaultBranchRef struct {
			Name string
		}
		Refs struct {
			Nodes    []ref
			PageInfo struct {
				EndCursor   githubv4.String
				HasNextPage bool
			}
		} `graphql:"refs(first: 50, refPrefix: $prefix, after: $cursor)"`
	} `graphql:"repository(name: $name, owner: $owner)"`
}

func (c *Client) References(ctx context.Context, owner string, name string) (types.RepositoryRefs, error) {
	result := types.RepositoryRefs{}
	cursor := ""

	for {
		var (
			err  error
			page types.RepositoryRefs
		)

		page, cursor, err = c.fetchReferencesPage(ctx, owner, name, cursor)
		if err != nil {
			return result, fmt.Errorf("failed to fetch references: %w", err)
		}

		result.DefaultBranch = page.DefaultBranch
		result.Branches = append(result.Branches, page.Branches...)
		result.Tags = append(result.Tags, page.Tags...)

		if cursor == "" {
			break
		}
	}

	return result, nil
}

func (c *Client) fetchReferencesPage(ctx context.Context, owner string, name string, cursor string) (types.RepositoryRefs, string, error) {
	variables := map[string]interface{}{
		"owner":  githubv4.String(owner),
		"name":   githubv4.String(name),
		"prefix": githubv4.String("refs/"),
	}

	if cursor == "" {
		variables["cursor"] = (*githubv4.String)(nil)
	} else {
		variables["cursor"] = githubv4.String(cursor)
	}

	c.log.WithField("cursor", cursor).Debug("fetchReferences()")

	var q refsQuery

	err := c.client.Query(ctx, &q, variables)
	if err != nil {
		return types.RepositoryRefs{}, "", err
	}

	cursor = ""
	if info := q.Repository.Refs.PageInfo; info.HasNextPage {
		cursor = string(info.EndCursor)
	}

	result := types.RepositoryRefs{
		DefaultBranch: q.Repository.DefaultBranchRef.Name,
	}

	for _, apiRef := range q.Repository.Refs.Nodes {
		switch {
		case strings.HasPrefix(apiRef.Name, "heads/"):
			apiRef.Name = strings.TrimPrefix(apiRef.Name, "heads/")
			result.Branches = append(result.Branches, convertRef(apiRef))

		case strings.HasPrefix(apiRef.Name, "tags/"):
			apiRef.Name = strings.TrimPrefix(apiRef.Name, "tags/")
			result.Tags = append(result.Tags, convertRef(apiRef))
		}
	}

	return result, cursor, nil
}

func convertRef(api ref) types.Ref {
	hash := api.Target.Tag.Target.OID
	if hash == "" {
		hash = api.Target.OID
	}

	return types.Ref{
		Name: api.Name,
		Hash: hash,
	}
}
