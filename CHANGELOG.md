# Changelog

## v0.4.2-alpha

- Control playback from the app: play/pause, seek, volume, full screen and stop in a now-playing strip.
- The MPV window now behaves like a normal window: real title, free resize, and controls that keep their size on small windows.
- On Wayland, the player gets a native titlebar by default, with a new Playback setting to use native Wayland instead.
- Local dev builds use their own config, cache, log and single-instance id (`miru-dev`), so they no longer collide with a release install.

## v0.4.1-alpha

- Group Library shelves (Available, Upcoming, On disk) by AniList airing season.
- Browse AniList by season in Search → Anime (season + year, popularity order, load more), with the same 15-minute cache as title search.
- Prefer denser poster mosaics when filling Profile share banners.
- Add torrent search title-language / prefix prefs, and keep Downloads cards collapsed by default.

## v0.3.5-alpha

- Cache AniList search and library progress per account, and invalidate user data when accounts or list entries change.
- Reuse cached anime details when linking library episodes and cache the AniList viewer ID for the active client.
- Show a clear message when an API returns HTTP 429.
