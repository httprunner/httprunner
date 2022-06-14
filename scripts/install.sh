#!/bin/bash
# install hrp with one shell command
# bash -c "$(curl -ksSL https://httprunner.com/script/install.sh)"

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
    curl -ksSL https://httprunner.oss-cn-beijing.aliyuncs.com/VERSION
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

function main() {
    echoInfo "Detect target hrp package..."
    version=$(get_latest_version)
    if [[ $version != v* ]]; then
        echo "get hrp latest version failed:"
        echo "$version"
        exit 1
    fi
    echo "Latest version: $version"

    os=$(get_os)
    echo "Current OS: $os"
    if [[ $os == mingw* ]]; then
        echoWarn "Current OS is MinGW, try to use windows package"
        os="windows"
    fi

    arch=$(get_arch)
    echo "Current ARCH: $arch"
    pkg_suffix=".tar.gz"
    pkg="hrp-$version-$os-$arch$pkg_suffix"
    echo "Download package: $pkg"

    # download from aliyun OSS or github packages
    aliyun_oss_url="https://httprunner.oss-cn-beijing.aliyuncs.com/$pkg"
    github_url="https://github.com/httprunner/httprunner/releases/download/$version/$pkg"
    valid_flag=false
    for url in "$aliyun_oss_url" "$github_url"; do
        if curl --output /dev/null --silent --head --fail "$url"; then
            valid_flag=true
            break
        fi
        echoWarn "Invalid download url: $url"
    done

    if [[ "$valid_flag" == false ]]; then
        echoError "No available download url found, exit!"
        exit 1
    fi
    echo "Download url: $url"
    echo

    echoInfo "Downloading..."
    echo "$ curl -kL $url -o $pkg"
    curl -kL $url -o "$pkg"
    echo

    # for windows, only extract package to current directory
    if [[ $os == windows ]]; then # windows
        # extract to current directory
        echoInfo "Extracting..."
        echo "$ unzip -o $pkg -d ."
        unzip -o $pkg -d .

        echo "$ hrp.exe -v"
        hrp.exe -v
        echo "$ hrp.exe -h"
        hrp.exe -h
        exit 0
    fi

    # for linux or darwin, install hrp to /usr/local/bin
    # extract to temp directory
    echoInfo "Created temp dir..."
    echo "$ mktemp -d -t hrp.XXXX"
    tmp_dir=$(mktemp -d -t hrp.XXXX)
    echo "$tmp_dir"
    echo "$ mv $pkg $tmp_dir && cd $tmp_dir"
    mv $pkg $tmp_dir
    cd "$tmp_dir"
    echo

    echoInfo "Extracting..."
    echo "$ tar -xzf $pkg"
    tar -xzf "$pkg"

    echo "$ ls -lh"
    ls -lh
    echo

    echoInfo "Installing..."
    if hrp -v > /dev/null && [ $(command -v hrp) != "./hrp" ]; then
        echoWarn "$(hrp -v) exists, remove first !!!"
        echo "$ rm -rf $(command -v hrp)"
        rm -rf "$(command -v hrp)"
    fi

    echo "$ chmod +x hrp && mv hrp /usr/local/bin/"
    chmod +x hrp
    mv hrp /usr/local/bin/
    echo

    echoInfo "Check installation..."
    echo "$ command -v hrp"
    command -v hrp
    echo "$ hrp -v"
    hrp -v
    echo "$ hrp -h"
    hrp -h
    echo

    if [[ -f $HOME/.hrp/venv/bin/pip3 ]]; then
        echoInfo "Upgrade httprunner..."
        echo "$ $HOME/.hrp/venv/bin/pip3 install --upgrade httprunner==$version --index-url https://pypi.org/simple"
        $HOME/.hrp/venv/bin/pip3 install --upgrade httprunner==$version --index-url https://pypi.org/simple
    fi
}

main
