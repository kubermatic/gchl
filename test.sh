#!/bin/bash

# Copyright 2020 The Kubermatic Kubernetes Platform contributors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Perform local linting for codestyle and correctness issues.

set -euo pipefail

# Checking for formatting errors
function check_gofmt {
    GFMT=$(find . -not \( \( -wholename "./vendor" \) -prune \) -name "*.go" | xargs gofmt -l)
    if [ -n "$GFMT" ]; then echo "gofmt command needs to be executed on " $GFMT && exit 1; fi
}

# Checking for correctness errors
function check_golint {
    PACKAGES=$(go list ./...)
    go vet $PACKAGES
    golint -set_exit_status $PACKAGES
}

# Checking for dependency issues
function check_godep {
    dep status -v
}

function run_tests {
    go test ./... -v --cover
}

check_gofmt
check_godep
check_golint
run_tests
echo "RUN ALL TEST SUCCESSFULLY - GOOD JOB! <3"
