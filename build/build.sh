#!/bin/bash

# Build script for Bushtalk Radio X-Plane Client
# Requires: Go 1.21+, Fyne dependencies (see https://developer.fyne.io/started/)

set -e

cd "$(dirname "$0")/.."

VERSION=${VERSION:-"1.0.0"}
OUTPUT_DIR="./dist"

mkdir -p "$OUTPUT_DIR"

echo "Building Bushtalk X-Plane Client v${VERSION}..."

# Windows (64-bit)
echo "Building for Windows..."
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
    go build -ldflags="-H windowsgui -s -w" -o "${OUTPUT_DIR}/bushtalk-xplane-windows-amd64.exe" .

# macOS Intel
echo "Building for macOS (Intel)..."
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
    go build -ldflags="-s -w" -o "${OUTPUT_DIR}/bushtalk-xplane-darwin-amd64" .

# macOS Apple Silicon
echo "Building for macOS (Apple Silicon)..."
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
    go build -ldflags="-s -w" -o "${OUTPUT_DIR}/bushtalk-xplane-darwin-arm64" .

# Linux
echo "Building for Linux..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o "${OUTPUT_DIR}/bushtalk-xplane-linux-amd64" .

echo ""
echo "Build complete! Binaries in ${OUTPUT_DIR}:"
ls -la "$OUTPUT_DIR"
