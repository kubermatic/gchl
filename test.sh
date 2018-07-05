#!/bin/bash
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
echo "RUN ALL TEST SUCCESSFULLY - GOOD JOB! <3"%