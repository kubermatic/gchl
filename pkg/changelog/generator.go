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
	"slices"
	"strings"

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

	// change groups are supposed to be sorted alphabetically, except for some
	// well known groups which we want to put in specific spots
	//
	// 1.   New Features
	// 2.   API Changes
	// 3.   Deprecations
	// ...
	// N-2. Miscellaneous
	// N-1. Chore
	// N.   Updates
	//
	// Everything between 3. and N-2 is sorted alphabetically.
	changeTypes = sortSemantically(changeTypes)

	result := []ChangeGroup{}
	for _, changeType := range changeTypes {
		changes := tempMap[changeType]

		result = append(result, ChangeGroup{
			Type:    changeType,
			Changes: changes,
		})
	}

	return result
}

func sortSemantically(changeTypes []ChangeType) []ChangeType {
	slices.SortStableFunc(changeTypes, func(a, b ChangeType) int {
		for _, check := range []ChangeType{
			ChangeTypeFeature,
			ChangeTypeAPIChange,
			ChangeTypeDeprecation,
		} {
			if a == check {
				return -1
			}
			if b == check {
				return 1
			}
		}

		for _, check := range []ChangeType{
			ChangeTypeUpdate,
			ChangeTypeChore,
			ChangeTypeMisc,
		} {
			if a == check {
				return 1
			}
			if b == check {
				return -1
			}
		}

		return strings.Compare(string(a), string(b))
	})

	return changeTypes
}
