#!/bin/bash
set -e

OUTPUT_DIR="build"
OUTPUT_BIN="$OUTPUT_DIR/wscollector"

echo "[build] cleaning previous build..."
rm -f "$OUTPUT_BIN"

echo "[build] building binary..."
mkdir -p "$OUTPUT_DIR"
go build -o "$OUTPUT_BIN" ./cmd/collector

echo "[build] done. binary at: $OUTPUT_BIN"
