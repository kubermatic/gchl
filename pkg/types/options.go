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

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/pflag"
)

type Options struct {
	Organization string
	Repository   string
	ForVersion   string
	GithubToken  string
	Verbose      bool
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Organization, "organization", "o", "", "Name of the GitHub organization")
	fs.StringVarP(&o.Repository, "repository", "r", "", "Name of the repository")
	fs.StringVarP(&o.ForVersion, "for-version", "v", "", "Name of the release to generate the changelog for")
	fs.BoolVarP(&o.Verbose, "verbose", "V", false, "Enable more verbose logging")
}

func (o *Options) Parse() error {
	o.GithubToken = os.Getenv("GCHL_GITHUB_TOKEN")
	if o.GithubToken == "" {
		return errors.New("no $GCHL_GITHUB_TOKEN defined")
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

	return nil
}
