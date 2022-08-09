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
	"regexp"
	"strings"

	"k8c.io/gchl/pkg/types"

	"github.com/go-openapi/inflect"
)

func generateChange(commit types.Commit) (*Change, error) {
	releaseNote, rnChangeType := extractReleaseNoteBlock(commit.PullRequest.Body)
	if releaseNote == "" || strings.ToLower(releaseNote) == "none" {
		return nil, nil
	}

	changeType := determineChangeType(commit, rnChangeType)
	changeType = humanReadableChangeType(changeType)

	return &Change{
		Commit: commit,

		Type:        changeType,
		ReleaseNote: releaseNote,
	}, nil
}

var (
	releaseNoteRegex = regexp.MustCompile(`___release-note(.*)(.*\n[\s\S]*?\n)___`)
	newlineRegex     = regexp.MustCompile(`\r?\n`)
)

func extractReleaseNoteBlock(message string) (string, string) {
	body := strings.Replace(message, "```", "___", -1)

	match := releaseNoteRegex.FindStringSubmatch(body)
	if match == nil {
		return "", ""
	}

	changeType := strings.ToLower(strings.TrimSpace(match[1]))
	releaseNote := match[2]

	// replace linebreaks
	releaseNote = strings.TrimSpace(newlineRegex.ReplaceAllString(releaseNote, ""))
	releaseNote = strings.TrimSuffix(releaseNote, ".")
	releaseNote = inflect.Capitalize(releaseNote)
	releaseNote = harmonizePrefixes(releaseNote)

	if changeType == "" {
		if isBreakingChange(releaseNote) {
			changeType = "breaking"
		} else if isBugfix(releaseNote) {
			changeType = "bugfix"
		} else if isUpdate(releaseNote) {
			changeType = "update"
		} else {
			changeType = "none"
		}
	}

	changeType = harmonizeChangeType(changeType)

	return releaseNote, changeType
}

func isBreakingChange(releaseNote string) bool {
	releaseNote = strings.ToLower(releaseNote)
	return strings.Contains(releaseNote, "action required") || strings.Contains(releaseNote, "breaking change")
}

func isBugfix(releaseNote string) bool {
	return strings.HasPrefix(releaseNote, "Fix ")
}

var isUpdateRegex = regexp.MustCompile(`update (.+)( version)? to (.+?)$`)

func isUpdate(releaseNote string) bool {
	return isUpdateRegex.MatchString(strings.ToLower(releaseNote))
}

func harmonizeChangeType(changeType string) string {
	if changeType == "breaking change" {
		changeType = "breaking"
	}

	return changeType
}

func harmonizePrefixes(text string) string {
	replacements := map[string]string{
		"Fixes":       "Fix",
		"Fixed":       "Fix",
		"Fixing":      "Fix",
		"Adds":        "Add",
		"Added":       "Add",
		"Adding":      "Add",
		"Updates":     "Update",
		"Updated":     "Update",
		"Updating":    "Update",
		"Upgrade":     "Update",
		"Upgrades":    "Update",
		"Upgraded":    "Update",
		"Upgrading":   "Update",
		"Bump":        "Update",
		"Bumps":       "Update",
		"Bumped":      "Update",
		"Bumping":     "Update",
		"Changes":     "Change",
		"Changed":     "Change",
		"Changing":    "Change",
		"Replaces":    "Replace",
		"Replaced":    "Replace",
		"Replacing":   "Replace",
		"Removes":     "Remove",
		"Removed":     "Remove",
		"Removing":    "Remove",
		"Deprecates":  "Deprecate",
		"Deprecated":  "Deprecate",
		"Deprecating": "Deprecate",
	}

	for old, new := range replacements {
		old += " "
		new += " "

		if strings.HasPrefix(text, old) {
			// only ever replace the first occurence
			text = strings.Replace(text, old, new, 1)
		}
	}

	return text
}

func determineChangeType(commit types.Commit, releaseNoteChangeType string) string {
	for _, label := range commit.PullRequest.Labels.List() {
		if strings.HasPrefix(label, "kind/") {
			return strings.TrimPrefix(label, "kind/")
		}
	}

	return releaseNoteChangeType
}

var knownChangeTypes = map[string]string{
	"api-change":      "api changes",
	"update":          "updates",
	"fix":             "bugfixes",
	"fixes":           "bugfixes",
	"bugfix":          "bugfixes",
	"bug":             "bugfixes",
	"feature":         "new feature",
	"features":        "new feature",
	"misc":            "miscellaneous",
	"none":            "miscellaneous",
	"breaking-change": "breaking changes",
	"breaking":        "breaking changes",
}

func humanReadableChangeType(identifier string) string {
	translated, ok := knownChangeTypes[identifier]
	if ok {
		identifier = translated
	}

	identifier = strings.ReplaceAll(identifier, "-", " ")
	identifier = inflect.Titleize(identifier)
	identifier = strings.ReplaceAll(identifier, "Api", "API")

	return identifier
}
