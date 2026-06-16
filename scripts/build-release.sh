#!/usr/bin/env bash
# Build & package release binaries for GitHub Releases
# Usage: ./scripts/build-release.sh [v0.1.0]

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

VERSION="${1:-}"
if [[ -z "$VERSION" ]]; then
    VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo "dev")"
fi
VER="${VERSION#v}"

echo "Building Whatsar release ${VERSION} (package ${VER})..."

make clean-dist 2>/dev/null || true
make release

DIST="${ROOT}/dist"
mkdir -p "$DIST/packages"

package_linux() {
    local arch="$1" bin="$2"
    local name="whatsar_${VER}_linux_${arch}"
    cp "${DIST}/${bin}" "${DIST}/whatsar"
    chmod +x "${DIST}/whatsar"
    tar -czf "${DIST}/packages/${name}.tar.gz" -C "${DIST}" whatsar
    rm -f "${DIST}/whatsar"
    echo "  → packages/${name}.tar.gz"
}

package_windows() {
    local arch="$1" bin="$2"
    local name="whatsar_${VER}_windows_${arch}"
    local tmp="${DIST}/winpack"
    rm -rf "$tmp"
    mkdir -p "$tmp"
    cp "${DIST}/${bin}" "${tmp}/whatsar.exe"
    (cd "$tmp" && zip -q "../packages/${name}.zip" whatsar.exe)
    rm -rf "$tmp"
    echo "  → packages/${name}.zip"
}

package_linux amd64 whatsar_linux_amd64
package_linux arm64 whatsar_linux_arm64
package_linux armv7 whatsar_linux_armv7
package_windows amd64 whatsar_windows_amd64.exe

(
    cd "${DIST}/packages"
    sha256sum *.tar.gz *.zip 2>/dev/null > SHA256SUMS || shasum -a 256 *.tar.gz *.zip > SHA256SUMS
)

echo ""
echo "Release packages ready in dist/packages/"
ls -lh "${DIST}/packages/"