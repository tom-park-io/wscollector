#!/bin/bash
set -e

BUILD_PATH="./build"
BIN_PATH=$BUILD_PATH/wscollector

if [ ! -f "$BIN_PATH" ]; then
    echo "[run] binary not found, building first..."
    ./scripts/build.sh
fi

echo "[run] starting binary..."
cd $BUILD_PATH
./wscollector "$@"
