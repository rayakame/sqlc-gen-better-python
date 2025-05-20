#!/bin/bash
set -euo pipefail

# ──────────────────────────────
# 1) CONFIGURATION
# ──────────────────────────────
TARGET_DIRS=("test/driver_asyncpg", "test\driver_aiosqlite")
SQLC_CONFIG_NAMES=("sqlc.yaml")

# ──────────────────────────────
# 2) BUILD THE WASM PLUGIN
# ──────────────────────────────
echo "=== Building the Go WASM plugin ================================="
export GOOS=wasip1
export GOARCH=wasm
go build -o sqlc-gen-better-python.wasm plugin/main.go

# ──────────────────────────────
# 3) CALCULATE SHA‑256
# ──────────────────────────────
SHA256_HASH=$(sha256sum sqlc-gen-better-python.wasm | awk '{print $1}')
echo "SHA-256: $SHA256_HASH"

# ──────────────────────────────
# 4) UPDATE ROOT yaml
# ──────────────────────────────
echo "Patching root sqlc.yaml..."
sed -i -E "s/(sha256: )\S+/\1$SHA256_HASH/" sqlc.yaml

# ──────────────────────────────
# 5) PROPAGATE TO TARGET FOLDERS
# ──────────────────────────────
for dir in "${TARGET_DIRS[@]}"; do
    echo "--------------------------------------------------------------"
    echo "  Processing $dir"
    mkdir -p "$dir"

    cp -f sqlc-gen-better-python.wasm "$dir/"

    for file in "${SQLC_CONFIG_NAMES[@]}"; do
        if [[ -f "$dir/$file" ]]; then
            echo "  Patching $dir/$file"
            sed -i -E "s/(sha256: )\S+/\1$SHA256_HASH/" "$dir/$file"
        else
            echo "  Warning: $dir/$file not found"
        fi
    done
done

echo "=== All done - every sqlc.yaml now has SHA-256 $SHA256_HASH ======"
