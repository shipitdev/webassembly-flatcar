#!/bin/bash
# Simple build script to compile Go to wasip1 WebAssembly
set -euo pipefail

cd "$(dirname "$0")"

echo "Compiling main.go to WebAssembly (wasip1/wasm)..."
GOOS=wasip1 GOARCH=wasm go build -ldflags="-s -w" -o app.wasm main.go

echo "Build complete: $(ls -lh app.wasm | awk '{print $5, $9}')"
