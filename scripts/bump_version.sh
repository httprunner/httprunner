#!/bin/bash
# build hrp cli binary for testing
# release will be triggered on github actions, see .github/workflows/release.yml

# Usage:
# $ make bump version=v4.3.0
# or
# $ bash scripts/bump_version.sh v4.3.0

set -e

version=$1

if [ -z "$version" ]; then
    echo "version is required"
    exit 1
fi

if [[ $version != v* ]]; then
    version="v$version"
fi

echo "bump hrp version to $version"
echo -n "$version" > hrp/internal/version/VERSION

echo "bump httprunner version to $version"
sed -i'.bak' "s/__version__ = \".*\"/__version__ = \"$version\"/g" httprunner/__init__.py

echo "bump pyproject.toml version to $version"
sed -i'.bak' "s/^version = \".*\"/version = \"$version\"/g" pyproject.toml
