#!/usr/bin/env bash
# Measure Miru idle RAM (process tree RSS) on Linux.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY="${MIRU_BENCH_BINARY:-$ROOT_DIR/build/bin/miru-linux-amd64}"
WARMUP_SEC="${MIRU_BENCH_WARMUP_SEC:-20}"
SAMPLE_COUNT="${MIRU_BENCH_SAMPLE_COUNT:-6}"
SAMPLE_INTERVAL_SEC="${MIRU_BENCH_SAMPLE_INTERVAL_SEC:-5}"
OUTPUT_JSON="${MIRU_BENCH_OUTPUT_JSON:-}"

if [[ ! -x "$BINARY" ]]; then
	printf 'Error: binary not found or not executable: %s\n' "$BINARY" >&2
	printf 'Build first: make build\n' >&2
	exit 1
fi

if [[ "$(uname -s)" != "Linux" ]]; then
	printf 'Error: idle RAM benchmark currently supports Linux only.\n' >&2
	exit 1
fi

launch_wrapper=(env)
if [[ -z "${DISPLAY:-}" ]]; then
	if command -v xvfb-run >/dev/null 2>&1; then
		launch_wrapper=(xvfb-run -a env)
	else
		printf 'Error: DISPLAY is unset and xvfb-run is not installed.\n' >&2
		exit 1
	fi
fi

bench_dir="$(mktemp -d "${TMPDIR:-/tmp}/miru-bench.XXXXXX")"
cleanup() {
	if [[ -n "${app_pid:-}" ]] && kill -0 "$app_pid" 2>/dev/null; then
		kill "$app_pid" 2>/dev/null || true
		wait "$app_pid" 2>/dev/null || true
	fi
	rm -rf "$bench_dir"
}
trap cleanup EXIT

export XDG_CONFIG_HOME="$bench_dir/config"
export XDG_CACHE_HOME="$bench_dir/cache"
export XDG_DATA_HOME="$bench_dir/data"
mkdir -p "$XDG_CONFIG_HOME" "$XDG_CACHE_HOME" "$XDG_DATA_HOME"

# Avoid unrelated background network work during the idle window.
export MIRU_BENCH_IDLE=1

tree_rss_kb() {
	local root_pid="$1"
	local total=0
	local queue=("$root_pid")
	local seen=" "

	while ((${#queue[@]} > 0)); do
		local pid="${queue[0]}"
		queue=("${queue[@]:1}")

		if [[ "$seen" == *" $pid "* ]]; then
			continue
		fi
		seen+=" $pid "

		if [[ ! -r "/proc/$pid/status" ]]; then
			continue
		fi

		local rss
		rss="$(awk '/^VmRSS:/ {print $2}' "/proc/$pid/status")"
		total=$((total + rss))

		local child
		while IFS= read -r child; do
			[[ -n "$child" ]] && queue+=("$child")
		done < <(pgrep -P "$pid" 2>/dev/null || true)
	done

	printf '%s' "$total"
}

find_miru_root_pid() {
	local launcher_pid="$1"
	local candidate
	for candidate in "$launcher_pid" $(pgrep -P "$launcher_pid" 2>/dev/null || true); do
		if [[ -r "/proc/$candidate/cmdline" ]] && tr '\0' ' ' < "/proc/$candidate/cmdline" | grep -q 'miru'; then
			printf '%s' "$candidate"
			return 0
		fi
	done
	printf '%s' "$launcher_pid"
}

printf 'Binary: %s\n' "$BINARY"
printf 'Warmup: %ss, samples: %s every %ss\n' "$WARMUP_SEC" "$SAMPLE_COUNT" "$SAMPLE_INTERVAL_SEC"
printf 'Isolated config: %s\n' "$XDG_CONFIG_HOME"

"${launch_wrapper[@]}" "$BINARY" >/dev/null 2>&1 &
launcher_pid=$!

for _ in $(seq 1 60); do
	if ! kill -0 "$launcher_pid" 2>/dev/null; then
		printf 'Error: Miru exited before benchmark warmup.\n' >&2
		exit 1
	fi
	app_pid="$(find_miru_root_pid "$launcher_pid")"
	if [[ -n "$app_pid" ]] && [[ "$(tree_rss_kb "$app_pid")" -gt 1024 ]]; then
		break
	fi
	sleep 1
done

printf 'Root PID: %s\n' "$app_pid"
printf 'Warming up %ss...\n' "$WARMUP_SEC"
sleep "$WARMUP_SEC"

samples_kb=()
for sample_index in $(seq 1 "$SAMPLE_COUNT"); do
	sample_kb="$(tree_rss_kb "$app_pid")"
	samples_kb+=("$sample_kb")
	printf 'Sample %s/%s: %s MB (%s KB RSS tree)\n' \
		"$sample_index" "$SAMPLE_COUNT" \
		"$(awk "BEGIN {printf \"%.1f\", $sample_kb / 1024}")" \
		"$sample_kb"
	if [[ "$sample_index" -lt "$SAMPLE_COUNT" ]]; then
		sleep "$SAMPLE_INTERVAL_SEC"
	fi
done

min_kb="${samples_kb[0]}"
max_kb="${samples_kb[0]}"
sum_kb=0
for sample_kb in "${samples_kb[@]}"; do
	sum_kb=$((sum_kb + sample_kb))
	if ((sample_kb < min_kb)); then
		min_kb=$sample_kb
	fi
	if ((sample_kb > max_kb)); then
		max_kb=$sample_kb
	fi
done
avg_kb=$((sum_kb / SAMPLE_COUNT))
target_kb=102400

printf '\n--- Summary ---\n'
printf 'Average idle RSS (process tree): %.1f MB (%s KB)\n' "$(awk "BEGIN {printf \"%.1f\", $avg_kb / 1024}")" "$avg_kb"
printf 'Min: %.1f MB  Max: %.1f MB\n' \
	"$(awk "BEGIN {printf \"%.1f\", $min_kb / 1024}")" \
	"$(awk "BEGIN {printf \"%.1f\", $max_kb / 1024}")"
printf 'PRD target: < %.1f MB (%s KB)\n' "$(awk "BEGIN {printf \"%.1f\", $target_kb / 1024}")" "$target_kb"
if ((avg_kb < target_kb)); then
	printf 'Result: PASS (below target)\n'
else
	printf 'Result: ABOVE TARGET (WebKitGTK overhead dominates on Linux)\n'
fi

if [[ -n "$OUTPUT_JSON" ]]; then
	mkdir -p "$(dirname "$OUTPUT_JSON")"
	{
		printf '{'
		printf '"platform":"linux",'
		printf '"binary":"%s",' "$BINARY"
		printf '"warmupSec":%s,' "$WARMUP_SEC"
		printf '"sampleCount":%s,' "$SAMPLE_COUNT"
		printf '"sampleIntervalSec":%s,' "$SAMPLE_INTERVAL_SEC"
		printf '"avgRssKb":%s,' "$avg_kb"
		printf '"minRssKb":%s,' "$min_kb"
		printf '"maxRssKb":%s,' "$max_kb"
		printf '"targetKb":%s,' "$target_kb"
		printf '"samplesKb":['
		for index in "${!samples_kb[@]}"; do
			if ((index > 0)); then
				printf ','
			fi
			printf '%s' "${samples_kb[$index]}"
		done
		printf ']}'
	} >"$OUTPUT_JSON"
	printf 'Wrote %s\n' "$OUTPUT_JSON"
fi
