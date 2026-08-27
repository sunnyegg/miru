import {useEffect, useMemo, useRef, useState} from 'react'
import {
  BindEpisode,
  ImportLocalFile,
  ListEpisodes,
  PlayEpisode,
  SearchAnime,
} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import {groupEpisodes, visibleLibraryEpisodes} from '../lib/groupEpisodes'
import type {AnimeView, EpisodeView, PlaybackEvent} from '../lib/types'
import {LibraryMatchSheet, type LibraryMatchPicker} from '../components/LibraryMatchSheet'
import {LibraryEpisodeList} from '../components/LibraryEpisodeList'
import {LibraryPosterGrid} from '../components/LibraryPosterGrid'
import {IconBack} from '../components/Icons'
import {Button} from '@/components/ui/button'

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
  const [picker, setPicker] = useState<LibraryMatchPicker | null>(null)
  const skippedMatchIds = useRef(new Set<number>())

  const libraryEpisodes = useMemo(() => visibleLibraryEpisodes(episodes), [episodes])
  const shows = useMemo(() => groupEpisodes(libraryEpisodes), [libraryEpisodes])
  const selectedShow = shows.find((show) => show.key === selectedKey) ?? null
  const playingShowKey = useMemo(() => {
    if (!playing) {
      return null
    }
    const owner = shows.find((show) => show.episodes.some((episode) => episode.id === playing.episodeId))
    return owner?.key ?? null
  }, [shows, playing])

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
      return
    }
    if (selectedKey && !shows.some((show) => show.key === selectedKey)) {
      setSelectedKey(null)
    }
  }, [shows, selectedKey])

  async function openMatcher(episode: EpisodeView, candidates: AnimeView[] = []) {
    const query = (episode.displayTitle || episode.animeTitle).replace(/\s+—\s+Episode\s+\d+\s*$/i, '').trim()
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
    if (loading || picker || importing) {
      return
    }
    const unbound = libraryEpisodes.find((episode) => {
      return !episode.bound && !skippedMatchIds.current.has(episode.id)
    })
    if (!unbound) {
      return
    }
    void openMatcher(unbound)
  }, [loading, picker, importing, libraryEpisodes])

  async function onImport() {
    setImporting(true)
    try {
      const result = await ImportLocalFile()
      if (!result?.episode?.id) {
        return
      }
      await reload()
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
    if (!show.bound) {
      const episode = show.episodes[0]
      if (episode) {
        skippedMatchIds.current.delete(episode.id)
        void openMatcher(episode)
      }
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
    setLoading(true)
    void reload()
  }

  return (
    <section className="flex h-full min-h-0 flex-col">
      <header className="flex shrink-0 flex-wrap items-center justify-between gap-3 px-5 pt-5 pb-3">
        {selectedShow ? (
          <div className="flex min-w-0 items-center gap-2">
            <Button type="button" variant="ghost" onClick={backToGrid} style={{gap: 8, paddingInline: 12}}>
              <span className="shrink-0" style={{width: 16, height: 16}}>
                <IconBack className="size-full" />
              </span>
              Back
            </Button>
            <h2 className="min-w-0 truncate text-2xl font-semibold tracking-tight">{selectedShow.title}</h2>
          </div>
        ) : (
          <h2 className="text-2xl font-semibold tracking-tight">Library</h2>
        )}
        <Button type="button" variant="secondary" onClick={() => void onImport()} disabled={importing}>
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
          <LibraryEpisodeList
            show={selectedShow}
            playing={playing}
            busyId={busyId}
            onPlay={(episodeId) => void onPlay(episodeId)}
          />
        ) : (
          <LibraryPosterGrid
            loading={loading}
            loadError={loadError}
            shows={shows}
            highlightedKey={playingShowKey}
            onSelectShow={selectShow}
            onRetry={retryLoad}
          />
        )}
      </div>
    </section>
  )
}
