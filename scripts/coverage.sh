#!/bin/bash

set -euo pipefail
export PATH="/opt/homebrew/bin:/usr/local/bin:$PATH"

if [ -f ".buildconfig" ]; then
    set -a
    . ./.buildconfig
    set +a
fi

if [ -n "${GOFLAGS:-}" ]; then
    export GOFLAGS="$GOFLAGS -mod=mod"
else
    export GOFLAGS="-mod=mod"
fi

COVERAGE_DIR=${COVERAGE_DIR:-coverage}
COVERAGE_THRESHOLD=${COVERAGE_THRESHOLD:-}

mkdir -p "$COVERAGE_DIR"

go test -coverpkg=./... -covermode=atomic -coverprofile="$COVERAGE_DIR/coverage.out" ./...

coverage=$(go tool cover -func="$COVERAGE_DIR/coverage.out" | awk '/^total:/ {gsub(/%/, "", $3); print $3}')

go tool cover -html="$COVERAGE_DIR/coverage.out" -o "$COVERAGE_DIR/coverage.html"

echo "Coverage report: $COVERAGE_DIR/coverage.html"
echo "Total coverage: $coverage%"

if [ -n "$COVERAGE_THRESHOLD" ]; then
    if ! awk -v coverage="$coverage" -v threshold="$COVERAGE_THRESHOLD" 'BEGIN { exit (coverage < threshold) ? 1 : 0 }'; then
        echo "Coverage too low: $coverage% (expected >= $COVERAGE_THRESHOLD%)"
        exit 1
    fi
fi
