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
	"strings"

	"k8c.io/gchl/pkg/types"

	"github.com/go-openapi/inflect"
)

type ChangeType string

const (
	ChangeTypeAPIChange ChangeType = "api-change"
	ChangeTypeUpdate    ChangeType = "update"
	ChangeTypeBugfix    ChangeType = "bugfix"
	ChangeTypeFeature   ChangeType = "feature"
	ChangeTypeMisc      ChangeType = "misc"
)

var knownChangeTypes = map[string]ChangeType{
	"api-change":    ChangeTypeAPIChange,
	"api change":    ChangeTypeAPIChange,
	"update":        ChangeTypeUpdate,
	"updates":       ChangeTypeUpdate,
	"fix":           ChangeTypeBugfix,
	"fixes":         ChangeTypeBugfix,
	"bugfix":        ChangeTypeBugfix,
	"bugfixes":      ChangeTypeBugfix,
	"bug":           ChangeTypeBugfix,
	"feature":       ChangeTypeFeature,
	"features":      ChangeTypeFeature,
	"misc":          ChangeTypeMisc,
	"miscellaneous": ChangeTypeMisc,
	"none":          ChangeTypeMisc,
}

func ParseChangeType(s string) ChangeType {
	s = strings.ToLower(s)
	if t, ok := knownChangeTypes[s]; ok {
		return t
	}

	return ChangeType(s)
}

func (t ChangeType) Title() string {
	if t == ChangeTypeMisc {
		t = "Miscellaneous"
	}

	title := strings.ReplaceAll(string(t), "-", " ")
	title = inflect.Titleize(title)
	title = strings.ReplaceAll(title, "Api", "API")

	return title
}

type Changelog struct {
	Version       string
	RepositoryURL string
	ChangeGroups  []ChangeGroup
}

type ChangeGroup struct {
	Title   string
	Changes []Change
}

type Change struct {
	types.Commit

	Type     ChangeType `yaml:"type"`
	Breaking bool       `yaml:"breaking,omitempty"`
	Text     string     `yaml:"releaseNote"`
}
