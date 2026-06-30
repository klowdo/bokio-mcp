#!/usr/bin/env bash
set -euo pipefail

# ponytail: recompute vendorHash via fakeHash + grep of the nix build error.
# ceiling: depends on nix's "got: sha256-..." wording; swap to nix-update if it changes.

cd "$(git rev-parse --show-toplevel)"

FLAKE=flake.nix
FAKE="sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
OPEN_PR=false
[ "${1:-}" = "--pr" ] && OPEN_PR=true

current=$(grep -oP 'vendorHash = "\K[^"]+' "$FLAKE")
[ -n "$current" ] || { echo "could not find vendorHash in $FLAKE" >&2; exit 1; }

sed -i "s|vendorHash = \"$current\"|vendorHash = \"$FAKE\"|" "$FLAKE"

got=$(nix build .#default 2>&1 | grep -oP 'got:\s+\Ksha256-\S+' || true)

if [ -z "$got" ]; then
  sed -i "s|vendorHash = \"$FAKE\"|vendorHash = \"$current\"|" "$FLAKE"
  echo "failed to determine new vendorHash (no hash mismatch from nix build)" >&2
  exit 1
fi

sed -i "s|vendorHash = \"$FAKE\"|vendorHash = \"$got\"|" "$FLAKE"

if [ "$got" = "$current" ]; then
  echo "vendorHash unchanged ($current)"
  exit 0
fi

echo "vendorHash updated: $current -> $got"

$OPEN_PR || exit 0

BRANCH=chore/update-vendor-hash
git switch -C "$BRANCH"
git add "$FLAKE"
git commit -m "chore: update nix vendorHash"
git push -u origin "$BRANCH" --force-with-lease
gh pr create --fill \
  --title "chore: update nix vendorHash" \
  --body "Automated \`vendorHash\` update for the Go module vendor tree."
