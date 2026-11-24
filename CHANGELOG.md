# Changelog

## v0.4.5-alpha

- Pick several torrents from a search and add them all at once: **Select** turns on checkboxes, the footer shows how many torrents and their total size, and one **Download N** runs the batch.
- A batch inspects up to four torrents in the background and opens the existing file sheet once per torrent, so a slow or broken result can be skipped or retried without losing its place.
- The end of a batch reports how many downloads were added and how many failed, and stays on Search with the selection kept when nothing could be read.

## v0.4.4-alpha

- Advanced anime search on AniList: combine a title, genres and excluded genres, year, season, format, status, minimum score and sort in one sheet; the filters are remembered and results share Load more with seasonal browse.
- When AniList cannot be reached, Miru labels the data it shows from cache with its age above every tab, and a **Retry** button re-runs the request that failed.
- List changes now also clear cached season and advanced search pages, so saved results stop showing a title as not on your list for 15 minutes.
- Cached AniList responses older than seven days are deleted at startup, which bounds the offline fallback.

## v0.4.3-alpha

- Autoplay: playing an episode queues the next downloaded episodes of the same show, so a binge continues in one MPV window, with each episode's own resume position, player title, Discord presence and AniList sync.
- A translucent **Next** button appears in the bottom-right corner of the video during an episode's last minute when a following episode is queued; clicking it jumps ahead, and every other click keeps MPV's own meaning.
- The now-playing strip above the content area is a timeline only; pause, volume and full screen stay with MPV's own controls.
- A dev build keeps its own AniList login, so signing in while developing no longer reads or overwrites a release install's account.

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
