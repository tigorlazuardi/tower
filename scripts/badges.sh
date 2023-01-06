#!/bin/bash

function build_badge {
	local PACKAGE=$1
	local COVERAGE=$2

	if [[ $PACKAGE == '.' ]]; then
		PACKAGE='tower'
	fi

	local COLOR=important

	if (($(echo "$COVERAGE <= 50" | bc -l))); then
		COLOR=critical
	elif (($(echo "$COVERAGE >= 80" | bc -l))); then
		COLOR=success
	fi

	local BADGE_URI=$(echo "$PACKAGE-$COVERAGE%25 coverage-$COLOR" | sed -r 's/\s+/%20/g')
	local URL="https://img.shields.io/badge/$BADGE_URI.svg"
	echo "Creating Badge for $PACKAGE: $URL"
	curl -sSL $URL --create-dirs -o "./dist/$PACKAGE.svg"
}

go test -coverprofile=coverage.out . >/dev/null
COV=$(go tool cover -func=coverage.out | grep total: | grep -Eo '[0-9]+\.[0-9]+')

build_badge '.' $COV

PACKAGES=$(find . -name go.mod | grep -v '^\./go.mod$' | cut -d/ -f2- | xargs dirname)
for pkg in $PACKAGES; do
	echo "Checking ./$pkg"
	go test -coverprofile=coverage.out ./$pkg >/dev/null
	COV=$(go tool cover -func=coverage.out | grep total: | grep -Eo '[0-9]+\.[0-9]+')
  if (($(echo "$COV < 1" | bc -l))); then
    echo "go tests reports 0 output. Skipping $pkg"
    continue
  fi
	build_badge $pkg $COV
done
