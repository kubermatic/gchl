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

package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"k8c.io/gchl/pkg/changelog"
	"k8c.io/gchl/pkg/github"
	"k8c.io/gchl/pkg/ranges"
	"k8c.io/gchl/pkg/render"
	"k8c.io/gchl/pkg/signals"
	"k8c.io/gchl/pkg/types"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
)

func main() {
	ctx := signals.SetupSignalHandler()

	opts := &types.Options{}
	opts.AddFlags(pflag.CommandLine)
	pflag.Parse()

	if err := opts.Parse(); err != nil {
		log.Fatalf("Invalid options: %v", err)
	}

	logger := logrus.New()
	if opts.Verbose {
		logger.SetLevel(logrus.DebugLevel)
	}
	flogger := logger.WithFields(logrus.Fields{
		"org":     opts.Organization,
		"repo":    opts.Repository,
		"version": opts.ForVersion,
	})

	client, err := github.NewClient(ctx, flogger, opts.GithubToken)
	if err != nil {
		log.Fatalf("Failed to create GitHub client: %v", err)
	}

	flogger.Info("Resolving release commit range…")
	head, stop, err := ranges.DetermineRange(ctx, client, flogger, opts)
	if err != nil {
		log.Fatalf("Failed to determine commit range: %v", err)
	}

	flogger.Info("Fetching commit history…")
	commits, err := client.History(ctx, opts.Organization, opts.Repository, head, stop)
	if err != nil {
		log.Fatalf("Failed to fetch repository history: %v", err)
	}
	flogger.WithField("total", len(commits)).Info("Done fetching history.")

	commits = stripUnwantedCommits(commits)
	flogger.WithField("remaining", len(commits)).Info("Filtered out unwanted commits.")

	commits, err = replaceCherrypicksWithOriginals(ctx, flogger, opts, client, commits)
	if err != nil {
		log.Fatalf("Failed to filter out cherry picks: %v", err)
	}

	if len(commits) > 0 {
		top := commits[0]
		bottom := commits[len(commits)-1]

		flogger.WithFields(logrus.Fields{
			"commit": top.Hash,
			"title":  top.Title,
		}).Info("Changelog start")

		flogger.WithFields(logrus.Fields{
			"commit": bottom.Hash,
			"title":  bottom.Title,
		}).Info("Changelog end")
	} else {
		flogger.Warn("Changelog is empty.")
	}

	url := fmt.Sprintf("https://github.com/%s/%s", opts.Organization, opts.Repository)
	gen := changelog.NewGenerator(opts.ForVersion, url, commits)
	changelog, err := gen.Generate()
	if err != nil {
		log.Fatalf("Failed to create changelog from commits: %v", err)
	}

	var renderer render.Renderer
	switch opts.OutputFormat {
	case "markdown":
		renderer = render.NewMarkdownRenderer()
	case "json":
		renderer = render.NewJSONRenderer()
	default:
		log.Fatalf("Unknown output format %q.", opts.OutputFormat)
	}

	output, err := renderer.Render(changelog)
	if err != nil {
		log.Fatalf("Failed to render changelog: %v", err)
	}

	fmt.Println(output)
}

func stripUnwantedCommits(commits []types.Commit) []types.Commit {
	result := []types.Commit{}

	for i, commit := range commits {
		// skip dependabot bumps
		if commit.Author == "dependabot" {
			continue
		}

		result = append(result, commits[i])
	}

	return result
}

func replaceCherrypicksWithOriginals(ctx context.Context, log logrus.FieldLogger, opts *types.Options, client *github.Client, commits []types.Commit) ([]types.Commit, error) {
	// walk through all commits and collect PR numbers to fetch
	toFetch := sets.NewInt()
	for _, commit := range commits {
		if number := getCherrypickedFrom(commit.PullRequest.Body); number != 0 {
			toFetch.Insert(number)
		}
	}

	// nothing to do
	if toFetch.Len() == 0 {
		return commits, nil
	}

	// fetch all the cherrypicked PRs
	log.WithField("total", toFetch.Len()).Info("Fetching cherrypicks…")
	pullRequests, err := client.FetchBatchPullRequests(ctx, opts.Organization, opts.Repository, toFetch.List())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Pull Requests: %w", err)
	}

	// walk again through the commits, this time replacing the PRs
	for i, commit := range commits {
		original := getCherrypickedFrom(commit.PullRequest.Body)

		// not a cherrypick, keep it unchanged
		if original <= 0 {
			continue
		}

		pr, ok := pullRequests[original]
		if !ok {
			return nil, fmt.Errorf("could not fetch PR #%d", original)
		}

		commits[i].PullRequest = pr
	}

	return commits, nil
}

var automatedCherrypickRegex = regexp.MustCompile(`This is an automated cherry-pick of #([0-9]+)`)

func getCherrypickedFrom(prBody string) int {
	matches := automatedCherrypickRegex.FindStringSubmatch(prBody)
	if matches != nil {
		number, err := strconv.ParseInt(matches[1], 10, 32)
		if err != nil {
			panic(fmt.Sprintf("Cannot parse %q as a number.", matches[1]))
		}

		return int(number)
	}

	return 0
}
