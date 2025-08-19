#!/usr/bin/env bash
#
# Script to update the AUR PKGBUILD with the correct version and checksums
# Usage: ./update-pkgbuild.sh <version>
#
# This script should be run after a GitHub release is created to update
# the PKGBUILD with the correct version and SHA256 checksums.

set -euo pipefail

# Check if version is provided
if [ $# -ne 1 ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 0.14.0"
    exit 1
fi

VERSION="$1"
PKGBUILD_PATH="$(dirname "$0")/PKGBUILD"

# Remove 'v' prefix if present
VERSION="${VERSION#v}"

echo "Updating PKGBUILD for version ${VERSION}..."

# Check if PKGBUILD exists
if [ ! -f "$PKGBUILD_PATH" ]; then
    echo "Error: PKGBUILD not found at $PKGBUILD_PATH"
    exit 1
fi

# Update version in PKGBUILD
sed -i "s/^pkgver=.*/pkgver=${VERSION}/" "$PKGBUILD_PATH"
echo "✓ Updated pkgver to ${VERSION}"

# Download archives and calculate checksums
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

echo "Downloading release archives..."

# Download x86_64 archive
X86_64_URL="https://github.com/axiomhq/cli/releases/download/v${VERSION}/axiom_${VERSION}_linux_amd64.tar.gz"
X86_64_FILE="$TEMP_DIR/axiom_${VERSION}_linux_amd64.tar.gz"
echo "  Downloading x86_64 archive..."
if ! curl -L -o "$X86_64_FILE" "$X86_64_URL" 2>/dev/null; then
    echo "Error: Failed to download x86_64 archive from $X86_64_URL"
    exit 1
fi
X86_64_SHA256=$(sha256sum "$X86_64_FILE" | cut -d' ' -f1)
echo "  ✓ x86_64 SHA256: $X86_64_SHA256"

# Download aarch64 archive
AARCH64_URL="https://github.com/axiomhq/cli/releases/download/v${VERSION}/axiom_${VERSION}_linux_arm64.tar.gz"
AARCH64_FILE="$TEMP_DIR/axiom_${VERSION}_linux_arm64.tar.gz"
echo "  Downloading aarch64 archive..."
if ! curl -L -o "$AARCH64_FILE" "$AARCH64_URL" 2>/dev/null; then
    echo "Error: Failed to download aarch64 archive from $AARCH64_URL"
    exit 1
fi
AARCH64_SHA256=$(sha256sum "$AARCH64_FILE" | cut -d' ' -f1)
echo "  ✓ aarch64 SHA256: $AARCH64_SHA256"

# Update checksums in PKGBUILD
sed -i "s/^sha256sums_x86_64=.*/sha256sums_x86_64=('${X86_64_SHA256}')/" "$PKGBUILD_PATH"
sed -i "s/^sha256sums_aarch64=.*/sha256sums_aarch64=('${AARCH64_SHA256}')/" "$PKGBUILD_PATH"

echo "✓ Updated SHA256 checksums"

# Optionally validate the PKGBUILD
if command -v makepkg >/dev/null 2>&1; then
    echo "Validating PKGBUILD..."
    if (cd "$(dirname "$PKGBUILD_PATH")" && makepkg --printsrcinfo > /dev/null 2>&1); then
        echo "✓ PKGBUILD validation successful"
    else
        echo "⚠ PKGBUILD validation failed (this might be normal if not on Arch Linux)"
    fi
else
    echo "ℹ makepkg not found, skipping PKGBUILD validation"
fi

echo ""
echo "PKGBUILD successfully updated for version ${VERSION}!"
echo ""
echo "Next steps:"
echo "1. Review the changes in $PKGBUILD_PATH"
echo "2. Test the package build locally if on Arch Linux:"
echo "   cd $(dirname "$PKGBUILD_PATH") && makepkg -si"
echo "3. Commit and push to the AUR repository:"
echo "   cd /path/to/aur/axiom-bin"
echo "   git add PKGBUILD .SRCINFO"
echo "   git commit -m \"Update to version ${VERSION}\""
echo "   git push"
