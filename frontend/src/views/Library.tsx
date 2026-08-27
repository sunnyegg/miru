import {useEffect, useMemo, useRef, useState} from 'react'
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
import {Alert} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'
import {Card} from '@/components/ui/card'
import {Input} from '@/components/ui/input'
import {Skeleton} from '@/components/ui/skeleton'

type Props = {
  notice: (msg: string, isError?: boolean) => void
  refreshKey: number
  playing: PlaybackEvent | null
}

export function LibraryView({notice, refreshKey, playing}: Props) {
  const [episodes, setEpisodes] = useState<EpisodeView[]>([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')
  const [busyId, setBusyId] = useState<number | null>(null)
  const [importing, setImporting] = useState(false)
  const [searching, setSearching] = useState(false)
  const [bindingAnimeId, setBindingAnimeId] = useState<number | null>(null)
  const [selectedKey, setSelectedKey] = useState<string | null>(null)
  const [selectedEpisodeId, setSelectedEpisodeId] = useState<number | null>(null)
  const [picker, setPicker] = useState<{episode: EpisodeView; candidates: AnimeView[]; query: string} | null>(null)
  const selectedEpisodeButtonRef = useRef<HTMLButtonElement>(null)

  const shows = useMemo(() => groupEpisodes(episodes), [episodes])
  const selectedShow = shows.find((show) => show.key === selectedKey) ?? null
  const selectedEpisode = selectedShow?.episodes.find((episode) => episode.id === selectedEpisodeId)
    ?? selectedShow?.episodes[0]
    ?? null

  async function reload() {
    setLoadError('')
    try {
      const rows = await ListEpisodes()
      setEpisodes(rows ?? [])
    } catch (err) {
      const message = errorMessage(err)
      setLoadError(message)
      if (episodes.length > 0) {
        notice(message, true)
      }
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

  useEffect(() => {
    selectedEpisodeButtonRef.current?.scrollIntoView({
      block: 'nearest',
      inline: 'center',
    })
  }, [selectedEpisodeId])

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
    const episodeId = picker.episode.id
    const query = picker.query.trim()
    if (!query) {
      return
    }
    setSearching(true)
    try {
      const found = await SearchAnime(query)
      setPicker((current) => {
        if (current?.episode.id !== episodeId || current.query.trim() !== query) {
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
    setBindingAnimeId(anilistId)
    try {
      await BindEpisode(episodeId, anilistId)
      setPicker((current) => current?.episode.id === episodeId ? null : current)
      await reload()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBindingAnimeId(null)
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

  const selectedEpisodeIsPlaying = Boolean(
    playing && selectedEpisode && playing.episodeId === selectedEpisode.id,
  )
  const progress = selectedEpisodeIsPlaying && Number.isFinite(playing?.percent)
    ? Math.min(100, Math.max(0, playing?.percent ?? 0))
    : 0

  const fileLabel = selectedEpisode
    ? selectedEpisode.filePath.split(/[\\/]/).pop() || selectedEpisode.displayTitle
    : ''

  return (
    <section className="flex h-full min-h-0 flex-col">
      <header className="flex shrink-0 flex-wrap items-center justify-between gap-3 px-5 pt-5 pb-3">
        <h2 className="text-2xl font-semibold tracking-tight">Library</h2>
        <Button type="button" variant="secondary" onClick={() => void onImport()} disabled={importing}>
          {importing ? 'Importing…' : 'Import file'}
        </Button>
      </header>

      {picker && (
        <Card className="mx-5 mb-3 border border-border" role="region" aria-labelledby="match-title">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <h3 id="match-title" className="text-base font-medium">Match AniList title</h3>
              <p className="mt-1 wrap-break-word text-sm text-muted-foreground">{picker.episode.displayTitle}</p>
            </div>
            <Button type="button" variant="ghost" onClick={() => setPicker(null)}>
              Skip
            </Button>
          </div>
          <div className="mt-3 flex flex-wrap gap-2">
            <label className="sr-only" htmlFor="anilist-search">Search AniList</label>
            <Input
              id="anilist-search"
              value={picker.query}
              onChange={(e) => setPicker({...picker, query: e.target.value})}
              className="min-w-0 flex-1 basis-56 border-border"
            />
            <Button
              type="button"
              variant="secondary"
              onClick={() => void onSearch()}
              disabled={searching || !picker.query.trim()}
            >
              {searching ? 'Searching…' : 'Search'}
            </Button>
          </div>
          <ul className="mt-3 flex flex-col gap-1">
            {picker.candidates.map((anime) => (
              <li key={anime.id}>
                <button
                  type="button"
                  onClick={() => void onBind(anime.id)}
                  disabled={bindingAnimeId !== null}
                  aria-busy={bindingAnimeId === anime.id}
                  className="flex min-h-11 w-full cursor-pointer items-center gap-3 bg-muted px-3 py-2 text-left hover:bg-secondary disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {anime.coverImage ? (
                    <img
                      src={anime.coverImage}
                      alt=""
                      width={40}
                      height={56}
                      loading="lazy"
                      decoding="async"
                      className="shrink-0 object-cover"
                      style={{width: 40, height: 56}}
                    />
                  ) : (
                    <span className="shrink-0 bg-muted" style={{width: 40, height: 56}} />
                  )}
                  <span className="min-w-0">
                    <span className="block truncate text-sm font-medium" title={anime.titleEnglish || anime.titleRomaji}>
                      {anime.titleEnglish || anime.titleRomaji}
                    </span>
                    <span className="block truncate text-xs text-muted-foreground" title={anime.titleRomaji}>
                      {anime.titleRomaji}
                    </span>
                  </span>
                </button>
              </li>
            ))}
            {picker.candidates.length === 0 && (
              <li className="text-sm text-muted-foreground">No matches. Search with a cleaner title.</li>
            )}
          </ul>
        </Card>
      )}

      <div className="min-h-0 flex-1 overflow-auto px-5 pb-4">
        {loading ? (
          <ul className="grid grid-cols-[repeat(auto-fill,minmax(min(8.5rem,100%),1fr))] gap-3" aria-busy="true" aria-label="Loading library">
            {Array.from({length: 8}, (_, index) => (
              <li key={index}>
                <Skeleton className="aspect-square w-full" />
              </li>
            ))}
          </ul>
        ) : loadError && shows.length === 0 ? (
          <Alert variant="destructive" className="flex min-h-48 flex-wrap items-end justify-between gap-4 p-4">
            <div className="min-w-0">
              <h3 className="font-medium text-foreground">Library could not be loaded</h3>
              <p className="mt-1 wrap-break-word text-sm text-destructive">{loadError}</p>
            </div>
            <Button
              type="button"
              variant="secondary"
              onClick={() => {
                setLoading(true)
                void reload()
              }}
            >
              Try again
            </Button>
          </Alert>
        ) : shows.length === 0 ? (
          <div className="flex h-full min-h-48 items-end">
            <p className="max-w-md text-sm text-muted-foreground">
              No local shows yet. Import a file, or finish a torrent download and it will land here.
            </p>
          </div>
        ) : (
          <ul className="grid grid-cols-[repeat(auto-fill,minmax(min(8.5rem,100%),1fr))] gap-3">
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
                        loading="lazy"
                        decoding="async"
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
        className="relative shrink-0 border-t border-border bg-bezel py-3"
        role="region"
        aria-label="Playback"
        style={{paddingInline: 16}}
      >
        <div
          className="pointer-events-none absolute inset-x-0 top-0 h-0.5 bg-muted"
          role={selectedEpisodeIsPlaying ? 'progressbar' : undefined}
          aria-hidden={selectedEpisodeIsPlaying ? undefined : true}
          aria-label={selectedEpisodeIsPlaying ? `Playback progress for ${selectedEpisode?.displayTitle}` : undefined}
          aria-valuemin={selectedEpisodeIsPlaying ? 0 : undefined}
          aria-valuemax={selectedEpisodeIsPlaying ? 100 : undefined}
          aria-valuenow={selectedEpisodeIsPlaying ? Math.round(progress) : undefined}
          aria-valuetext={selectedEpisodeIsPlaying ? `${Math.round(progress)}% played` : undefined}
        >
          <div
            className="osc-progress-fill h-full bg-accent transition-[width] duration-200"
            style={{width: `${progress}%`}}
          />
        </div>
        {selectedShow && selectedEpisode ? (
          <div key={selectedShow.key} className="osc-drop flex min-w-0 items-center" style={{gap: 12}}>
            <p className="min-w-0 max-w-[28%] shrink-0 truncate text-sm text-foreground" title={fileLabel}>
              {fileLabel}
            </p>
            <div className="min-w-0 flex-1 overflow-x-auto py-1" role="group" aria-label="Episodes">
              <div className="relative flex w-max min-w-full items-center justify-between px-1">
                <div className="pointer-events-none absolute inset-x-1 top-1/2 h-px bg-border" aria-hidden="true" />
                {selectedShow.episodes.map((episode) => {
                  const current = episode.id === selectedEpisode.id
                  const label = episode.episodeNumber > 0 ? `Episode ${episode.episodeNumber}` : episode.displayTitle
                  return (
                    <button
                      key={episode.id}
                      ref={current ? selectedEpisodeButtonRef : undefined}
                      type="button"
                      aria-label={label}
                      aria-pressed={current}
                      onClick={() => setSelectedEpisodeId(episode.id)}
                      className="relative flex h-11 min-w-6 shrink-0 cursor-pointer items-center justify-center"
                    >
                      <span
                        className={`block h-3 w-0.5 ${current ? 'h-4 bg-accent' : 'bg-muted-foreground'}`}
                      />
                    </button>
                  )
                })}
              </div>
            </div>
            <Button
              type="button"
              onClick={() => void onPlay(selectedEpisode.id)}
              disabled={busyId === selectedEpisode.id}
              style={{minWidth: 96, gap: 8, paddingInline: 20}}
            >
              <span className="shrink-0" style={{width: 16, height: 16}}>
                <IconPlay className="size-full" />
              </span>
              {busyId === selectedEpisode.id ? 'Starting…' : 'Play'}
            </Button>
          </div>
        ) : (
          <div className="osc-drop flex min-w-0 items-center" style={{gap: 12}}>
            <p className="min-w-0 flex-1 truncate text-sm text-muted-foreground">
              {loading ? 'Loading library…' : loadError ? 'Library unavailable' : 'No local shows yet'}
            </p>
            <Button type="button" disabled style={{minWidth: 96, gap: 8, paddingInline: 20}}>
              <span className="shrink-0" style={{width: 16, height: 16}}>
                <IconPlay className="size-full" />
              </span>
              Play
            </Button>
          </div>
        )}
      </div>
    </section>
  )
}
