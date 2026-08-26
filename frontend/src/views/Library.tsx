import {useEffect, useMemo, useState} from 'react'
import {
  BindEpisode,
  ImportLocalFile,
  ListEpisodes,
  PlayEpisode,
  SearchAnime,
} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import {groupEpisodes} from '../lib/groupEpisodes'
import type {AnimeView, EpisodeView, PlaybackEvent} from '../lib/types'
import {IconPlay} from '../components/Icons'

type Props = {
  notice: (msg: string, isError?: boolean) => void
  refreshKey: number
  playing: PlaybackEvent | null
}

export function LibraryView({notice, refreshKey, playing}: Props) {
  const [episodes, setEpisodes] = useState<EpisodeView[]>([])
  const [loading, setLoading] = useState(true)
  const [busyId, setBusyId] = useState<number | null>(null)
  const [importing, setImporting] = useState(false)
  const [selectedKey, setSelectedKey] = useState<string | null>(null)
  const [selectedEpisodeId, setSelectedEpisodeId] = useState<number | null>(null)
  const [picker, setPicker] = useState<{episode: EpisodeView; candidates: AnimeView[]; query: string} | null>(null)

  const shows = useMemo(() => groupEpisodes(episodes), [episodes])
  const selectedShow = shows.find((show) => show.key === selectedKey) ?? null
  const selectedEpisode = selectedShow?.episodes.find((episode) => episode.id === selectedEpisodeId)
    ?? selectedShow?.episodes[0]
    ?? null

  async function reload() {
    try {
      const rows = await ListEpisodes()
      setEpisodes(rows ?? [])
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void reload()
  }, [refreshKey])

  useEffect(() => {
    if (shows.length === 0) {
      setSelectedKey(null)
      setSelectedEpisodeId(null)
      return
    }

    if (playing) {
      const owner = shows.find((show) => show.episodes.some((episode) => episode.id === playing.episodeId))
      if (owner) {
        setSelectedKey(owner.key)
        setSelectedEpisodeId(playing.episodeId)
        return
      }
    }

    const stillThere = selectedKey && shows.some((show) => show.key === selectedKey)
    if (!stillThere) {
      setSelectedKey(shows[0].key)
      setSelectedEpisodeId(shows[0].episodes[0]?.id ?? null)
    }
  }, [shows, playing, selectedKey])

  async function onImport() {
    setImporting(true)
    try {
      const result = await ImportLocalFile()
      if (!result?.episode?.id) {
        return
      }
      await reload()
      if (!result.autoBound) {
        setPicker({
          episode: result.episode,
          candidates: result.candidates ?? [],
          query: result.episode.displayTitle,
        })
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
    try {
      const found = await SearchAnime(picker.query)
      setPicker({...picker, candidates: found ?? []})
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  async function onBind(anilistId: number) {
    if (!picker) {
      return
    }
    try {
      await BindEpisode(picker.episode.id, anilistId)
      setPicker(null)
      await reload()
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  function selectShow(key: string) {
    const show = shows.find((item) => item.key === key)
    if (!show) {
      return
    }
    setSelectedKey(key)
    const playingInShow = playing
      ? show.episodes.find((episode) => episode.id === playing.episodeId)
      : null
    setSelectedEpisodeId(playingInShow?.id ?? show.episodes[0]?.id ?? null)
  }

  const progress = playing && selectedEpisode && playing.episodeId === selectedEpisode.id
    ? playing.percent
    : 0

  const fileLabel = selectedEpisode
    ? selectedEpisode.filePath.split(/[\\/]/).pop() || selectedEpisode.displayTitle
    : ''

  return (
    <section className="flex h-full min-h-0 flex-col">
      <header className="flex shrink-0 items-center justify-between gap-3 px-5 pt-5 pb-3">
        <h2 className="text-2xl font-semibold tracking-tight">Library</h2>
        <button
          type="button"
          onClick={() => void onImport()}
          disabled={importing}
          className="inline-flex min-h-11 cursor-pointer items-center bg-secondary px-4 text-sm text-on-secondary transition-colors duration-200 hover:bg-muted disabled:cursor-not-allowed disabled:opacity-50"
        >
          {importing ? 'Importing…' : 'Import file'}
        </button>
      </header>

      {picker && (
        <div className="mx-5 mb-3 border border-border bg-card p-4" role="region" aria-labelledby="match-title">
          <div className="flex items-start justify-between gap-3">
            <div>
              <h3 id="match-title" className="text-base font-medium">Match AniList title</h3>
              <p className="mt-1 text-sm text-muted-foreground">{picker.episode.displayTitle}</p>
            </div>
            <button
              type="button"
              className="min-h-11 cursor-pointer px-3 text-sm text-muted-foreground hover:text-foreground"
              onClick={() => setPicker(null)}
            >
              Skip
            </button>
          </div>
          <div className="mt-3 flex flex-wrap gap-2">
            <label className="sr-only" htmlFor="anilist-search">Search AniList</label>
            <input
              id="anilist-search"
              value={picker.query}
              onChange={(e) => setPicker({...picker, query: e.target.value})}
              className="min-h-11 min-w-56 flex-1 border border-border bg-muted px-3 text-sm text-foreground"
            />
            <button
              type="button"
              onClick={() => void onSearch()}
              className="min-h-11 cursor-pointer bg-secondary px-4 text-sm text-on-secondary"
            >
              Search
            </button>
          </div>
          <ul className="mt-3 flex flex-col gap-1">
            {picker.candidates.map((anime) => (
              <li key={anime.id}>
                <button
                  type="button"
                  onClick={() => void onBind(anime.id)}
                  className="flex min-h-11 w-full cursor-pointer items-center gap-3 bg-muted px-3 py-2 text-left hover:bg-secondary"
                >
                  {anime.coverImage ? (
                    <img src={anime.coverImage} alt="" width={40} height={56} className="h-14 w-10 object-cover" />
                  ) : (
                    <span className="h-14 w-10 bg-primary" />
                  )}
                  <span>
                    <span className="block text-sm font-medium">{anime.titleEnglish || anime.titleRomaji}</span>
                    <span className="block text-xs text-muted-foreground">{anime.titleRomaji}</span>
                  </span>
                </button>
              </li>
            ))}
            {picker.candidates.length === 0 && (
              <li className="text-sm text-muted-foreground">No matches. Search with a cleaner title.</li>
            )}
          </ul>
        </div>
      )}

      <div className="min-h-0 flex-1 overflow-auto px-5 pb-4">
        {loading ? (
          <ul className="grid grid-cols-[repeat(auto-fill,minmax(8.5rem,1fr))] gap-3" aria-busy="true" aria-label="Loading library">
            {Array.from({length: 8}, (_, index) => (
              <li key={index} className="aspect-square bg-muted" />
            ))}
          </ul>
        ) : shows.length === 0 ? (
          <div className="flex h-full min-h-48 items-end">
            <p className="max-w-md text-sm text-muted-foreground">
              No local shows yet. Import a file, or finish a torrent download and it will land here.
            </p>
          </div>
        ) : (
          <ul className="grid grid-cols-[repeat(auto-fill,minmax(8.5rem,1fr))] gap-3">
            {shows.map((show) => {
              const active = show.key === selectedKey
              return (
                <li key={show.key}>
                  <button
                    type="button"
                    onClick={() => selectShow(show.key)}
                    aria-pressed={active}
                    className={`group flex w-full cursor-pointer flex-col text-left ${
                      active ? 'outline-2 outline-offset-2 outline-accent' : ''
                    }`}
                  >
                    {show.coverImage ? (
                      <img
                        src={show.coverImage}
                        alt=""
                        className="aspect-square w-full bg-muted object-cover"
                      />
                    ) : (
                      <span className="flex aspect-square w-full items-end bg-muted p-2 text-xs text-muted-foreground">
                        {show.title}
                      </span>
                    )}
                    <span className="mt-2 truncate text-sm font-medium">{show.title}</span>
                    <span className="truncate text-xs text-muted-foreground">
                      {show.episodes.length} episode{show.episodes.length === 1 ? '' : 's'}
                      {show.unlinkedCount > 0 ? ' · not linked' : ''}
                    </span>
                  </button>
                </li>
              )
            })}
          </ul>
        )}
      </div>

      <div
        className="osc-drop relative shrink-0 border-t border-border bg-bezel px-4 py-3"
        role="region"
        aria-label="Playback"
      >
        <div className="pointer-events-none absolute inset-x-0 top-0 h-0.5 bg-muted" aria-hidden="true">
          <div
            className="h-full bg-accent transition-[width] duration-200"
            style={{width: `${Math.min(100, Math.max(0, progress))}%`}}
          />
        </div>
        {selectedShow && selectedEpisode ? (
          <div className="flex min-w-0 items-center gap-3">
            <p className="min-w-0 max-w-[28%] shrink-0 truncate text-sm text-foreground" title={fileLabel}>
              {fileLabel}
            </p>
            <div className="relative min-w-0 flex-1" role="tablist" aria-label="Episodes">
              <div className="pointer-events-none absolute inset-x-0 top-1/2 h-px bg-border" aria-hidden="true" />
              <div className="flex items-center justify-between">
                {selectedShow.episodes.map((episode) => {
                  const current = episode.id === selectedEpisode.id
                  const label = episode.episodeNumber > 0 ? `Episode ${episode.episodeNumber}` : episode.displayTitle
                  return (
                    <button
                      key={episode.id}
                      type="button"
                      role="tab"
                      aria-label={label}
                      aria-selected={current}
                      onClick={() => setSelectedEpisodeId(episode.id)}
                      className="relative flex min-h-11 min-w-3 cursor-pointer items-center justify-center"
                    >
                      <span
                        className={`block h-3 w-0.5 ${current ? 'h-4 bg-accent' : 'bg-muted-foreground'}`}
                      />
                    </button>
                  )
                })}
              </div>
            </div>
            <button
              type="button"
              onClick={() => void onPlay(selectedEpisode.id)}
              disabled={busyId === selectedEpisode.id}
              className="inline-flex min-h-11 min-w-24 shrink-0 cursor-pointer items-center justify-center gap-2 bg-accent px-5 text-sm font-medium text-on-accent disabled:cursor-not-allowed disabled:opacity-50"
            >
              <IconPlay className="h-4 w-4" />
              {busyId === selectedEpisode.id ? 'Starting…' : 'Play'}
            </button>
          </div>
        ) : (
          <div className="flex min-w-0 items-center gap-3">
            <p className="min-w-0 flex-1 truncate text-sm text-muted-foreground">
              {loading ? 'Loading library…' : 'No local shows yet'}
            </p>
            <button
              type="button"
              disabled
              className="inline-flex min-h-11 min-w-24 shrink-0 items-center justify-center gap-2 bg-accent px-5 text-sm font-medium text-on-accent disabled:cursor-not-allowed disabled:opacity-50"
            >
              <IconPlay className="h-4 w-4" />
              Play
            </button>
          </div>
        )}
      </div>
    </section>
  )
}
