#!/usr/bin/env bash
set -e
SOLANG_URL="https://github.com/hyperledger-labs/solang/releases/download/v0.1.7/solang-linux"
SOLANG_BIN="$1"

wget -O "$SOLANG_BIN" "$SOLANG_URL"

chmod +x "$SOLANG_BIN"
