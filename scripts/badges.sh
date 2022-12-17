#!/bin/bash

go test -coverprofile=coverage.out .
COVERAGE=$(go tool cover -func=coverage.out | grep total: | grep -Eo '[0-9]+\.[0-9]+')
COLOR=orange
if (($(echo "$COVERAGE <= 50" | bc -l))); then
	COLOR=red
elif (($(echo "$COVERAGE > 80" | bc -l))); then
	COLOR=green
fi
BADGE_URI=$(echo "tower-$COVERAGE%25 coverage-$COLOR" | sed -r 's/\s+/%20/g')
URL="https://img.shields.io/badge/$BADGE_URI.svg"
echo "Creating Badge for tower: $URL"
curl -sSL $URL --create-dirs -o "./dist/tower.svg"

PACKAGES=$(find . -name go.mod | grep -v '^\./go.mod$' | cut -d/ -f2- | xargs dirname)
for pkg in $PACKAGES; do
	echo "Checking ./$pkg"
	go test -coverprofile=coverage.out ./$pkg
	# go tool cover -func=coverage.out
	COVERAGE=$(go tool cover -func=coverage.out | grep total: | grep -Eo '[0-9]+\.[0-9]+')
	COLOR=important
	if (($(echo "$COVERAGE < 1" | bc -l))); then
		continue
	fi
	if (($(echo "$COVERAGE <= 50" | bc -l))); then
		COLOR=critical
	elif (($(echo "$COVERAGE > 80" | bc -l))); then
		COLOR=success
	fi
	BADGE_URI=$(echo "$pkg-$COVERAGE%25 coverage-$COLOR" | sed -r 's/\s+/%20/g')
	URL="https://img.shields.io/badge/$BADGE_URI.svg"
	echo "Creating Badge for $pkg: $URL"
	curl -sSL $URL --create-dirs -o "./dist/$pkg.svg"
done
