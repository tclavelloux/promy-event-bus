#!/bin/sh
# Non-destructive go-mod-tidy check for the pre-commit local hook. Replaces
# the archived dnephin/pre-commit-golang go-mod-tidy hook: it verifies
# go.mod/go.sum are tidy WITHOUT mutating the working tree — if `go mod tidy`
# would change either file, it restores the originals and fails with
# instructions, rather than leaving surprise edits staged for commit.
set -e

REPO_ROOT=$(git rev-parse --show-toplevel)
cd "$REPO_ROOT"

MOD_BACKUP=$(mktemp)
SUM_BACKUP=$(mktemp)

cp go.mod "$MOD_BACKUP"
cp go.sum "$SUM_BACKUP"

# Restore on ANY exit path, including a `go mod tidy` failure or an interrupt —
# otherwise a partially-tidied tree is left behind for the developer to discover.
restore() {
  cp "$MOD_BACKUP" go.mod
  cp "$SUM_BACKUP" go.sum
  rm -f "$MOD_BACKUP" "$SUM_BACKUP"
}
trap restore EXIT INT TERM

go mod tidy

if ! diff -q "$MOD_BACKUP" go.mod >/dev/null 2>&1 || ! diff -q "$SUM_BACKUP" go.sum >/dev/null 2>&1; then
  echo "error: go.mod/go.sum are not tidy. Run 'go mod tidy' manually and commit the result." >&2
  exit 1
fi

exit 0
