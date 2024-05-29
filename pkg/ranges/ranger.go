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

package ranges

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"k8c.io/gchl/pkg/github"
	"k8c.io/gchl/pkg/types"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
)

var releaseBranchRegex = regexp.MustCompile(`^release/v([0-9]+)\.([0-9]+)$`)

func DetermineRange(ctx context.Context, client *github.Client, log logrus.FieldLogger, opts *types.Options) (string, github.Stopper, error) {
	targetVersion := opts.ForVersion

	allRepoRefs, err := client.References(ctx, opts.Organization, opts.Repository)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch references: %w", err)
	}

	// check if the target version exists as a tag in the repo
	var targetTag *types.Ref
	for i, tag := range allRepoRefs.Tags {
		if tag.Name == targetVersion {
			targetTag = &allRepoRefs.Tags[i]
			break
		}
	}

	sv, err := semver.NewVersion(opts.ForVersion)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse version %q: %w", opts.ForVersion, err)
	}

	if targetTag != nil {
		log.WithField("commit", targetTag.Hash).Info("Resolved version to be an existing tag.")
	} else {
		// No tag for the version could be found, we assume that the release isn't tagged
		// yet and we should be looking at the release branch instead. Note that for new
		// releases, there might not be a release branch yet, so we will fallback to the
		// primary branch.

		releaseBranch := fmt.Sprintf("release/v%d.%d", sv.Major(), sv.Minor())
		for i, branch := range allRepoRefs.Branches {
			if branch.Name == releaseBranch {
				targetTag = &allRepoRefs.Branches[i]
				break
			}
		}

		if targetTag != nil {
			log.WithField("branch", targetTag.Name).Infof("Resolved version to be the head a release branch.")
		} else {
			log.WithField("primary", allRepoRefs.DefaultBranch).Warn("No release branch exists for this version, falling back to primary branch.")

			for i, branch := range allRepoRefs.Branches {
				if branch.Name == allRepoRefs.DefaultBranch {
					targetTag = &allRepoRefs.Branches[i]
					break
				}
			}

			if targetTag == nil {
				return "", nil, fmt.Errorf("no commit details exist for primary branch %q", allRepoRefs.DefaultBranch)
			}
		}
	}

	// Now we know the top (latest) commit for the changelog (most likely
	// the commit that is tagged with opts.ForVersion). Now we need to figure
	// out far back we need to go to collect all relevant commits.

	// If a custom --end flag is given, this is trivial.

	if opts.End != "" {
		return targetTag.Hash, func(c types.Commit) bool {
			return strings.HasPrefix(c.Hash, opts.End)
		}, nil
	}

	// This algorithm is tailored a bit towards KKP which uses release branches
	// and sometimes tags new versions on the master branch and sometimes only
	// on the release branch (i.e. the point when a new release branch is opened
	// varies somewhere between the beta and final release tag).
	// The plan is simple: From the start (latest) we go back until
	//
	//   a) We hit another (non-alpha/beta) tag, e.g. if we're creating the changelog
	//      for v1.2.3, we go back until we find v1.2.2 (or any other tag). For
	//      patch relases this will work just fine. Note that we skip over beta and alpha
	//      so that the release notes for v1.2.3 include all changes, not just those since
	//      (for example) v1.2.3-rc.3
	//
	//      -- or --
	//
	//   b) We go back until we hit a commit that is part of the *previous* release
	//      branch. This will catch the cases where the previous relases were *all*
	//      tagged on their own release branch (i.e. when going back from 1.12, we
	//      would never encounter a 1.11 tag anywhere) or if the tag was the split
	//      point. Since for this variant we rely on commits and not on tags, it
	//      doesn't matter when and where the tags for the previous release were set,
	//      we only care about the point where the old and new release branches meet.
	//
	// To achieve (b), we need a list of commits that belong to the previous release
	// branch.

	// go back one minor release, handle underflows (do not go from v2.0 to v1.-1)
	prevMajor := int(sv.Major())
	prevMinor := int(sv.Minor()) - 1
	if prevMinor < 0 {
		prevMajor--

		// find the most recent (highest) minor release for the given major
		for _, branch := range allRepoRefs.Branches {
			match := releaseBranchRegex.FindStringSubmatch(branch.Name)
			if match != nil {
				minor, err := strconv.Atoi(match[2])
				if err == nil {
					major, err := strconv.Atoi(match[1])

					if err == nil && major == prevMajor && minor > prevMinor {
						prevMinor = minor
					}
				}
			}
		}

		if prevMinor < 0 {
			return "", nil, fmt.Errorf("could not find a release branch for any minor in the v%d release", prevMajor)
		}
	}

	prevReleaseBranch, err := findPreviousReleaseBranch(sv, allRepoRefs)
	if err != nil {
		return "", nil, err
	}

	// determine the HEAD of this previous release branch
	prevReleaseHead := ""
	for _, branch := range allRepoRefs.Branches {
		if branch.Name == prevReleaseBranch {
			prevReleaseHead = branch.Hash
			break
		}
	}

	if prevReleaseHead == "" {
		return "", nil, fmt.Errorf("could not find HEAD for release branch %q", prevReleaseBranch)
	}

	log.WithField("previous", prevReleaseBranch).Info("Detected previous release branch.")

	// 250 commits should be way more than enough (250 commits are only the number of patches since the
	// release was done, not everything that went into the release).
	log.Info("Fetching previous release commits…")
	previousReleaseCommits, err := client.Log(ctx, opts.Organization, opts.Repository, prevReleaseHead, 250)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch commits from previous release branch: %w", err)
	}

	previousReleaseCommitHashes := toLookupTable(previousReleaseCommits)
	tags := toTagLookupTable(allRepoRefs.Tags, sv)

	return targetTag.Hash, func(c types.Commit) bool {
		// We found the intersection between the previous release branch and
		// the current release branch.
		if previousReleaseCommitHashes.Has(c.Hash) {
			return true
		}

		// We found another tag
		if tags.Has(c.Hash) {
			return true
		}

		return false
	}, nil
}

