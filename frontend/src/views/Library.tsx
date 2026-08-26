import {useEffect, useState} from 'react'
import {
  BindEpisode,
  ImportLocalFile,
  ListEpisodes,
  PlayEpisode,
  SearchAnime,
} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {AnimeView, EpisodeView} from '../lib/types'
import {IconPlay} from '../components/Icons'

type Props = {
  notice: (msg: string, isError?: boolean) => void
  refreshKey: number
}

export function LibraryView({notice, refreshKey}: Props) {
  const [episodes, setEpisodes] = useState<EpisodeView[]>([])
  const [busyId, setBusyId] = useState<number | null>(null)
  const [importing, setImporting] = useState(false)
  const [picker, setPicker] = useState<{episode: EpisodeView; candidates: AnimeView[]; query: string} | null>(null)

  async function reload() {
    try {
      const rows = await ListEpisodes()
      setEpisodes(rows ?? [])
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  useEffect(() => {
    void reload()
  }, [refreshKey])

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

  return (
    <section className="flex h-full flex-col gap-6">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-2xl font-semibold">Library</h2>
          <p className="mt-1 text-sm text-muted-foreground">Local episodes, ready to play in MPV.</p>
        </div>
        <button
          type="button"
          onClick={() => void onImport()}
          disabled={importing}
          className="inline-flex min-h-11 cursor-pointer items-center rounded-lg bg-accent px-4 text-sm font-medium text-on-accent transition-opacity duration-200 hover:opacity-90 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:cursor-not-allowed disabled:opacity-50"
        >
          {importing ? 'Importing…' : 'Import file'}
        </button>
      </header>

      {picker && (
        <div className="rounded-xl border border-border/50 bg-card p-4" role="dialog" aria-labelledby="match-title">
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
              className="min-h-11 min-w-56 flex-1 rounded-lg border border-border/40 bg-muted px-3 text-sm text-foreground focus-visible:outline-2 focus-visible:outline-ring"
            />
            <button
              type="button"
              onClick={() => void onSearch()}
              className="min-h-11 cursor-pointer rounded-lg bg-secondary px-4 text-sm text-on-secondary"
            >
              Search
            </button>
          </div>
          <ul className="mt-3 flex flex-col gap-2">
            {picker.candidates.map((anime) => (
              <li key={anime.id}>
                <button
                  type="button"
                  onClick={() => void onBind(anime.id)}
                  className="flex min-h-11 w-full cursor-pointer items-center gap-3 rounded-lg bg-muted px-3 py-2 text-left hover:bg-secondary"
                >
                  {anime.coverImage ? (
                    <img src={anime.coverImage} alt="" width={40} height={56} className="h-14 w-10 rounded object-cover" />
                  ) : (
                    <span className="h-14 w-10 rounded bg-primary" />
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

      {episodes.length === 0 ? (
        <p className="rounded-xl border border-dashed border-border/40 p-8 text-sm text-muted-foreground">
          No episodes yet. Import a local file or finish a torrent download.
        </p>
      ) : (
        <ul className="flex flex-col gap-3">
          {episodes.map((ep) => (
            <li key={ep.id} className="flex items-center gap-4 rounded-xl bg-card p-3">
              {ep.coverImage ? (
                <img src={ep.coverImage} alt="" width={48} height={64} className="h-16 w-12 rounded object-cover" />
              ) : (
                <span className="h-16 w-12 rounded bg-muted" />
              )}
              <div className="min-w-0 flex-1">
                <p className="truncate font-medium">{ep.animeTitle || ep.displayTitle}</p>
                <p className="truncate text-sm text-muted-foreground">
                  {ep.episodeNumber > 0 ? `Episode ${ep.episodeNumber}` : ep.displayTitle}
                  {ep.bound ? '' : ' · not linked'}
                </p>
              </div>
              <button
                type="button"
                onClick={() => void onPlay(ep.id)}
                disabled={busyId === ep.id}
                className="inline-flex min-h-11 min-w-11 cursor-pointer items-center justify-center gap-2 rounded-lg bg-accent px-4 text-sm font-medium text-on-accent disabled:cursor-not-allowed disabled:opacity-50"
              >
                <IconPlay className="h-4 w-4" />
                {busyId === ep.id ? 'Starting…' : 'Play'}
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
