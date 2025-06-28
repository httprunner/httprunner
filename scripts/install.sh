#!/bin/bash
# Install HttpRunner (hrp) with one shell command
# Usage: bash -c "$(curl -ksSL https://httprunner.com/script/install.sh)"

set -e

# =============================================================================
# Constants and Global Variables
# =============================================================================

readonly GITHUB_API_URL="https://api.github.com/repos/httprunner/httprunner/releases/latest"
readonly INSTALL_PREFIX="/usr/local/bin"

# Global variables
VERSION=""
OS=""
ARCH=""
PACKAGE_NAME=""
DOWNLOAD_URL=""
TEMP_DIR=""

# =============================================================================
# Logging Functions
# =============================================================================

function log_error() {
    echo -e "\033[31m✘ $1\033[0m" # red
}

function log_success() {
    echo -e "\033[32m✔ $1\033[0m" # green
}

function log_warn() {
    echo -e "\033[33m! $1\033[0m" # yellow
}

function log_info() {
    echo -e "\033[34mℹ $1\033[0m" # blue
}

# =============================================================================
# Utility Functions
# =============================================================================

function cleanup() {
    if [[ -n "$TEMP_DIR" && -d "$TEMP_DIR" ]]; then
        log_info "Cleaning up temporary directory: $TEMP_DIR"
        rm -rf "$TEMP_DIR"
    fi
}

function exit_with_error() {
    log_error "$1"
    cleanup
    exit 1
}

function check_command() {
    if ! command -v "$1" &> /dev/null; then
        exit_with_error "Required command '$1' not found. Please install it first."
    fi
}

# =============================================================================
# System Detection Functions
# =============================================================================

function detect_os() {
    local os_name
    os_name=$(uname -s | tr '[:upper:]' '[:lower:]')

    case "$os_name" in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        mingw*|cygwin*|msys*)
            log_warn "Detected MinGW/Cygwin environment, using Windows package"
            OS="windows"
            ;;
        *)
            exit_with_error "Unsupported operating system: $os_name"
            ;;
    esac

    log_info "Detected OS: $OS"
}

function detect_arch() {
    local arch_name
    arch_name=$(uname -m)

    case "$arch_name" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            exit_with_error "Unsupported architecture: $arch_name"
            ;;
    esac

    log_info "Detected architecture: $ARCH"
}

# =============================================================================
# Version and Download Functions
# =============================================================================

function get_latest_version() {
    log_info "Fetching latest version from GitHub API..."

    check_command "curl"

    local response
    response=$(curl -s "$GITHUB_API_URL") || exit_with_error "Failed to fetch release information from GitHub"

    VERSION=$(echo "$response" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | head -n1)

    if [[ -z "$VERSION" || "$VERSION" != v* ]]; then
        exit_with_error "Failed to get latest version. Response: $response"
    fi

    log_success "Latest version: $VERSION"
}

function construct_download_info() {
    local pkg_suffix

    if [[ "$OS" == "windows" ]]; then
        pkg_suffix=".zip"
    else
        pkg_suffix=".tar.gz"
    fi

    PACKAGE_NAME="hrp-$VERSION-$OS-$ARCH$pkg_suffix"

    local response
    response=$(curl -s "$GITHUB_API_URL") || exit_with_error "Failed to fetch download URL"

    DOWNLOAD_URL=$(echo "$response" | grep "browser_download_url.*$PACKAGE_NAME" | cut -d '"' -f 4 | head -n1)

    if [[ -z "$DOWNLOAD_URL" ]]; then
        exit_with_error "Package not found: $PACKAGE_NAME"
    fi

    log_success "Package: $PACKAGE_NAME"
    log_success "Download URL: $DOWNLOAD_URL"
}

# =============================================================================
# Download and Installation Functions
# =============================================================================

function download_package() {
    log_info "Downloading package..."
    echo "$ curl -kL $DOWNLOAD_URL -o $PACKAGE_NAME"

    if ! curl -kL "$DOWNLOAD_URL" -o "$PACKAGE_NAME"; then
        exit_with_error "Failed to download package from $DOWNLOAD_URL"
    fi

    if [[ ! -f "$PACKAGE_NAME" ]]; then
        exit_with_error "Downloaded package not found: $PACKAGE_NAME"
    fi

    log_success "Package downloaded successfully"
}

