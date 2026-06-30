#!/bin/bash
set -euo pipefail

MIN_COVERAGE="${MIN_COVERAGE:-40}"
COVERPROFILE="${COVERPROFILE:-coverage.out}"

echo "Running tests with coverage..."
PACKAGES="$(go list ./... | grep -v '/scripts/ast-metrics$')"
go test $PACKAGES -coverprofile="$COVERPROFILE" -covermode=atomic -count=1

COVERAGE="$(go tool cover -func="$COVERPROFILE" | awk '/^total:/ { sub(/%/, "", $3); print $3 }')"
if [ -z "$COVERAGE" ]; then
    echo "Could not read total coverage from $COVERPROFILE"
    exit 1
fi

awk -v actual="$COVERAGE" -v minimum="$MIN_COVERAGE" 'BEGIN {
    if ((actual + 0) < (minimum + 0)) {
        exit 1
    }
}' || {
    echo "Coverage gate failed: ${COVERAGE}% < ${MIN_COVERAGE}%"
    exit 1
}

echo "Coverage gate passed: ${COVERAGE}% >= ${MIN_COVERAGE}%"

echo "Running AST metrics gate..."
go run ./scripts/ast-metrics \
    -max-func-lines "${MAX_FUNC_LINES:-80}" \
    -max-branches "${MAX_BRANCHES:-20}" \
    .

echo "AST metrics gate passed"
