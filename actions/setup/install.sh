#!/usr/bin/env bash
#
# Install a released Axiom CLI binary and add it to PATH. Driven by the
# INPUT_VERSION environment variable, set from the action's `version` input.

set -euo pipefail

readonly repo="axiomhq/cli"
: "${RUNNER_TOOL_CACHE:=${RUNNER_TEMP:-/tmp}/tool-cache}"

die() {
	echo "::error::$*"
	exit 1
}

case "${INPUT_VERSION:-latest}" in
latest)
	# The releases/latest page redirects to the tag. Reading the redirect
	# target keeps this off api.github.com, which is rate limited per runner
	# IP and would otherwise need a token.
	url=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/${repo}/releases/latest")
	tag=${url##*/}
	[[ $tag == v* ]] || die "cannot resolve latest release, got '${tag}' from ${url}"
	;;
v*) tag=${INPUT_VERSION} ;;
*) tag=v${INPUT_VERSION} ;;
esac
readonly tag
readonly version=${tag#v}

case "${RUNNER_OS}" in
Linux) os=linux ;;
macOS) os=darwin ;;
Windows) os=windows ;;
*) die "unsupported runner OS '${RUNNER_OS}'" ;;
esac
readonly os

case "${RUNNER_ARCH}" in
X64) arch=amd64 ;;
ARM64) arch=arm64 ;;
*) die "unsupported runner architecture '${RUNNER_ARCH}', see https://github.com/${repo}/releases for available builds" ;;
esac
readonly arch

if [[ $os == windows && $arch == arm64 ]]; then
	die "there is no windows/arm64 build, see https://github.com/${repo}/releases"
fi

if [[ $os == windows ]]; then
	readonly ext=zip binary=axiom.exe
else
	readonly ext=tar.gz binary=axiom
fi

# The tool cache layout matches @actions/tool-cache: the entry lives at
# <tool>/<version>/<arch> with a sibling <arch>.complete marker, and the arch
# is the Node os.arch() spelling rather than the Go one.
dest="${RUNNER_TOOL_CACHE}/axiom/${version}/$(echo "${RUNNER_ARCH}" | tr '[:upper:]' '[:lower:]')"
# RUNNER_TOOL_CACHE is a backslash path on Windows, which bash reads as escapes.
readonly dest=${dest//\\//}

publish() {
	echo "${dest}" >>"${GITHUB_PATH}"
	echo "version=${tag}" >>"${GITHUB_OUTPUT}"
}

if [[ -f ${dest}.complete ]]; then
	echo "axiom ${tag} found in the tool cache"
	publish
	exit 0
fi

tmp=$(mktemp -d)
readonly tmp
trap 'rm -rf "${tmp}"' EXIT

readonly archive="axiom_${version}_${os}_${arch}.${ext}"
readonly base="https://github.com/${repo}/releases/download/${tag}"

echo "Downloading ${archive} from ${tag}"
curl -fsSL --retry 3 -o "${tmp}/${archive}" "${base}/${archive}" ||
	die "cannot download ${base}/${archive}"
curl -fsSL --retry 3 -o "${tmp}/checksums.txt" "${base}/checksums.txt" ||
	die "cannot download ${base}/checksums.txt"

(
	cd "${tmp}"
	grep " ${archive}\$" checksums.txt >checksum ||
		die "${archive} is not listed in checksums.txt"
	if command -v sha256sum >/dev/null; then
		sha256sum -c checksum
	else
		shasum -a 256 -c checksum
	fi
) || die "checksum verification failed for ${archive}"

if [[ $ext == zip ]]; then
	# Git Bash ships no unzip and its tar is GNU tar, which cannot read zip.
	powershell -NoProfile -NonInteractive -Command \
		"Expand-Archive -LiteralPath '$(cygpath -w "${tmp}/${archive}")' -DestinationPath '$(cygpath -w "${tmp}")' -Force"
else
	tar -xzf "${tmp}/${archive}" -C "${tmp}"
fi

# The nix archives wrap their contents in a directory, the Windows one does not.
src=$(find "${tmp}" -type f -name "${binary}" -print -quit)
[[ -n $src ]] || die "${binary} not found in ${archive}"

mkdir -p "${dest}"
mv "${src}" "${dest}/${binary}"
chmod +x "${dest}/${binary}"
: >"${dest}.complete"

echo "Installed axiom ${tag} to ${dest}"
publish
