#!/bin/sh

run_goimports() {
  find . -type f -name '*.go' ! -path './vendor/*' ! -path './gen/*' -exec goimports -w -local github.com/hedisam/goactor {} +
}

if [ "$1" = "goimports" ]; then
  echo "Running goimports excluding vendor/* and gen/*"
  run_goimports
else
  echo "Invalid argument. Valid arguments are [goimports]."
  exit 1
fi