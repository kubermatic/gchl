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

package changelog

import (
	"fmt"

	"k8c.io/gchl/pkg/types"

	"k8s.io/apimachinery/pkg/util/sets"
)

type Generator struct {
	version       string
	repositoryURL string
	commits       []types.Commit
}

func NewGenerator(version string, repositoryURL string, commits []types.Commit) *Generator {
	return &Generator{
		version:       version,
		repositoryURL: repositoryURL,
		commits:       commits,
	}
}

func (g *Generator) Generate() (*Changelog, error) {
	changes, err := g.generateChanges()
	if err != nil {
		return nil, err
	}

	groups := g.groupChanges(changes)

	return &Changelog{
		Version:       g.version,
		RepositoryURL: g.repositoryURL,
		ChangeGroups:  groups,
	}, nil
}

func (g *Generator) generateChanges() ([]Change, error) {
	result := []Change{}

	for _, commit := range g.commits {
		changes, err := processCommit(commit)
		if err != nil {
			return nil, fmt.Errorf("cannot process commit: %w", err)
		}

		result = append(result, changes...)
	}

	return result, nil
}

func (g *Generator) groupChanges(changes []Change) []ChangeGroup {
	tempMap := map[ChangeType][]Change{}

	for i, change := range changes {
		if _, exists := tempMap[change.Type]; !exists {
			tempMap[change.Type] = []Change{}
		}

		tempMap[change.Type] = append(tempMap[change.Type], changes[i])
	}

	// List() sorts for us
	changeTypes := sets.List(sets.KeySet(tempMap))

	result := []ChangeGroup{}
	for _, changeType := range changeTypes {
		changes := tempMap[changeType]

		result = append(result, ChangeGroup{
			Title:   changeType.Title(),
			Changes: changes,
		})
	}

	return result
}
