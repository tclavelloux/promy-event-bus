#!/bin/sh
# Wrapper around golangci-lint that pins the fleet-wide version from
# .golangci-version, installing it on demand — the same on-demand-bootstrap
# shape as .githooks/pre-push's go-test-coverage install. Used by
# `make lint`, `make git-check`, and the golangci-lint* pre-commit hooks so
# all three invoke byte-identical tooling.
set -e

REPO_ROOT=$(git rev-parse --show-toplevel)

if [ -n "$GOLANGCI_LINT_VERSION" ]; then
  VERSION="$GOLANGCI_LINT_VERSION"
elif [ -f "$REPO_ROOT/.golangci-version" ]; then
  VERSION=$(cat "$REPO_ROOT/.golangci-version")
else
  echo "error: golangci-lint version not set. Set \$GOLANGCI_LINT_VERSION or add $REPO_ROOT/.golangci-version" >&2
  exit 1
fi

# Normalize: strip a leading 'v' if present, then re-add it consistently.
VERSION=${VERSION#v}
VERSION=$(printf '%s' "$VERSION" | tr -d '[:space:]')

GOBIN=$(go env GOPATH)/bin

need_install=1
if [ -x "$GOBIN/golangci-lint" ]; then
  installed_version=$("$GOBIN/golangci-lint" --version 2>/dev/null | grep -o "version v\{0,1\}[0-9][0-9.]*" | grep -o "[0-9][0-9.]*" | head -n1)
  if [ "$installed_version" = "$VERSION" ]; then
    need_install=0
  fi
fi

if [ "$need_install" = "1" ]; then
  echo "Installing golangci-lint v$VERSION..."
  go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v$VERSION"
fi

exec "$GOBIN/golangci-lint" "$@"