func findPreviousReleaseBranch(currentVersion *semver.Version, allRepoRefs types.RepositoryRefs) (string, error) {
	// go back one minor release, handle underflows (do not go from v2.0 to v1.-1)
	prevMajor := int(currentVersion.Major())
	prevMinor := int(currentVersion.Minor()) - 1
	if prevMinor < 0 {
		prevMajor--

		// find the most recent (highest) minor release for the given major
		for _, branch := range allRepoRefs.Branches {
			match := releaseBranchRegex.FindStringSubmatch(branch.Name)
			if match != nil {
				minor, err := strconv.Atoi(match[2])
				if err == nil {
					major, err := strconv.Atoi(match[1])

					if err == nil && major == prevMajor && minor > prevMinor {
						prevMinor = minor
					}
				}
			}
		}

		if prevMinor < 0 {
			return "", fmt.Errorf("could not find a release branch for any minor in the v%d release", prevMajor)
		}
	}

	return fmt.Sprintf("release/v%d.%d", prevMajor, prevMinor), nil
}

func toLookupTable(commits []types.Commit) sets.Set[string] {
	result := sets.New[string]()
	for _, commit := range commits {
		result.Insert(commit.Hash)
	}

	return result
}

// toTagLookupTable will automatically skip all non-final releases (e.g. alphas and betas),
// so that the lookup only contains stable versions. This is because later when we use it,
// we do not want to stop at alpha versions, but only on stable versions. The same goes for
// the start version (otherwise we would stop immediately on the first commit).
func toTagLookupTable(tags []types.Ref, targetVersion *semver.Version) sets.Set[string] {
	result := sets.New[string]()
	for _, tag := range tags {
		sv, err := semver.NewVersion(tag.Name)
		if err == nil && sv.Prerelease() == "" && !targetVersion.Equal(sv) {
			result.Insert(tag.Hash)
		}
	}

	return result
}
