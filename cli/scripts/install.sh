#!/bin/bash
# install hrp with one shell command
# bash -c "$(curl -ksSL https://httprunner.oss-cn-beijing.aliyuncs.com/install.sh)"

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

function get_os() {
    os=$(uname -s)
    echo "$os" | tr '[:upper:]' '[:lower:]'
}

function get_arch() {
    arch=$(uname -m)
    if [ "$arch" == "x86_64" ]; then
        arch="amd64"
    fi
    echo "$arch"
}

function get_pkg_suffix() {
    os=$1
    if [ "$os" == "windows" ]; then
        echo ".zip"
    else
        echo ".tar.gz"
    fi
}

function get_download_url() {
    # github
    # url="https://github.com/httprunner/hrp/releases/download/$version/$1"
    # aliyun oss
    url="https://httprunner.oss-cn-beijing.aliyuncs.com/$1"
    echo $url
}

function main() {
    echoInfo "Detect target hrp package..."
    version=$(get_latest_version)
    echo "Latest version: $version"
    os=$(get_os)
    echo "Current OS: $os"
    arch=$(get_arch)
    echo "Current ARCH: $arch"
    pkg="hrp-$version-$os-$arch$(get_pkg_suffix $os)"
    url=$(get_download_url $pkg)
    echo "Selected package: $url"
    echo

    echoInfo "Created temp dir..."
    echo "$ mktemp -d -t hrp.XXXX"
    tmp_dir=$(mktemp -d -t hrp.XXXX)
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
    echo "$ ls -lh"
    ls -lh
    echo

    echoInfo "Installing..."
    if hrp -v > /dev/null; then
        echoWarn "$(hrp -v) exists, remove first !!!"
        echo "$ rm -rf $(which hrp)"
        rm -rf "$(which hrp)"
    fi

    echo "$ chmod +x hrp && mv hrp /usr/local/bin/"
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
