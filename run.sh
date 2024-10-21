#!/bin/sh

run_goimports() {
  find . -type f -name '*.go' ! -path './vendor/*' ! -path './gen/*' -exec goimports -w -local github.com/hedisam/goactor {} +
}

run_gen_protos() {
  buf generate
}

if [ "$1" == "goimports" ]; then
  echo "Running goimports excluding vendor/* and gen/*"
  run_goimports
elif [ "$1" == "gen_protos" ]; then
  echo "Running buf generate..."
  run_gen_protos
else
  echo "Invalid argument. Valid arguments are [goimports, gen_protos]."
  exit 1
fi