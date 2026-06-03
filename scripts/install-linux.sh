#!/usr/bin/env bash
#
# Jamshid installation script for Linux systems.
# Automatically detects package manager and installs the appropriate package.
# Falls back to binary download if no supported package manager is found.
#

set -euo pipefail

REPO="PapaDanielVi/jamshid"
BINARY_NAME="jamshid"

# Detect the system architecture (for packages: linux_amd64, linux_arm64).
detect_arch_pkg() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64) echo "linux_amd64" ;;
        aarch64 | arm64) echo "linux_arm64" ;;
        *) echo "unsupported" ;;
    esac
}

# Detect the system architecture (for tarballs: Linux_x86_64, Linux_arm64).
detect_arch_tarball() {
    local arch
    arch=$(uname -m)
    case "$arch" in
        x86_64) echo "Linux_x86_64" ;;
        aarch64 | arm64) echo "Linux_arm64" ;;
        *) echo "unsupported" ;;
    esac
}

# Get the latest release tag from GitHub API.
get_latest_release() {
    curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep -oP '"tag_name": "\K[^"]+'
}

# Detect package manager.
detect_package_manager() {
    if command -v apk &> /dev/null; then
        echo "apk"
    elif command -v apt-get &> /dev/null; then
        echo "deb"
    elif command -v dnf &> /dev/null || command -v yum &> /dev/null; then
        echo "rpm"
    elif command -v pacman &> /dev/null; then
        echo "pacman"
    else
        echo "none"
    fi
}

# Install using Alpine/apk.
install_apk() {
    local arch="$1"
    local release="$2"
    local url="https://github.com/${REPO}/releases/download/${release}/jamshid_${release#v}_${arch}.apk"

    echo "Downloading .apk package for ${arch}..."
    curl -Lo /tmp/jamshid.apk "$url"
    sudo apk add --allow-untrusted /tmp/jamshid.apk
    rm -f /tmp/jamshid.apk
}

# Install using Debian/Ubuntu.
install_deb() {
    local arch="$1"
    local release="$2"
    local url="https://github.com/${REPO}/releases/download/${release}/jamshid_${release#v}_${arch}.deb"

    echo "Downloading .deb package for ${arch}..."
    curl -Lo /tmp/jamshid.deb "$url"
    sudo apt-get install -y /tmp/jamshid.deb
    rm -f /tmp/jamshid.deb
}

# Install using RPM-based systems (Fedora, RHEL, CentOS).
install_rpm() {
    local arch="$1"
    local release="$2"
    local url="https://github.com/${REPO}/releases/download/${release}/jamshid_${release#v}_${arch}.rpm"

    echo "Downloading .rpm package for ${arch}..."
    curl -Lo /tmp/jamshid.rpm "$url"
    if command -v dnf &> /dev/null; then
        sudo dnf install -y /tmp/jamshid.rpm
    else
        sudo yum install -y /tmp/jamshid.rpm
    fi
    rm -f /tmp/jamshid.rpm
}

# Install using Pacman (Arch Linux, Manjaro).
install_pacman() {
    local arch="$1"
    local release="$2"
    local url="https://github.com/${REPO}/releases/download/${release}/jamshid_${release#v}_${arch}.pkg.tar.zst"

    echo "Downloading .pkg.tar.zst package for ${arch}..."
    curl -Lo /tmp/jamshid.pkg.tar.zst "$url"
    sudo pacman -U --noconfirm /tmp/jamshid.pkg.tar.zst
    rm -f /tmp/jamshid.pkg.tar.zst
}

# Install using package manager.
install_package() {
    local pkg_manager="$1"
    local arch="$2"
    local release="$3"

    case "$pkg_manager" in
        apk) install_apk "$arch" "$release" ;;
        deb) install_deb "$arch" "$release" ;;
        rpm) install_rpm "$arch" "$release" ;;
        pacman) install_pacman "$arch" "$release" ;;
    esac
}

# Install from binary tarball (fallback).
install_binary() {
    local arch="$1"
    local release="$2"
    local url="https://github.com/${REPO}/releases/download/${release}/jamshid_${arch}.tar.gz"
    local install_dir="${INSTALL_DIR:-/usr/local/bin}"

    echo "Downloading binary tarball for ${arch}..."
    curl -Lo /tmp/jamshid.tar.gz "$url"

    echo "Extracting and installing to ${install_dir}..."
    tar -xzf /tmp/jamshid.tar.gz -C /tmp
    sudo install -m 755 /tmp/jamshid "$install_dir/jamshid"
    rm -f /tmp/jamshid.tar.gz /tmp/jamshid
}

# Check if jamshid is already installed and update to latest if so.
check_installed() {
    if command -v jamshid &> /dev/null; then
        echo "Jamshid is already installed: $(jamshid --version 2>/dev/null || echo 'installed')"
        return 0
    fi
    return 1
}

# Main installation logic.
main() {
    local arch pkg_manager release

    arch=$(detect_arch_pkg)
    if [[ "$arch" == "unsupported" ]]; then
        echo "Error: Unsupported architecture $(uname -m)"
        exit 1
    fi

    pkg_manager=$(detect_package_manager)
    release=$(get_latest_release)

    echo "Installing jamshid ${release} for ${arch} using ${pkg_manager}..."

    if [[ "$pkg_manager" != "none" ]]; then
        install_package "$pkg_manager" "$arch" "$release"
    else
        local arch_tarball
        arch_tarball=$(detect_arch_tarball)
        install_binary "$arch_tarball" "$release"
    fi

    echo "Jamshid installed successfully!"
    jamshid --version
}

# Run installation.
main "$@"