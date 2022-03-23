#!/bin/bash
# build hrp cli binary for testing
# release will be triggered on github actions, see .github/workflows/release.yml

# Usage:
# $ make bump version=v0.5.2
# or
# $ bash cli/scripts/bump_version.sh v0.5.2

set -e

version=$1

if [ -z "$version" ]; then
    echo "version is required"
    exit 1
fi

echo "bump hrp version to $version"
sed -i'.bak' "s/\".*\"/\"v$version\"/g" hrp/internal/version/init.go

echo "bump install.sh version to $version"
sed -i'.bak' "s/LATEST_VERSION=\".*\"/LATEST_VERSION=\"v$version\"/g" scripts/install.sh

echo "bump httprunner version to $version"
sed -i'.bak' "s/__version__ = \".*\"/__version__ = \"$version\"/g" httprunner/__init__.py

echo "bump pyproject.toml version to $version"
sed -i'.bak' "s/^version = \".*\"/version = \"$version\"/g" pyproject.toml
