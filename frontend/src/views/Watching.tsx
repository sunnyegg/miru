import {useEffect, useState} from 'react'
import {ListAnimeList, SearchAnime, SetAnimeListStatus} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {AnimeView, WatchingEntryView} from '../lib/types'
import {Alert, AlertAction, AlertDescription} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'
import {Card} from '@/components/ui/card'
import {Input} from '@/components/ui/input'

type Props = {
  refreshKey: number
  notice: (msg: string, isError?: boolean) => void
  onSettings: () => void
}

type ListFilter = 'CURRENT' | 'COMPLETED'

function listStatusLabel(status: string): string {
  switch (status) {
    case 'CURRENT':
      return 'Watching'
    case 'COMPLETED':
      return 'Completed'
    case 'PLANNING':
      return 'Planning'
    case 'DROPPED':
      return 'Dropped'
    case 'PAUSED':
      return 'Paused'
    case 'REPEATING':
      return 'Repeating'
    default:
      return status
  }
}

export function WatchingView({refreshKey, notice, onSettings}: Props) {
  const [listFilter, setListFilter] = useState<ListFilter>('CURRENT')
  const [entries, setEntries] = useState<WatchingEntryView[]>([])
  const [loading, setLoading] = useState(true)
  const [notConnected, setNotConnected] = useState(false)
  const [error, setError] = useState('')
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<AnimeView[]>([])
  const [searching, setSearching] = useState(false)
  const [searchError, setSearchError] = useState('')
  const [busyMediaId, setBusyMediaId] = useState<number | null>(null)

  async function loadList(filter: ListFilter = listFilter) {
    setLoading(true)
    setNotConnected(false)
    setError('')
    try {
      const result = await ListAnimeList(filter)
      setEntries(result ?? [])
    } catch (err) {
      const message = errorMessage(err)
      if (message === 'AniList not connected') {
        setNotConnected(true)
        setEntries([])
      } else {
        setError(message)
      }
    } finally {
      setLoading(false)
    }
  }

  async function searchAnime() {
    const trimmed = searchQuery.trim()
    if (!trimmed) {
      setSearchError('Enter an anime title to search.')
      return
    }
    setSearching(true)
    setSearchError('')
    try {
      const found = await SearchAnime(trimmed)
      setSearchResults(found ?? [])
    } catch (err) {
      setSearchError(errorMessage(err))
      setSearchResults([])
    } finally {
      setSearching(false)
    }
  }

  async function setStatus(mediaId: number, status: ListFilter, totalEpisodes: number) {
    setBusyMediaId(mediaId)
    try {
      await SetAnimeListStatus(mediaId, status, totalEpisodes)
      const label = status === 'CURRENT' ? 'Watching' : 'Completed'
      notice(`Marked as ${label}`)
      setSearchResults((current) =>
        current.map((anime) =>
          anime.id === mediaId ? {...anime, listStatus: status} : anime
        )
      )
      await loadList(listFilter)
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBusyMediaId(null)
    }
  }

  function selectFilter(filter: ListFilter) {
    setListFilter(filter)
    void loadList(filter)
  }

  useEffect(() => {
    void loadList()
  }, [refreshKey])

  const filterLabel = listFilter === 'CURRENT' ? 'Currently Watching' : 'Completed'
  const emptyCopy =
    listFilter === 'CURRENT'
      ? 'Nothing on your Currently Watching list.'
      : 'Nothing on your Completed list.'

  return (
    <section className="flex h-full flex-col gap-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-2xl font-semibold">{filterLabel}</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Your AniList anime list and watch progress.
          </p>
        </div>
        <div className="flex gap-1">
          <Button
            type="button"
            variant={listFilter === 'CURRENT' ? 'muted' : 'ghost'}
            onClick={() => selectFilter('CURRENT')}
          >
            Watching
          </Button>
          <Button
            type="button"
            variant={listFilter === 'COMPLETED' ? 'muted' : 'ghost'}
            onClick={() => selectFilter('COMPLETED')}
          >
            Completed
          </Button>
        </div>
      </header>

      {!notConnected && (
        <Card className="border border-border/40 p-4">
          <h3 className="text-base font-medium">Search AniList</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Find anime and add it to your Currently Watching list.
          </p>
          <div className="mt-3 flex flex-wrap gap-2">
            <label className="sr-only" htmlFor="watching-anilist-search">
              Search AniList
            </label>
            <Input
              id="watching-anilist-search"
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
              className="min-w-0 flex-1 basis-56 border-border/40"
              placeholder="Anime title"
            />
            <Button
              type="button"
              variant="secondary"
              onClick={() => void searchAnime()}
              disabled={searching || !searchQuery.trim()}
            >
              {searching ? 'Searching…' : 'Search'}
            </Button>
          </div>
          {searchError && (
            <p className="mt-2 text-sm text-destructive">{searchError}</p>
          )}
          {searchResults.length > 0 && (
            <ul className="mt-3 flex flex-col gap-1">
              {searchResults.map((anime) => {
                const title = anime.titleEnglish || anime.titleRomaji
                const onList = anime.listStatus === 'CURRENT'
                const listCaption = anime.listStatus
                  ? listStatusLabel(anime.listStatus)
                  : 'Not on your list'
                return (
                  <li key={anime.id}>
                    <div className="flex min-h-11 items-center gap-3 bg-muted px-3 py-2">
                      {anime.coverImage ? (
                        <img
                          src={anime.coverImage}
                          alt=""
                          width={40}
                          height={56}
                          referrerPolicy="no-referrer"
                          className="shrink-0 object-cover"
                          style={{width: 40, height: 56}}
                        />
                      ) : (
                        <span className="shrink-0 bg-muted" style={{width: 40, height: 56}} />
                      )}
                      <div className="min-w-0 flex-1">
                        <p className="truncate font-medium">{title}</p>
                        <p className="mt-0.5 text-xs text-muted-foreground">{listCaption}</p>
                      </div>
                      {!onList && (
                        <Button
                          type="button"
                          onClick={() => void setStatus(anime.id, 'CURRENT', 0)}
                          disabled={busyMediaId !== null}
                          aria-busy={busyMediaId === anime.id}
                        >
                          {busyMediaId === anime.id ? 'Saving…' : 'Watch'}
                        </Button>
                      )}
                    </div>
                  </li>
                )
              })}
            </ul>
          )}
        </Card>
      )}

      {loading ? (
        <Card className="border border-border/40 p-8" role="status">
          Loading your list…
        </Card>
      ) : notConnected ? (
        <Card>
          <h3 className="font-medium">Connect AniList to see your list</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Sign in with AniList from Settings, then return here to load your anime list.
          </p>
          <Button type="button" variant="secondary" className="mt-4" onClick={onSettings}>
            Open Settings
          </Button>
        </Card>
      ) : error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
          <AlertAction>
            <Button type="button" variant="secondary" onClick={() => void loadList()}>
              Try again
            </Button>
          </AlertAction>
        </Alert>
      ) : entries.length === 0 ? (
        <p className="border border-dashed border-border/40 p-8 text-sm text-muted-foreground">
          {emptyCopy}
        </p>
      ) : (
        <ul className="grid gap-3 lg:grid-cols-2">
          {entries.map((entry) => {
            const title = entry.titleEnglish || entry.titleRomaji
            const total = entry.totalEpisodes > 0 ? entry.totalEpisodes : '?'
            const busy = busyMediaId === entry.mediaId
            return (
              <li key={entry.mediaId}>
                <Card className="flex-row items-center gap-4 p-3">
                  {entry.coverImage ? (
                    <img
                      src={entry.coverImage}
                      alt=""
                      width={48}
                      height={64}
                      className="h-16 w-12 object-cover"
                    />
                  ) : (
                    <span className="h-16 w-12 bg-muted" aria-hidden="true" />
                  )}
                  <div className="min-w-0 flex-1">
                    <p className="truncate font-medium">{title}</p>
                    <p className="mt-1 text-sm text-muted-foreground">
                      Episode {entry.progress} / {total}
                    </p>
                    {entry.mediaStatus && (
                      <p className="mt-1 text-xs text-muted-foreground">{entry.mediaStatus}</p>
                    )}
                  </div>
                  {listFilter === 'CURRENT' ? (
                    <Button
                      type="button"
                      variant="secondary"
                      onClick={() => void setStatus(entry.mediaId, 'COMPLETED', entry.totalEpisodes)}
                      disabled={busyMediaId !== null}
                      aria-busy={busy}
                    >
                      {busy ? 'Saving…' : 'Completed'}
                    </Button>
                  ) : (
                    <Button
                      type="button"
                      variant="secondary"
                      onClick={() => void setStatus(entry.mediaId, 'CURRENT', 0)}
                      disabled={busyMediaId !== null}
                      aria-busy={busy}
                    >
                      {busy ? 'Saving…' : 'Watching'}
                    </Button>
                  )}
                </Card>
              </li>
            )
          })}
        </ul>
      )}
    </section>
  )
}