function extract_package() {
    log_info "Extracting package..."

    if [[ "$OS" == "windows" ]]; then
        check_command "unzip"
        echo "$ unzip -o $PACKAGE_NAME -d ."
        unzip -o "$PACKAGE_NAME" -d . || exit_with_error "Failed to extract package"
    else
        check_command "tar"
        echo "$ tar -xzf $PACKAGE_NAME"
        tar -xzf "$PACKAGE_NAME" || exit_with_error "Failed to extract package"
    fi

    log_success "Package extracted successfully"
}

function install_for_windows() {
    log_success "Installation completed for Windows"
    echo
    echo "$ hrp.exe -v"
    ./hrp.exe -v || exit_with_error "Failed to verify hrp installation"
    echo
    echo "$ hrp.exe -h"
    ./hrp.exe -h
}

function install_for_unix() {
    log_info "Installing hrp to $INSTALL_PREFIX..."

    # Check if hrp already exists and remove it
    if command -v hrp &> /dev/null && [[ "$(command -v hrp)" != "./hrp" ]]; then
        local existing_version
        existing_version=$(hrp -v 2>/dev/null || echo "unknown")
        log_warn "$existing_version exists, removing first..."
        echo "$ rm -rf $(command -v hrp)"
        rm -rf "$(command -v hrp)" || exit_with_error "Failed to remove existing hrp"
    fi

    # Install new version
    echo "$ chmod +x hrp && mv hrp $INSTALL_PREFIX/"
    chmod +x hrp || exit_with_error "Failed to set executable permission"
    mv hrp "$INSTALL_PREFIX/" || exit_with_error "Failed to move hrp to $INSTALL_PREFIX (try running with sudo)"

    log_success "Installation completed"
}

function verify_installation() {
    log_info "Verifying installation..."

    local hrp_cmd="hrp"
    if [[ "$OS" == "windows" ]]; then
        hrp_cmd="./hrp.exe"
    fi

    echo "$ command -v $hrp_cmd"
    if [[ "$OS" != "windows" ]]; then
        command -v hrp || exit_with_error "hrp command not found in PATH"
    fi

    echo "$ $hrp_cmd -v"
    $hrp_cmd -v || exit_with_error "Failed to verify hrp installation"

    echo "$ $hrp_cmd -h"
    $hrp_cmd -h || exit_with_error "Failed to show hrp help"

    log_success "Installation verified successfully"
}

# =============================================================================
# Main Installation Logic
# =============================================================================

function setup_temp_directory() {
    if [[ "$OS" != "windows" ]]; then
        log_info "Creating temporary directory..."
        check_command "mktemp"

        TEMP_DIR=$(mktemp -d -t hrp.XXXX) || exit_with_error "Failed to create temporary directory"
        log_info "Temporary directory: $TEMP_DIR"

        echo "$ mv $PACKAGE_NAME $TEMP_DIR && cd $TEMP_DIR"
        mv "$PACKAGE_NAME" "$TEMP_DIR" || exit_with_error "Failed to move package to temp directory"
        cd "$TEMP_DIR" || exit_with_error "Failed to change to temp directory"
    fi
}

function main() {
    # Set up cleanup trap
    trap cleanup EXIT

    log_success "Starting HttpRunner (hrp) installation..."
    echo

    # System detection
    log_info "Detecting system information..."
    detect_os
    detect_arch
    echo

    # Version and download preparation
    get_latest_version
    construct_download_info
    echo

    # Download and extract
    download_package
    setup_temp_directory
    extract_package

    if [[ "$OS" == "windows" ]]; then
        echo "$ ls -lh"
        ls -lh
        echo
        install_for_windows
    else
        echo "$ ls -lh"
        ls -lh
        echo
        install_for_unix
        echo
        verify_installation
    fi

    echo
    log_success "HttpRunner (hrp) installation completed successfully!"
}

# =============================================================================
# Script Entry Point
# =============================================================================

main "$@"
