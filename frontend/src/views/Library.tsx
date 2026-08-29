import {useEffect, useMemo, useRef, useState} from 'react'
import {
  BindEpisode,
  ImportLocalFile,
  ListStreamingEpisodeThumbnails,
  PlayEpisode,
  SearchAnime,
  SetAnimeListStatus,
  UnbindEpisode,
} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import {groupEpisodes, visibleLibraryEpisodes} from '../lib/groupEpisodes'
import type {AnimeView, EpisodeView} from '../lib/types'
import {useLibraryStore} from '../stores/libraryStore'
import {usePlaybackStore} from '../stores/playbackStore'
import {pickContinueHeroKey} from '../lib/libraryWatching'
import {LibraryContinueHero} from '../components/LibraryContinueHero'
import {
  LibraryMatchSheet,
  type LibraryMatchPicker,
} from '../components/LibraryMatchSheet'
import {LibraryEpisodeList} from '../components/LibraryEpisodeList'
import {LibraryShowDetailHero} from '../components/LibraryShowDetailHero'
import {LibraryUnlistedSection} from '../components/LibraryUnlistedSection'
import {LibraryWatchingSection} from '../components/LibraryWatchingSection'
import {IconBack} from '../components/Icons'
import {Button} from '@/components/ui/button'

type Props = {
  notice: (msg: string, isError?: boolean) => void
  onFindTorrent: (query: string) => void
  onReady?: () => void
}

