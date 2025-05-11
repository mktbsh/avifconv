#!/usr/bin/env bash
set -e

OUTPUT_NAME="avifconv"
BUILD_DIR="./bin"

if [ -n "$(git tag --points-at HEAD 2>/dev/null)" ]; then
    VERSION=$(git tag --points-at HEAD | head -n 1)
elif [ -n "$(git describe --tags 2>/dev/null)" ]; then
    VERSION=$(git describe --tags)
else
    VERSION="dev"
fi

BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

mkdir -p "$BUILD_DIR"

echo "Building avifconv ${VERSION} (${GIT_COMMIT})..."

LDFLAGS=(
    "-s"
    "-w"
    "-X 'main.Version=${VERSION}'"
    "-X 'main.BuildDate=${BUILD_DATE}'"
    "-X 'main.GitCommit=${GIT_COMMIT}'"
    "-extldflags '-static'"
)

export CGO_ENABLED=0

echo "Building for $(go env GOOS)/$(go env GOARCH)..."
go build \
    -trimpath \
    -ldflags="${LDFLAGS[*]}" \
    -o "$BUILD_DIR/$OUTPUT_NAME" \
    .

BINARY_SIZE=$(du -h "$BUILD_DIR/$OUTPUT_NAME" | cut -f1)
echo "Build complete: $BUILD_DIR/$OUTPUT_NAME (${BINARY_SIZE})"
echo "Version: ${VERSION}"
echo "Git commit: ${GIT_COMMIT}"
echo "Build date: ${BUILD_DATE}"

build_cross_platform() {
    echo
    echo "Building for multiple platforms..."

    PLATFORMS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

    for PLATFORM in "${PLATFORMS[@]}"; do
        GOOS=${PLATFORM%/*}
        GOARCH=${PLATFORM#*/}
        OUTPUT_FILE="$BUILD_DIR/$OUTPUT_NAME-$GOOS-$GOARCH"

        if [ "$GOOS" = "windows" ]; then
            OUTPUT_FILE="$OUTPUT_FILE.exe"
        fi

        echo "Building for $GOOS/$GOARCH..."
        GOOS=$GOOS GOARCH=$GOARCH go build \
            -trimpath \
            -ldflags="${LDFLAGS[*]}" \
            -o "$OUTPUT_FILE" \
            .
    done

    echo "Cross-platform builds complete in $BUILD_DIR/"
    ls -lh "$BUILD_DIR"
}

# NOTE: クロスプラットフォームビルド時は下記を有効化
# build_cross_platform

echo
echo "To install avifconv to your system, run:"
echo "  sudo ./install-local.sh"
