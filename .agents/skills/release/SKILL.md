---
name: release
description: >-
  Cut a Miru GitHub Release: bump the alpha patch in CHANGELOG, merge the PR if
  needed, push an annotated v*-alpha tag, wait for release CI, then rewrite the
  GitHub release notes. Use when the user asks to release, ship, tag, bump
  version, or publish a Miru build.
argument-hint: "[patch|minor|major]"
---

# Release

Ship a tagged Miru pre-release (`vX.Y.Z-alpha`) through GitHub Actions.

## Versioning

- Tags look like `v0.4.1-alpha`. The product version (asset names) is the tag
  without the leading `v` (`0.4.1-alpha`).
- Default bump is **patch** (`v0.4.0-alpha` → `v0.4.1-alpha`). Parse
  `$ARGUMENTS` for `minor` or `major` when the user asks; otherwise patch.
- Runtime version is stamped from the **git tag** in
  [`.github/workflows/release.yml`](../../../.github/workflows/release.yml)
  (`-ldflags "-X main.version=…"`). `version.go` stays `"dev"` in the tree.
- Human-facing bump in-repo: add a section at the top of [`CHANGELOG.md`](../../../CHANGELOG.md).

## Prerequisites

- Clean enough git state to commit the changelog (or a ready PR branch).
- `gh` authenticated for `sunnyegg/miru`.
- Do **not** start `make dev` for a release.

## Workflow

### 1. Decide the next tag

```bash
git fetch --tags origin
git tag --sort=-v:refname | head -5
gh release list --limit 3
```

Next tag = latest `v*-alpha` bumped per `$ARGUMENTS` (default patch).

Collect unreleased commits for notes:

```bash
git log <latest-tag>..HEAD --oneline
```

Include every merged feature/fix since the previous tag in CHANGELOG and the
GitHub release body (not only the latest PR).

### 2. Bump CHANGELOG

Prepend a section to `CHANGELOG.md`:

```markdown
## vX.Y.Z-alpha

- <user-facing bullets; match release notes tone>
```

Commit on the feature branch (preferred) or on `main` if the work is already
merged:

```bash
git add CHANGELOG.md
git commit -m "$(cat <<'EOF'
chore(release): bump changelog to vX.Y.Z-alpha

EOF
)"
git push
```

If this is still on a PR: wait for CI, then merge (merge commit, match recent
history):

```bash
gh pr checks <N> --watch --fail-fast
gh pr merge <N> --merge --delete-branch
```

### 3. Tag main and push

```bash
git checkout main
git pull origin main
git tag -a vX.Y.Z-alpha -m "vX.Y.Z-alpha"
git push origin vX.Y.Z-alpha
```

Pushing the tag starts [`.github/workflows/release.yml`](../../../.github/workflows/release.yml).
Do not force-push tags.

### 4. Wait for release CI

```bash
gh run list --workflow=release.yml --limit 1
gh run watch <run-id> --exit-status
```

Expect Linux, Windows, and macOS artifacts on the GitHub Release. Softprops
creates a stub body (Downloads only) — replace it next.

### 5. Update the release description

Match the previous release shape (`gh release view <prev-tag> --json body`):

```bash
gh release edit vX.Y.Z-alpha --notes "$(cat <<'EOF'
## Miru vX.Y.Z-alpha

### What's new

- <same bullets as CHANGELOG, polished for users>

### Downloads

- `miru-X.Y.Z-alpha-linux-amd64` — Linux x86_64 standalone ELF.
- `miru-X.Y.Z-alpha-windows-amd64.exe` — Windows x86_64 standalone with embedded WebView2.
- `miru-X.Y.Z-alpha-mac-universal.zip` — macOS universal `.app` bundle (Apple Silicon + Intel). **Unsigned** — see README for the Gatekeeper workaround.
EOF
)"
```

Return the release URL: `https://github.com/sunnyegg/miru/releases/tag/vX.Y.Z-alpha`.

## Notes

- CI release body is auto-generated; always rewrite with **What's new**.
- macOS build is unsigned; keep the Gatekeeper sentence in Downloads.
- If release CI fails, fix on a branch, retag only after the user agrees (moving
  tags needs an explicit ask).
- Link any PR you merge for this thread with T3 `link_pull_request` when that
  tool is available.
