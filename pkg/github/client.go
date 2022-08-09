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
	"errors"
	"fmt"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Client struct {
	client *githubv4.Client
	log    logrus.FieldLogger
}

func NewClient(ctx context.Context, log logrus.FieldLogger, token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("token cannot be empty")
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token,
		},
	)
	httpClient := oauth2.NewClient(ctx, src)
	client := githubv4.NewClient(httpClient)

	return &Client{
		client: client,
		log:    log,
	}, nil
}

func getNumberedQueryVariables(numbers []int, max int) map[string]interface{} {
	if len(numbers) > max {
		panic(fmt.Sprintf("List contains more (%d) than possible (%d) PR numbers.", len(numbers), max))
	}

	variables := map[string]interface{}{}

	for i := 0; i < max; i++ {
		number := 0
		has := false

		if i < len(numbers) {
			number = numbers[i]
			has = true
		}

		variables[fmt.Sprintf("number%d", i)] = githubv4.Int(number)
		variables[fmt.Sprintf("has%d", i)] = githubv4.Boolean(has)
	}

	return variables
}