export function LibraryView({notice, onFindTorrent, onReady}: Props) {
  const episodes = useLibraryStore((state) => state.episodes)
  const watchingEntries = useLibraryStore((state) => state.watchingEntries)
  const loading = useLibraryStore((state) => state.loading)
  const watchingLoading = useLibraryStore((state) => state.watchingLoading)
  const loadError = useLibraryStore((state) => state.loadError)
  const selectedKey = useLibraryStore((state) => state.selectedKey)
  const setSelectedKey = useLibraryStore((state) => state.setSelectedKey)
  const reload = useLibraryStore((state) => state.reload)
  const reloadWatching = useLibraryStore((state) => state.reloadWatching)
  const playing = usePlaybackStore((state) => state.playing)
  const lastPlayback = usePlaybackStore((state) => state.lastPlayback)

  const [busyId, setBusyId] = useState<number | null>(null)
  const [importing, setImporting] = useState(false)
  const [searching, setSearching] = useState(false)
  const [bindingAnimeId, setBindingAnimeId] = useState<number | null>(null)
  const [addingToWatching, setAddingToWatching] = useState(false)
  const [unmatching, setUnmatching] = useState(false)
  const [unmatchingEpisodeId, setUnmatchingEpisodeId] = useState<number | null>(
    null,
  )
  const [picker, setPicker] = useState<LibraryMatchPicker | null>(null)
  const [episodeThumbnails, setEpisodeThumbnails] = useState<
    Record<number, string>
  >({})
  const skippedMatchIds = useRef(new Set<number>())
  const bindingEpisodeId = useRef<number | null>(null)
  const readySignaled = useRef(false)

  const libraryEpisodes = useMemo(
    () => visibleLibraryEpisodes(episodes),
    [episodes],
  )
  const shows = useMemo(() => groupEpisodes(libraryEpisodes), [libraryEpisodes])
  const watchingKeys = useMemo(
    () => new Set(watchingEntries.map((entry) => `anilist:${entry.mediaId}`)),
    [watchingEntries],
  )
  const gridShows = useMemo(
    () => shows.filter((show) => !watchingKeys.has(show.key)),
    [shows, watchingKeys],
  )
  const selectedShow = shows.find((show) => show.key === selectedKey) ?? null
  const selectedShowIsUnlisted = Boolean(
    selectedShow && !watchingKeys.has(selectedShow.key),
  )
  const playingShowKey = useMemo(() => {
    if (!playing) {
      return null
    }
    const owner = shows.find((show) =>
      show.episodes.some((episode) => episode.id === playing.episodeId),
    )
    return owner?.key ?? null
  }, [shows, playing])

  const continueHeroKey = useMemo(
    () =>
      pickContinueHeroKey(
        watchingEntries,
        shows,
        playingShowKey,
        lastPlayback?.episodeId ?? null,
        episodes,
      ),
    [watchingEntries, shows, playingShowKey, lastPlayback?.episodeId, episodes],
  )

  const selectedAnilistId = useMemo(() => {
    if (!selectedShow?.bound) {
      return 0
    }
    const match = selectedShow.key.match(/^anilist:(\d+)$/)
    if (!match) {
      return 0
    }
    return Number(match[1])
  }, [selectedShow])

  const selectedWatchingEntry = useMemo(() => {
    if (selectedAnilistId <= 0) {
      return null
    }
    return (
      watchingEntries.find((entry) => entry.mediaId === selectedAnilistId) ??
      null
    )
  }, [watchingEntries, selectedAnilistId])

  useEffect(() => {
    void reload(notice)
    void reloadWatching()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (!loading && !readySignaled.current) {
      readySignaled.current = true
      onReady?.()
    }
  }, [loading, onReady])

  useEffect(() => {
    if (shows.length === 0) {
      setSelectedKey(null)
      return
    }
    if (selectedKey && !shows.some((show) => show.key === selectedKey)) {
      setSelectedKey(null)
    }
  }, [shows, selectedKey, setSelectedKey])

  useEffect(() => {
    if (selectedAnilistId <= 0) {
      setEpisodeThumbnails({})
      return
    }

    let cancelled = false
    void (async () => {
      try {
        const rows = await ListStreamingEpisodeThumbnails(selectedAnilistId)
        if (cancelled) {
          return
        }
        const mapped: Record<number, string> = {}
        for (const row of rows ?? []) {
          if (row.thumbnail) {
            mapped[row.episodeNumber] = row.thumbnail
          }
        }
        setEpisodeThumbnails(mapped)
      } catch {
        if (!cancelled) {
          setEpisodeThumbnails({})
        }
      }
    })()

    return () => {
      cancelled = true
    }
  }, [selectedAnilistId])

  async function openMatcher(
    episode: EpisodeView,
    candidates: AnimeView[] = [],
  ) {
    const query = (episode.displayTitle || episode.animeTitle)
      .replace(/\s+—\s+Episode\s+\d+\s*$/i, '')
      .trim()
    setPicker({episode, candidates, query})
    if (!query || candidates.length > 0) {
      return
    }
    const episodeId = episode.id
    setSearching(true)
    try {
      const found = await SearchAnime(query)
      setPicker((current) => {
        if (current?.episode.id !== episodeId) {
          return current
        }
        return {...current, candidates: found ?? []}
      })
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setSearching(false)
    }
  }

  useEffect(() => {
    if (loading || picker || importing || bindingEpisodeId.current !== null) {
      return
    }
    const unbound = libraryEpisodes.find((episode) => {
      return !episode.bound && !skippedMatchIds.current.has(episode.id)
    })
    if (!unbound) {
      return
    }
    void openMatcher(unbound)
    // openMatcher is a local handler; libraryEpisodes/loading/picker/importing are the real triggers.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loading, picker, importing, libraryEpisodes])

  async function onImport() {
    setImporting(true)
    try {
      const result = await ImportLocalFile()
      if (!result?.episode?.id) {
        return
      }
      await reload(notice)
      if (!result.autoBound) {
        await openMatcher(result.episode, result.candidates ?? [])
      }
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setImporting(false)
    }
  }

  async function onPlay(id: number) {
    setBusyId(id)
    try {
      await PlayEpisode(id)
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBusyId(null)
    }
  }

  async function onSearch() {
    if (!picker) {
      return
    }
    const episodeId = picker.episode.id
    const query = picker.query.trim()
    if (!query) {
      return
    }
    setSearching(true)
    try {
      const found = await SearchAnime(query)
      setPicker((current) => {
        if (
          current?.episode.id !== episodeId ||
          current.query.trim() !== query
        ) {
          return current
        }
        return {...current, candidates: found ?? []}
      })
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setSearching(false)
    }
  }

  async function onBind(anilistId: number) {
    if (!picker || bindingAnimeId !== null) {
      return
    }
    const episodeId = picker.episode.id
    bindingEpisodeId.current = episodeId
    setBindingAnimeId(anilistId)
    try {
      await BindEpisode(episodeId, anilistId)
      skippedMatchIds.current.add(episodeId)
      await reload(notice)
      setPicker((current) =>
        current?.episode.id === episodeId ? null : current,
      )
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      bindingEpisodeId.current = null
      setBindingAnimeId(null)
    }
  }

  function selectShow(key: string) {
    const show = shows.find((item) => item.key === key)
    if (!show) {
      return
    }
    setSelectedKey(key)
    if (!show.bound) {
      const episode = show.episodes[0]
      if (episode) {
        skippedMatchIds.current.delete(episode.id)
        void openMatcher(episode)
      }
    }
  }

  function openWatchingShow(localShowKey: string) {
    selectShow(localShowKey)
  }

  async function addSelectedToWatching() {
    if (!selectedShow || addingToWatching) {
      return
    }
    const match = selectedShow.key.match(/^anilist:(\d+)$/)
    if (!match) {
      return
    }
    const mediaId = Number(match[1])
    setAddingToWatching(true)
    try {
      await SetAnimeListStatus(mediaId, 'CURRENT', selectedShow.totalEpisodes)
      notice('Added to Watching')
      await reloadWatching()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setAddingToWatching(false)
    }
  }

  function openMatcherForSelectedShow() {
    if (!selectedShow) {
      return
    }
    const episode = selectedShow.episodes[0]
    if (!episode) {
      return
    }
    skippedMatchIds.current.delete(episode.id)
    void openMatcher(episode)
  }

  async function unmatchEpisode(episodeId: number) {
    if (unmatching || unmatchingEpisodeId !== null) {
      return
    }

    bindingEpisodeId.current = episodeId
    setUnmatchingEpisodeId(episodeId)
    try {
      await UnbindEpisode(episodeId)
      skippedMatchIds.current.delete(episodeId)
      await reload(notice)
      const updatedShows = groupEpisodes(
        visibleLibraryEpisodes(useLibraryStore.getState().episodes),
      )
      const updatedShow = updatedShows.find((show) => {
        return show.episodes.some((episode) => episode.id === episodeId)
      })
      if (updatedShow) {
        setSelectedKey(updatedShow.key)
      } else {
        setSelectedKey(null)
      }
      notice('AniList match removed')
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      bindingEpisodeId.current = null
      setUnmatchingEpisodeId(null)
    }
  }

  async function unmatchSelectedShow() {
    if (
      !selectedShow ||
      !selectedShow.bound ||
      unmatching ||
      unmatchingEpisodeId !== null
    ) {
      return
    }
    const boundEpisodes = selectedShow.episodes.filter(
      (episode) => episode.bound,
    )
    if (boundEpisodes.length === 0) {
      return
    }

    const firstEpisodeId = boundEpisodes[0].id
    bindingEpisodeId.current = firstEpisodeId
    setUnmatching(true)
    try {
      for (const episode of boundEpisodes) {
        await UnbindEpisode(episode.id)
        skippedMatchIds.current.delete(episode.id)
      }
      await reload(notice)
      const updatedShows = groupEpisodes(
        visibleLibraryEpisodes(useLibraryStore.getState().episodes),
      )
      const updatedShow = updatedShows.find((show) => {
        return show.episodes.some((episode) => episode.id === firstEpisodeId)
      })
      if (updatedShow) {
        setSelectedKey(updatedShow.key)
      } else {
        setSelectedKey(null)
      }
      notice('AniList match removed')
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      bindingEpisodeId.current = null
      setUnmatching(false)
    }
  }

  function backToGrid() {
    setSelectedKey(null)
  }

  function skipMatch() {
    if (!picker) {
      return
    }
    skippedMatchIds.current.add(picker.episode.id)
    setPicker(null)
  }

  function retryLoad() {
    useLibraryStore.setState({loading: true})
    void reload(notice)
  }

  return (
    <section className="flex h-full min-h-0 flex-col">
      <header className="flex shrink-0 flex-wrap items-end justify-between gap-3 px-5 pt-5 pb-3">
        {selectedShow ? (
          <Button
            type="button"
            variant="ghost"
            onClick={backToGrid}
            style={{gap: 8, paddingInline: 12}}
          >
            <span className="shrink-0" style={{width: 16, height: 16}}>
              <IconBack className="size-full" />
            </span>
            Back
          </Button>
        ) : (
          <div className="min-w-0">
            <h2 className="text-2xl font-semibold tracking-tight">Library</h2>
            <p className="mt-0.5 text-sm text-muted-foreground">
              {watchingLoading && loading
                ? 'Loading…'
                : `${watchingEntries.length} watching · ${gridShows.length} local`}
            </p>
          </div>
        )}
        <Button
          type="button"
          variant="secondary"
          onClick={() => void onImport()}
          disabled={importing}
        >
          {importing ? 'Importing…' : 'Import file'}
        </Button>
      </header>

      {picker && (
        <LibraryMatchSheet
          picker={picker}
          searching={searching}
          bindingAnimeId={bindingAnimeId}
          onQueryChange={(query) => setPicker({...picker, query})}
          onSearch={() => void onSearch()}
          onBind={(anilistId) => void onBind(anilistId)}
          onSkip={skipMatch}
        />
      )}

      <div className="min-h-0 flex-1 overflow-auto px-5 pt-2 pb-4">
        {selectedShow ? (
          <>
            <LibraryShowDetailHero
              show={selectedShow}
              bannerImage={selectedWatchingEntry?.bannerImage ?? ''}
              showAddToWatching={selectedShowIsUnlisted}
              saving={addingToWatching}
              unmatching={unmatching}
              onAddToWatching={() => void addSelectedToWatching()}
              onMatchAnilist={openMatcherForSelectedShow}
              onUnmatchAnilist={() => void unmatchSelectedShow()}
            />
            <LibraryEpisodeList
              show={selectedShow}
              playing={playing}
              lastPlayback={lastPlayback}
              busyId={busyId}
              unmatchingEpisodeId={unmatchingEpisodeId}
              episodeThumbnails={episodeThumbnails}
              onPlay={(episodeId) => void onPlay(episodeId)}
              onUnmatch={
                selectedShow.bound
                  ? (episodeId) => void unmatchEpisode(episodeId)
                  : undefined
              }
              onFindTorrent={onFindTorrent}
            />
          </>
        ) : (
          <div className="flex flex-col">
            <LibraryContinueHero
              entries={watchingEntries}
              localShows={shows}
              libraryEpisodes={episodes}
              playing={playing}
              lastPlayback={lastPlayback}
              playingShowKey={playingShowKey}
              onOpenShow={openWatchingShow}
              onFindTorrent={onFindTorrent}
            />
            <LibraryWatchingSection
              entries={watchingEntries}
              localShows={shows}
              loading={watchingLoading}
              highlightedKey={playingShowKey}
              excludeHeroKey={continueHeroKey}
              onOpenShow={openWatchingShow}
              onFindTorrent={onFindTorrent}
            />
            <LibraryUnlistedSection
              loading={loading}
              loadError={loadError}
              shows={gridShows}
              highlightedKey={playingShowKey}
              onSelectShow={selectShow}
              onRetry={retryLoad}
              suppressEmptyState={watchingEntries.length > 0}
            />
          </div>
        )}
      </div>
    </section>
  )
}
