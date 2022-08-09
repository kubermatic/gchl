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

	"k8c.io/gchl/pkg/types"

	"github.com/shurcooL/githubv4"
	"k8s.io/apimachinery/pkg/util/sets"
)

type graphqlPullRequest struct {
	Number int
	Title  string
	Body   string
	Author struct {
		Login string
	}

	Labels struct {
		Nodes []struct {
			Name string
		}
	} `graphql:"labels(first: 50)"`
}

func (c *Client) FetchBatchPullRequests(ctx context.Context, owner string, name string, numbers []int) (map[int]types.PullRequest, error) {
	result := map[int]types.PullRequest{}

	for {
		size := MaxPullRequestsPerQuery
		if len(numbers) < size {
			size = len(numbers)
		}
		chunk := numbers[:size]

		// do stuff
		chunkResult, err := c.fetchPullRequests(ctx, owner, name, chunk)
		if err != nil {
			return nil, err
		}

		for k, v := range chunkResult {
			result[k] = v
		}

		// shrink list
		numbers = numbers[size:]
		if len(numbers) == 0 {
			break
		}
	}

	return result, nil
}

func (c *Client) fetchPullRequests(ctx context.Context, owner string, name string, numbers []int) (map[int]types.PullRequest, error) {
	variables := getNumberedQueryVariables(numbers, MaxPullRequestsPerQuery)
	variables["owner"] = githubv4.String(owner)
	variables["name"] = githubv4.String(name)

	c.log.WithField("prs", len(numbers)).Debug("fetchPullRequests()")

	var q numberedPullRequestQuery

	err := c.client.Query(ctx, &q, variables)
	if err != nil {
		return nil, err
	}

	prs := map[int]types.PullRequest{}
	for _, pr := range q.GetAll() {
		prs[pr.Number] = convertPullRequest(pr)
	}

	return prs, nil
}

func convertPullRequest(api graphqlPullRequest) types.PullRequest {
	labels := sets.NewString()
	for _, label := range api.Labels.Nodes {
		labels.Insert(label.Name)
	}

	return types.PullRequest{
		Number: api.Number,
		Title:  api.Title,
		Body:   api.Body,
		Labels: labels,
	}
}
