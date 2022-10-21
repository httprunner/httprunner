#!/bin/bash
# build hrp cli binary for testing
# release will be triggered on github actions, see .github/workflows/release.yml

# Usage:
# $ make build
# $ make build tags=opencv
# or
# $ bash scripts/build.sh
# $ bash scripts/build.sh opencv

set -e
set -x

# prepare path
mkdir -p "output"
bin_path="output/hrp"

# optional build tags: opencv
tags=$1

# build
if [ -z "$tags" ]; then
    go build -ldflags '-s -w' -o "$bin_path" hrp/cmd/cli/main.go
else
    go build -ldflags '-s -w' -tags "$tags" -o "$bin_path" hrp/cmd/cli/main.go
fi

# check output and version
ls -lh "$bin_path"
chmod +x "$bin_path"
./"$bin_path" -v
