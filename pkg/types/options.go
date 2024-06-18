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

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/pflag"
)

type Options struct {
	Organization string
	Repository   string
	ForVersion   string
	GithubToken  string
	End          string
	Verbose      bool
	OutputFormat string
}

var outputFormats = []string{"markdown", "json"}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Organization, "organization", "o", "", "Name of the GitHub organization")
	fs.StringVarP(&o.Repository, "repository", "r", "", "Name of the repository")
	fs.StringVarP(&o.ForVersion, "for-version", "v", "", "Name of the release to generate the changelog for")
	fs.StringVarP(&o.End, "end", "e", "", "Commit hash where to stop (instead of following the branch until the previous version)")
	fs.StringVarP(&o.OutputFormat, "format", "f", "markdown", fmt.Sprintf("Output format (one of %v)", outputFormats))
	fs.BoolVarP(&o.Verbose, "verbose", "V", false, "Enable more verbose logging")
}

func (o *Options) Parse() error {
	o.GithubToken = os.Getenv("GCHL_GITHUB_TOKEN")
	if o.GithubToken == "" {
		return errors.New("no $GCHL_GITHUB_TOKEN environment variable defined")
	}

	if o.Organization == "" {
		return errors.New("no --organization given")
	}

	if o.Repository == "" {
		return errors.New("no --repository given")
	}

	if o.ForVersion == "" {
		return errors.New("no --for-version given")
	}

	if _, err := semver.NewVersion(o.ForVersion); err != nil {
		return fmt.Errorf("--for-version %q is not a valid semver: %w", o.ForVersion, err)
	}

	// ensure no matter the user preference, we're consistent in our code and templating
	o.ForVersion = strings.TrimPrefix(o.ForVersion, "v")

	if o.OutputFormat != "" && !slices.Contains(outputFormats, o.OutputFormat) {
		return fmt.Errorf("invalid --format %q, must be one of %v", o.OutputFormat, outputFormats)
	}

	if o.OutputFormat == "" {
		o.OutputFormat = "markdown"
	}

	return nil
}
