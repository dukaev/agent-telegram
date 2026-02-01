#!/bin/bash

# Build binaries for all supported platforms
# Run from project root: ./scripts/build-all.sh

set -e

VERSION=${1:-$(node -p "require('./package.json').version")}
OUTPUT_DIR="dist"
BINARY_NAME="agent-telegram"

# Platforms to build for
PLATFORMS=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
)

echo "Building ${BINARY_NAME} v${VERSION}..."
mkdir -p "${OUTPUT_DIR}"

for PLATFORM in "${PLATFORMS[@]}"; do
  GOOS="${PLATFORM%/*}"
  GOARCH="${PLATFORM#*/}"

  OUTPUT="${OUTPUT_DIR}/${BINARY_NAME}-${GOOS}-${GOARCH}"
  if [ "$GOOS" = "windows" ]; then
    OUTPUT="${OUTPUT}.exe"
  fi

  echo "Building ${GOOS}/${GOARCH}..."
  GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w -X main.version=${VERSION}" -o "$OUTPUT" .
done

echo ""
echo "Built binaries:"
ls -lh "${OUTPUT_DIR}/"
