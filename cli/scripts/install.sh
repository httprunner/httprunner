#!/bin/bash
# install hrp with one shell command
# bash -c "$(curl -ksSL https://raw.githubusercontent.com/httprunner/hrp/main/cli/scripts/install.sh)"

set -e

function echoError() {
    echo -e "\033[31m✘ $1\033[0m" # red
}
export -f echoError

function echoInfo() {
    echo -e "\033[32m✔ $1\033[0m" # green
}
export -f echoInfo

function echoWarn() {
    echo -e "\033[33m! $1\033[0m" # yellow
}
export -f echoError

function get_latest_version() {
    #   <title>Release v0.4.0 · httprunner/hrp · GitHub</title>
    curl -sL https://github.com/httprunner/hrp/releases/latest | grep '<title>Release' | cut -d" " -f4
}

function get_arch() {
    arch=$(uname -m)
    if [ "$arch" == "x86_64" ]; then
        arch="amd64"
    fi
    echo "$arch"
}

function main() {
    echoInfo "Detect target hrp package..."
    version=$(get_latest_version)
    echo "Latest version: $version"
    os=$(uname -s)
    echo "Current OS: $os"
    arch=$(get_arch)
    echo "Current ARCH: $arch"
    pkg="hrp-$version-$os-$arch.tar.gz"
    url="https://github.com/httprunner/hrp/releases/download/$version/$pkg"
    echo "Selected package: $url"
    echo

    echoInfo "Created temp dir..."
    echo "$ mktemp -d -t hrp"
    tmp_dir=$(mktemp -d -t hrp)
    echo "$tmp_dir"
    cd "$tmp_dir"
    echo

    echoInfo "Downloading..."
    echo "$ curl -kL $url -o $pkg"
    curl -kL $url -o "$pkg"
    echo

    echoInfo "Extracting..."
    echo "$ tar -xzf $pkg"
    tar -xzf "$pkg"
    ls -lh
    echo

    echoInfo "Installing..."
    if hrp -v > /dev/null; then
        echoWarn "$(hrp -v) exists, remove first !!!"
        echo "$ rm -rf $(which hrp)"
        rm -rf "$(which hrp)"
    fi

    echo "chmod +x hrp && mv hrp /usr/local/bin/"
    chmod +x hrp
    mv hrp /usr/local/bin/
    echo

    echoInfo "Check installation..."
    echo "$ which hrp"
    which hrp
    echo "$ hrp -v"
    hrp -v
    echo "$ hrp -h"
    hrp -h
}

main
