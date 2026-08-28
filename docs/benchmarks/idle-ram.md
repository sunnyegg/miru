# Idle RAM benchmark (Linux)

Miru idle memory is measured as the **sum of resident set size (RSS)** for the main process and all descendant processes, sampled from `/proc/<pid>/status` (`VmRSS`).

## Run

```bash
make build          # requires .env
make bench-idle-ram
```

Optional environment variables:

| Variable | Default | Meaning |
| --- | --- | --- |
| `MIRU_BENCH_BINARY` | `build/bin/miru-linux-amd64` | Binary to launch |
| `MIRU_BENCH_WARMUP_SEC` | `20` | Seconds after startup before sampling |
| `MIRU_BENCH_SAMPLE_COUNT` | `6` | Number of RSS samples |
| `MIRU_BENCH_SAMPLE_INTERVAL_SEC` | `5` | Seconds between samples |
| `MIRU_BENCH_OUTPUT_JSON` | _(none)_ | Write machine-readable results to this path |

The script uses an isolated temp `XDG_CONFIG_HOME` / `XDG_CACHE_HOME` (empty library, no AniList token) and sets `MIRU_BENCH_IDLE=1` so the RSS feed poller does not run during the measurement window.

If `DISPLAY` is unset, the script requires `xvfb-run`.

## Latest result (2026-08-28)

| Field | Value |
| --- | --- |
| Platform | Linux amd64 (CachyOS, production `wails build`) |
| Average idle RSS (process tree) | **683.5 MB** |
| Min / max | 683.5 / 683.5 MB |
| PRD target | < 100 MB |
| Verdict | **Above target** — WebKitGTK + embedded React UI dominate idle footprint |

Raw data: [`idle-ram-linux.json`](idle-ram-linux.json)

Samples were stable (±28 KB) over 30 seconds of polling after a 20-second warmup, which indicates a settled idle state rather than post-startup allocation noise.

## Interpretation

The PRD target (< 100 MB idle) reflects an aspiration for a lightweight native client. Miru’s stack (Wails + WebKitGTK webview + Go backend + embedded frontend assets) is inherently heavier on Linux than a pure native toolkit would be. The Go backend alone is small; most RSS comes from the WebKit renderer process tree.

This benchmark establishes a **measured baseline** so future optimizations (asset size, webview lifecycle, deferred module init) can be tracked against real numbers instead of estimates.

## Reproducing on other machines

1. Build a production binary (`make build`).
2. Close other Miru instances.
3. Run `make bench-idle-ram`.
4. Commit an updated `idle-ram-linux.json` (or add platform-specific JSON) if you re-run on a new environment worth recording.

Windows and macOS are not automated yet; the same methodology can be adapted with Task Manager / Activity Monitor process-tree sums.
