#!/usr/bin/env bash
set -euo pipefail

BIN=./bin/golangci-lint
VERSION="v$(grep "^golangci-lint" .tool-versions | awk '{print $2}')"

install_linter() {
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/main/install.sh | sh -s -- -b ./bin "$VERSION"
}

if [ ! -f "$BIN" ]; then
  echo "golangci-lint not found, installing $VERSION..."
  install_linter
elif [ "v$("$BIN" --version | awk '{print $4}')" != "$VERSION" ]; then
  echo "updating golangci-lint to $VERSION"
  install_linter
fi

if [ $# -eq 0 ]; then
  "$BIN" run --timeout 2m00s --max-same-issues 0 --concurrency 5
else
  "$BIN" "$@"
fi
