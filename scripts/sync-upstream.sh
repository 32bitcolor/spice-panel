#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# sync-upstream.sh — bring spice-panel current with the latest UPSTREAM release
# while keeping the fork's own UI.
#
# It merges the latest upstream release tag, then deterministically resolves:
#   • backend (cmd/, db-routines/, go.*, data, scripts …) → taken from upstream
#   • web/ (the whole frontend), README, .claude/          → kept as the fork's
#   • a small ALLOWLIST of UI-agnostic frontend paths       → taken from upstream
# then rebuilds the binary (frontend + Go, embedded) and swaps it in.
#
# It does NOT restart the service — the caller (the /update/sync handler) re-execs
# into the freshly-swapped binary, mirroring the existing self-update path.
#
# Exit codes: 0 = updated & swapped · 3 = already up to date (no-op) · other = error.
# Progress is emitted as `STEP <n>/<N> <message>` lines so the UI can show it.
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

SRC="${SPICE_SRC:-/home/ladmin/spice-panel}"
DEPLOY="${SPICE_DEPLOY:-/home/ladmin/dune-admin/dune-admin}"
GO_BIN="${GO_BIN:-/home/ladmin/go-sdk/go/bin/go}"
UPSTREAM_REPO="${UPSTREAM_REPO:-Icehunter/dune-admin}"
# Skip the "reset build clone to origin/main" step (used only for local testing).
NO_RESET="${SYNC_NO_RESET:-0}"

# UI-agnostic frontend paths safe to adopt from upstream (no rendered components).
ALLOW_FRONTEND=(web/src/locales web/src/data)

export PATH="$(dirname "$GO_BIN"):$PATH"
step() { echo "STEP $1 $2"; }
fail() { echo "ERROR $*" >&2; exit "${2:-1}"; }

cd "$SRC" || fail "build clone not found at $SRC" 2

step 1 "Fetching origin + upstream"
git fetch --quiet origin
git fetch --quiet upstream --tags
git checkout --quiet main
[ "$NO_RESET" = "1" ] || git reset --hard --quiet origin/main

step 2 "Resolving latest upstream release"
TAG="$(curl -fsSL "https://api.github.com/repos/${UPSTREAM_REPO}/releases/latest" \
  | grep -oE '"tag_name"[^,]*' | head -1 | sed -E 's/.*"([^"]+)"$/\1/')"
[ -n "$TAG" ] || fail "could not determine latest upstream release tag" 4
echo "INFO latest upstream release: $TAG"

if git merge-base --is-ancestor "$TAG" HEAD 2>/dev/null; then
  echo "INFO already up to date with $TAG"
  exit 3
fi

step 3 "Merging $TAG (keeping the spice-panel UI)"
git merge --no-commit --no-ff "$TAG" >/dev/null 2>&1 || true
# Make the frontend + docs EXACTLY the fork's — removing any files upstream added
# (stray HeroUI components / tests would otherwise break the typecheck).
git rm -rq --ignore-unmatch web README.md .claude >/dev/null 2>&1 || true
git checkout HEAD -- web README.md .claude
# Adopt only the UI-agnostic frontend allowlist from upstream.
for p in "${ALLOW_FRONTEND[@]}"; do
  if git cat-file -e "$TAG:$p" 2>/dev/null; then
    git checkout "$TAG" -- "$p" && echo "INFO took upstream $p"
  fi
done
# Version must reflect the upstream release we synced to (merge can keep 'ours').
git checkout "$TAG" -- VERSION 2>/dev/null || true
git add -A
if git diff --name-only --diff-filter=U | grep -q .; then
  echo "ERROR unresolved conflicts:" >&2; git diff --name-only --diff-filter=U >&2
  git merge --abort 2>/dev/null || git reset --hard --quiet origin/main
  fail "unexpected merge conflicts" 5
fi

restore() { git merge --abort 2>/dev/null || true; git reset --hard --quiet "${1:-origin/main}"; }

step 4 "Building frontend"
( cd web && pnpm install --frozen-lockfile && pnpm build ) >/tmp/sync-fe.log 2>&1 \
  || { tail -20 /tmp/sync-fe.log >&2; restore; fail "frontend build failed" 6; }

step 5 "Building backend"
rm -rf cmd/dune-admin/dist && cp -r web/dist cmd/dune-admin/dist
V="$(cat VERSION)"; C="$(git rev-parse --short HEAD 2>/dev/null || echo synced)"
BT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
"$GO_BIN" build -trimpath \
  -ldflags "-s -w -X main.AppVersion=$V -X main.GitCommit=$C -X main.BuildTime=$BT" \
  -tags embed -o "$SRC/dune-admin.synced" ./cmd/dune-admin >/tmp/sync-be.log 2>&1 \
  || { tail -20 /tmp/sync-be.log >&2; restore; fail "backend build failed" 7; }

step 6 "Verifying + swapping in the new binary"
"$SRC/dune-admin.synced" --version >/dev/null 2>&1 \
  || { rm -f "$SRC/dune-admin.synced"; restore; fail "built binary failed to run" 8; }
cp -p "$DEPLOY" "${DEPLOY}.prev-sync" 2>/dev/null || true
mv -f "$SRC/dune-admin.synced" "$DEPLOY"
chmod +x "$DEPLOY"

step 7 "Recording the merge"
git -c user.name="${GIT_AUTHOR:-spice-panel-sync}" \
    -c user.email="${GIT_EMAIL:-sync@spice-panel.local}" \
    commit --quiet -m "sync: merge upstream $TAG (keep spice-panel UI)" || true
if [ "${SYNC_NO_PUSH:-0}" = "1" ]; then
  echo "INFO push skipped (SYNC_NO_PUSH)"
else
  git push --quiet origin main >/dev/null 2>&1 && echo "INFO pushed merge to origin" \
    || echo "INFO push skipped (non-fatal)"
fi

echo "DONE synced to $TAG ($V @ $C) — caller should re-exec"
exit 0
