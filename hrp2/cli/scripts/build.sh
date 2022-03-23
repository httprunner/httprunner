#!/bin/bash
# build hrp cli binary for testing
# release will be triggered on github actions, see .github/workflows/release.yml

# Usage:
# $ make build
# or
# $ bash cli/scripts/build.sh

set -e
set -x

# prepare path
mkdir -p "output"
bin_path="output/hrp"

# build
go build -ldflags '-s -w' -o "$bin_path" cli/hrp/main.go

# check output and version
ls -lh "$bin_path"
chmod +x "$bin_path"
./"$bin_path" -v
