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
	"sort"
	"strings"

	"k8c.io/gchl/pkg/types"

	"github.com/go-openapi/inflect"
	"k8s.io/apimachinery/pkg/util/sets"
)

func processCommit(commit types.Commit) ([]Change, error) {
	commitType := commitChangeType(commit)
	releaseNotes := extractReleaseNotes(commitType, commit.PullRequest.Body)

	var changes []Change
	for _, rn := range releaseNotes {
		changes = append(changes, rn.Changes()...)
	}

	for i := range changes {
		changes[i].Commit = commit
	}

	// sort changes by text
	sort.Slice(changes, func(i, j int) bool {
		return strings.ToLower(changes[i].Text) < strings.ToLower(changes[j].Text)
	})

	return changes, nil
}

var releaseNoteRegex = regexp.MustCompile(`___release-note(.*)(.*\n[\s\S]*?\n)___`)

type releaseNote struct {
	Type     ChangeType
	Breaking bool
	Text     string
}

func extractReleaseNotes(commitType ChangeType, body string) []releaseNote {
	body = strings.Replace(body, "```", "___", -1)

	matches := releaseNoteRegex.FindAllStringSubmatch(body, -1)
	if matches == nil {
		return nil
	}

	var releaseNotes []releaseNote
	for _, match := range matches {
		changeType := ParseChangeType(strings.TrimSpace(match[1]))
		text := strings.TrimSpace(match[2])
		breaking := false

		// if the release-note block has no explicit type, use the commit's type
		if changeType == "" {
			changeType = commitType
		}

		// any change can be a breaking change, so release-note blocks declared
		// as just "breaking" technically have no type
		if strings.Contains(string(changeType), "breaking") {
			changeType = ""
			breaking = true
		}

		releaseNotes = append(releaseNotes, releaseNote{
			Type:     changeType,
			Breaking: breaking,
			Text:     text,
		})
	}

	return releaseNotes
}

func (rn *releaseNote) Changes() []Change {
	if rn.Text == "" || strings.ToLower(rn.Text) == "none" {
		return nil
	}

	items := splitIntoLines(rn.Text)

	var changes []Change
	for _, item := range items {
		changes = append(changes, itemToChange(rn.Type, rn.Breaking, item))
	}

	return changes
}

var (
	itemPrefixes    = []string{"- ", "* "}
	newlineRegex    = regexp.MustCompile(`\r?\n`)
	whitespaceRegex = regexp.MustCompile(`\s+`)
)

func splitIntoLines(text string) []string {
	lines := strings.Split(text, "\n")

	var items []string
	for _, line := range lines {
		isListItem := false
		for _, prefix := range itemPrefixes {
			if strings.HasPrefix(line, prefix) {
				line = strings.TrimSpace(strings.TrimPrefix(line, prefix))
				isListItem = true
				break
			}
		}

		// if we find something that does not look like a list item, we assume the
		// release-note block does simply not contain a list, but maybe it contains
		// just multiline text
		if !isListItem {
			// collapse newlines
			text = strings.TrimSpace(newlineRegex.ReplaceAllLiteralString(text, " "))
			text = whitespaceRegex.ReplaceAllLiteralString(text, " ")
			return []string{text}
		}

		items = append(items, line)
	}

	return items
}

func itemToChange(explicitType ChangeType, breaking bool, text string) Change {
	text = strings.TrimSuffix(text, ".")
	text = inflect.Capitalize(text)
	text = harmonizeLinePrefixes(text)

	// item is breaking if it's in its text or the entire release-note block
	// is marked as breaking
	breaking = breaking || isBreakingChange(text)

	if explicitType == "" {
		switch {
		case isBugfix(text):
			explicitType = ChangeTypeBugfix
		case isUpdate(text):
			explicitType = ChangeTypeUpdate
		default:
			explicitType = ChangeTypeMisc
		}
	}

	return Change{
		Type:     explicitType,
		Breaking: breaking,
		Text:     text,
	}
}

func isBreakingChange(text string) bool {
	text = strings.ToLower(text)
	return strings.Contains(text, "action required") || strings.Contains(text, "breaking change")
}

func isBugfix(releaseNote string) bool {
	return strings.HasPrefix(releaseNote, "Fix ")
}

var isUpdateRegex = regexp.MustCompile(`updat(es|ed|ing|e) (.+)( version)? to (.+?)$`)

func isUpdate(releaseNote string) bool {
	return isUpdateRegex.MatchString(strings.ToLower(releaseNote))
}

func harmonizeLinePrefixes(text string) string {
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

func commitChangeType(commit types.Commit) ChangeType {
	for _, label := range sets.List(commit.PullRequest.Labels) {
		if strings.HasPrefix(label, "kind/") {
			return ParseChangeType(strings.TrimPrefix(label, "kind/"))
		}
	}

	return ""
}
