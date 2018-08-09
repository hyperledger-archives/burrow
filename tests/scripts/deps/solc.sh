#!/usr/bin/env bash
set -e
# Static solc that will run on linux included Alpine
SOLC_URL="https://github.com/ethereum/solidity/releases/download/v0.4.24/solc-static-linux"
SOLC_BIN="$1"

wget -O "$SOLC_BIN" "$SOLC_URL"

chmod +x "$SOLC_BIN"
