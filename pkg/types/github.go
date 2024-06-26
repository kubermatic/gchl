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

package types

type Commit struct {
	Hash        string      `yaml:"hash" json:"hash"`
	Title       string      `yaml:"title" json:"title"`
	PullRequest PullRequest `yaml:"pullRequest" json:"pullRequest"`
	Author      string      `yaml:"author" json:"author"`
}

type PullRequest struct {
	Number int      `yaml:"number" json:"number"`
	Title  string   `yaml:"title" json:"title"`
	Body   string   `yaml:"body" json:"body"`
	Labels []string `yaml:"labels" json:"labels"`
}

type RepositoryRefs struct {
	DefaultBranch string
	Tags          []Ref
	Branches      []Ref
}

type Ref struct {
	Name string
	Hash string
}
