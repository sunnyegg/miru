---
name: release
description: >-
  Cut a Miru GitHub Release: bump the alpha version in CHANGELOG, merge the PR
  if needed, push an annotated v*-alpha tag, wait for release CI, rewrite the
  release notes, and verify the published binaries carry that version. Use when
  the user asks to release, ship, tag, bump version, or publish a Miru build.
argument-hint: "[patch|minor|major]"
---

# Release

Ship a tagged Miru pre-release (`vX.Y.Z-alpha`) through GitHub Actions, then
prove the artifacts carry the tag.

## Versioning

- Tags look like `v0.4.4-alpha`. The product version is the tag without the
  leading `v` (`0.4.4-alpha`); asset names and `wails.json` use it.
- Default bump is **patch** (`v0.4.3-alpha` -> `v0.4.4-alpha`). Use `minor` or
  `major` only when the user asks.
- The runtime version comes from the **git tag** in
  [`.github/workflows/release.yml`](../../../.github/workflows/release.yml)
  (`-ldflags "-X main.version=…"`). `version.go` stays `"dev"` in the tree;
  never edit it for a release.
- The human-facing bump in the repo is a new section at the top of
  [`CHANGELOG.md`](../../../CHANGELOG.md).

## Preflight

- Work on `main`, clean tree, in sync with `origin/main`.
- If the release needs a PR, wait for its CI and merge first. `ci.yml` runs only
  on pull requests, so checks must be green before the merge; a changelog-only
  push to `main` runs no lint or test job.
- `gh` is authenticated for `sunnyegg/miru`.
- Do not start `make dev` during a release.

## Workflow

### 1. Pick the next tag

```bash
git fetch --tags origin
git tag --sort=-v:refname | head -5
gh release list --limit 3
```

Collect everything unreleased, not just the last PR:

```bash
git log <latest-tag>..HEAD --oneline
```

### 2. Bump CHANGELOG

Prepend one section. Bullets describe what a user gets, one line each, and
cover every merged feature, fix, or behavior change since the previous tag.

```markdown
## vX.Y.Z-alpha

- <user-facing change>
- <user-facing change>
```

Commit with the existing convention:

```bash
git add CHANGELOG.md
git commit -F - <<'EOF'
chore(release): bump changelog to vX.Y.Z-alpha
EOF
git push origin main
```

If the changelog commit lives on the PR branch, merge after CI:

```bash
gh pr checks <N> --watch --fail-fast
gh pr merge <N> --merge
```

Use `--delete-branch` only when the user wants the branch cleaned up.

### 3. Tag and push

```bash
git checkout main
git pull origin main
git tag -a vX.Y.Z-alpha -m "vX.Y.Z-alpha"
git push origin vX.Y.Z-alpha
```

The tag must point at the changelog commit. Pushing it starts the release
workflow. Do not force-push tags and do not create the GitHub release by hand;
`softprops/action-gh-release` makes it.

### 4. Wait for release CI

```bash
gh run list --workflow=release.yml --limit 1
gh run watch <run-id> --exit-status
gh run view <run-id> --json jobs
```

Expect three jobs: `Build Linux + Windows`, `Build macOS (universal)`, and
`Attach binaries to release`. A release takes roughly 5 to 10 minutes.

### 5. Rewrite the release notes

The workflow leaves a Downloads-only stub. Replace it, matching the shape of
the previous release (`gh release view <prev-tag> --json body`):

```bash
cat > /tmp/miru-release-notes.md <<'EOF'
## What's new in vX.Y.Z-alpha

- **Feature**: one sentence on what the user can do now.
- **Fix**: one sentence on what stopped going wrong.

## Downloads

- `miru-X.Y.Z-alpha-linux-amd64`: Linux x86_64 standalone ELF.
- `miru-X.Y.Z-alpha-windows-amd64.exe`: Windows x86_64 standalone with embedded WebView2.
- `miru-X.Y.Z-alpha-mac-universal.zip`: macOS universal `.app` bundle (Apple Silicon + Intel). **Unsigned**, see README for the Gatekeeper workaround.

MPV must be on `PATH` or set in Settings → Playback.
EOF
gh release edit vX.Y.Z-alpha --notes-file /tmp/miru-release-notes.md
```

Keep the release out of draft and **out of prerelease**, like the earlier alpha
releases. Keep the macOS unsigned sentence; it is the only install warning.

### 6. Verify the published artifacts

A green workflow is not proof that the binary has the right version.

```bash
gh release view vX.Y.Z-alpha --json tagName,isPrerelease,isDraft,assets
tmp=$(mktemp -d)
gh release download vX.Y.Z-alpha --pattern '*-linux-amd64' --dir "$tmp"
file "$tmp"/*
strings "$tmp"/*-linux-amd64 | grep -F 'main.version=vX.Y.Z-alpha'
rm -rf "$tmp"
```

Expect `isPrerelease: false`, three assets with sizes, an ELF 64-bit x86-64
file, and the version string in the binary. When the release adds a Wails
binding, grep one new method name as a second check that the build is current.

Finish the release note edit for macOS too: `gh release view` should show the
rewritten body, not the stub.

## If CI fails

Fix the cause on a branch, run the checks for the changed area, and ship a new
tag. Moving or deleting a published tag needs the user's explicit approval.

## Report

Give the release URL, the tag and its commit, the version bump, the jobs that
passed, each asset with its size, and the verification output you actually ran.
Say plainly what you did not do: merge request review, labels, branch cleanup,
or a visual check of the notes page.
