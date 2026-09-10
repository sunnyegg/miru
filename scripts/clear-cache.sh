#!/usr/bin/env bash
# Clear Miru's cache before a dev run: the OS cache dir and the SQLite
# api_cache table (AniList/detail/airing responses). Settings, the AniList
# token, shaders, logs, and downloads are left untouched.
set -euo pipefail

case "$(uname -s)" in
	Linux)
		config_root="${XDG_CONFIG_HOME:-$HOME/.config}"
		cache_root="${XDG_CACHE_HOME:-$HOME/.cache}"
		;;
	Darwin)
		config_root="$HOME/Library/Application Support"
		cache_root="$HOME/Library/Caches"
		;;
	*)
		printf 'Error: unsupported OS: %s\n' "$(uname -s)" >&2
		exit 1
		;;
esac

cache_dir="$cache_root/miru"
db_file="$config_root/miru/app_data.db"

if [[ -d "$cache_dir" ]]; then
	rm -rf "$cache_dir"
	printf 'Cleared cache dir: %s\n' "$cache_dir"
else
	printf 'Cache dir not present, skipping: %s\n' "$cache_dir"
fi

if [[ ! -f "$db_file" ]]; then
	printf 'Database not present, skipping: %s\n' "$db_file"
	exit 0
fi

if ! command -v sqlite3 >/dev/null 2>&1; then
	printf 'Warning: sqlite3 not found; api_cache in %s was not cleared.\n' "$db_file" >&2
	exit 0
fi

if ! sqlite3 "$db_file" 'BEGIN IMMEDIATE; ROLLBACK;' >/dev/null 2>&1; then
	printf 'Error: database is locked, close Miru first: %s\n' "$db_file" >&2
	exit 1
fi

if ! sqlite3 "$db_file" "SELECT 1 FROM sqlite_master WHERE type='table' AND name='api_cache';" | grep -q 1; then
	printf 'api_cache table not found, skipping: %s\n' "$db_file"
	exit 0
fi

deleted=$(sqlite3 "$db_file" 'DELETE FROM api_cache; SELECT changes();')
printf 'Cleared %s api_cache rows: %s\n' "$deleted" "$db_file"
