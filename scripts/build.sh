#!/bin/bash
# ==============================================================================
# Build Script: Compiles the microservice to WebAssembly (wasip1/wasm)
# ==============================================================================
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${ROOT_DIR}/bin"
mkdir -p "${BIN_DIR}"

echo "Building Flatcar WebAssembly Engine (wasip1/wasm)..."
cd "${ROOT_DIR}"

# Strips debug symbols (-s) and DWARF tables (-w) to keep binary lean for edge ECUs
GOOS=wasip1 GOARCH=wasm go build -ldflags="-s -w" -o "${BIN_DIR}/app.wasm" ./cmd/server/main.go

echo "✅ Build complete: $(ls -lh "${BIN_DIR}/app.wasm" | awk '{print $5, $9}')"
