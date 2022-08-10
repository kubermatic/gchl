# gchl

A Go-written Changelog Generator. Create Changelogs based on GitHub pull requests.

## Installation

```
go install k8c.io/gchl
```

## Usage

You will need a GitHub personal access token for API calls, create one [here](https://github.com/settings/tokens).
You can pass the token to `gchl` via the environment variable `GCHL_GITHUB_TOKEN`.

The generate is configured with a version to generate the changelog for. It will automatically determine the commit range by scanning the given repository and will then extract all release notes from all commits in the determined range. The changes are then cleaned up, grouped and printed to stdout as Markdown.

```bash
export GCHL_GITHUB_TOKEN=MYTOKENHERE
gchl --organization kubermatic --repository kubermatic --for-version v2.21.0
```

Use `--verbose` to see the API calls being made.

### Get release notes via PR message annotation

In your pull request use a Markdown code block annotated with `release-note` (Don't copy paste the example below as it uses `'` ;))

```
'''release-note
This text will be visible in changelog
'''
```

### Change Types

By default, `gchl` reads the labels from pull requests and uses the first one that starts with `kind/` as the change's type (with the `kind/` prefix stripped). If no such label exists, the release-note block can also be annotated with the type by adding it right next to `release-note`:

```
'''release-note bugfix
The important functionality has been fixed
'''
```

## Overview

```
Usage of ./gchl:
  -e, --end string            Commit hash where to stop (instead of following the branch until the previous version)
  -v, --for-version string    Name of the release to generate the changelog for
  -o, --organization string   Name of the GitHub organization
  -r, --repository string     Name of the repository
  -V, --verbose               Enable more verbose logging
```
