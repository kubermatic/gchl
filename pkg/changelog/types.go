/*
Copyright 2024 The Kubermatic Kubernetes Platform contributors.

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
	"strings"

	"k8c.io/gchl/pkg/types"
)

type ChangeType string

const (
	ChangeTypeAPIChange     ChangeType = "api-change"
	ChangeTypeBugfix        ChangeType = "bugfix"
	ChangeTypeCleanup       ChangeType = "cleanup"
	ChangeTypeDeprecation   ChangeType = "deprecation"
	ChangeTypeDocumentation ChangeType = "documentation"
	ChangeTypeFeature       ChangeType = "feature"
	ChangeTypeMisc          ChangeType = "misc"
	ChangeTypeChore         ChangeType = "chore"
	ChangeTypeRegression    ChangeType = "regresssion"
	ChangeTypeUpdate        ChangeType = "update"
)

var knownChangeTypes = map[string]ChangeType{
	"api change":    ChangeTypeAPIChange,
	"api-change":    ChangeTypeAPIChange,
	"bug":           ChangeTypeBugfix,
	"bugfix":        ChangeTypeBugfix,
	"bugfixes":      ChangeTypeBugfix,
	"chore":         ChangeTypeChore,
	"cleanup":       ChangeTypeCleanup,
	"deprecates":    ChangeTypeDeprecation,
	"deprecation":   ChangeTypeDeprecation,
	"doc":           ChangeTypeDocumentation,
	"docs":          ChangeTypeDocumentation,
	"documentation": ChangeTypeDocumentation,
	"feature":       ChangeTypeFeature,
	"features":      ChangeTypeFeature,
	"fix":           ChangeTypeBugfix,
	"fixes":         ChangeTypeBugfix,
	"misc":          ChangeTypeMisc,
	"miscellaneous": ChangeTypeMisc,
	"none":          ChangeTypeMisc,
	"regression":    ChangeTypeRegression,
	"update":        ChangeTypeUpdate,
	"updates":       ChangeTypeUpdate,
}

func ParseChangeType(s string) ChangeType {
	s = strings.ToLower(s)
	if t, ok := knownChangeTypes[s]; ok {
		return t
	}

	return ChangeType(s)
}

type Changelog struct {
	Version       string        `yaml:"version" json:"version"`
	RepositoryURL string        `yaml:"repository" json:"repository"`
	ChangeGroups  []ChangeGroup `yaml:"groups" json:"groups"`
}

type ChangeGroup struct {
	Type    ChangeType `yaml:"type" json:"type"`
	Changes []Change   `yaml:"changes" json:"changes"`
}

type Change struct {
	Commit types.Commit `yaml:"commit" json:"commit"`

	Type     ChangeType `yaml:"type"`
	Breaking bool       `yaml:"breaking,omitempty"`
	Text     string     `yaml:"releaseNote"`
}

func (c *Changelog) BreakingChanges() []Change {
	var breaks []Change

	for _, group := range c.ChangeGroups {
		for i, change := range group.Changes {
			if change.Breaking {
				breaks = append(breaks, group.Changes[i])
			}
		}
	}

	return breaks
}
