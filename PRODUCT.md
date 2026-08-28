# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Linux desktop anime viewers who want to manage and watch local anime files, download episodes, and keep their watch progress synchronized with AniList.

## Product Purpose

Miru is a portable desktop application for playing and managing anime media. It combines local library management, MPV playback, built-in torrent downloading, and AniList synchronization in one application. Product success means a user can discover or add anime, download or open an episode, watch it, and maintain accurate viewing progress without assembling several external tools.

The full product scope includes the capabilities described in `PRD.md`. The current MVP is narrower; the implemented scope and deferred work are recorded in `README.md`.

## Positioning

Miru is a portable, resource-conscious anime media client that keeps playback, downloading, and AniList progress in one desktop application. Its portable operation and native Go-based engine are central product commitments.

## Operating Context

- The application runs as a portable binary on a user's desktop and uses an embedded Wails webview for its interface.
- MPV is an external runtime dependency, discovered from the system or selected manually by the user.
- Anime files may be local files or files obtained through magnet links and `.torrent` downloads.
- AniList is accessed through its GraphQL API after browser-based account authorization.
- Configuration and local media data are stored in the user's platform-specific configuration and cache directories.

## Capabilities and Constraints

- Provide a local anime library, playback through MPV, download management, and AniList watch-progress synchronization.
- Support native BitTorrent handling without requiring a separate torrent client.
- Parse media filenames into useful anime metadata such as title, episode number, resolution, and fansub group.
- Provide settings for MPV path, download location, sync threshold, network mode, and transfer limits.
- The PRD also defines RSS indexing, an airing calendar, seeding policy, bandwidth controls, and cross-platform packaging; these remain active target capabilities even where the MVP is deferred.
- Use Wails v2 with a Go backend and React/Vite frontend.
- Keep the Go implementation CGO-free where required for cross-platform builds.
- Support portable builds for Linux, Windows, and macOS as a product target.
- Store AniList credentials using the platform secret service when available, with the documented local fallback protected by mode `0600`.
- The PRD aspires to idle memory below 100 MB; the Linux benchmark (2026-08-28, production build) measures approximately **684 MB** for the application process tree — see `docs/benchmarks/idle-ram.md`.

## Brand Commitments

- The product name is Miru (見る).
- The product should preserve a portable, efficient, self-contained experience.
- Existing user-facing terminology includes MPV, AniList, library, watching, search, downloads, calendar, and settings.

## Evidence on Hand

- `README.md` documents the current Linux MVP, first-run workflow, runtime dependencies, data locations, and deferred features.
- `docs/benchmarks/idle-ram.md` documents the idle RAM benchmark methodology and measured Linux results (2026-08-28).
- `PRD.md` documents the intended product scope, architecture, functional requirements, database direction, non-functional requirements, and release targets.
- The implemented application source is in `app.go`, `internal/`, and `frontend/src/`.
- No testimonials, customer evidence, or marketing claims are available beyond the documented idle RAM benchmark; future work must not fabricate them.

## Product Principles

1. Keep the core anime workflow in one portable application.
2. Prefer dependable local control and transparent configuration over mandatory external services.
3. Automate repetitive metadata and watch-progress tasks while keeping user control over playback and downloads.
4. Treat low operational overhead and cross-platform portability as product requirements.
5. Distinguish shipped behavior from planned capability and communicate deferred work honestly.
