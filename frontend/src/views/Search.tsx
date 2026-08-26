import {useState} from 'react'
import {SearchNyaa, StartMagnet} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {NyaaResultView} from '../lib/types'

type Props = {
  notice: (msg: string, isError?: boolean) => void
  onDownloads: () => void
}

const dateFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: 'medium',
  timeStyle: 'short',
})

export function SearchView({notice, onDownloads}: Props) {
  const [query, setQuery] = useState('')
  const [submittedQuery, setSubmittedQuery] = useState('')
  const [results, setResults] = useState<NyaaResultView[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [starting, setStarting] = useState<number | null>(null)

  async function search(searchQuery = query) {
    const trimmed = searchQuery.trim()
    if (!trimmed) {
      setError('Enter an anime title to search.')
      return
    }
    setLoading(true)
    setError('')
    setSubmittedQuery(trimmed)
    try {
      setResults((await SearchNyaa(trimmed)) ?? [])
    } catch (err) {
      const message = errorMessage(err)
      setError(message)
      notice(message, true)
    } finally {
      setLoading(false)
    }
  }

  async function download(result: NyaaResultView, index: number) {
    setStarting(index)
    try {
      await StartMagnet(result.magnet)
      notice('Download added')
      onDownloads()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setStarting(null)
    }
  }

  return (
    <section className="flex h-full flex-col gap-6">
      <header>
        <h2 className="text-2xl font-semibold">Search Nyaa</h2>
        <p className="mt-1 text-sm text-muted-foreground">Find English-translated anime torrents from the latest RSS results.</p>
      </header>

      <form
        className="flex flex-wrap gap-2 rounded-xl bg-card p-4"
        onSubmit={(event) => {
          event.preventDefault()
          void search()
        }}
      >
        <label htmlFor="nyaa-query" className="sr-only">Anime title</label>
        <input
          id="nyaa-query"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Search anime title"
          className="min-h-11 min-w-0 flex-1 rounded-lg border border-border/40 bg-muted px-3 text-sm text-foreground focus-visible:outline-2 focus-visible:outline-ring"
        />
        <button
          type="submit"
          disabled={loading}
          className="min-h-11 cursor-pointer rounded-lg bg-accent px-4 text-sm font-medium text-on-accent disabled:cursor-not-allowed disabled:opacity-50"
        >
          {loading ? 'Searching…' : 'Search'}
        </button>
      </form>

      {loading ? (
        <p className="rounded-xl border border-border/40 bg-card p-8 text-sm text-muted-foreground" role="status">
          Loading Nyaa results…
        </p>
      ) : error ? (
        <div className="rounded-xl border border-destructive/60 bg-card p-8" role="alert">
          <p className="text-sm text-destructive">{error}</p>
          {submittedQuery && (
            <button
              type="button"
              onClick={() => void search(submittedQuery)}
              className="mt-4 min-h-11 cursor-pointer rounded-lg bg-secondary px-4 text-sm text-on-secondary"
            >
              Try again
            </button>
          )}
        </div>
      ) : results.length === 0 ? (
        <p className="rounded-xl border border-dashed border-border/40 p-8 text-sm text-muted-foreground">
          Search for an anime to see available torrents.
        </p>
      ) : (
        <ul className="flex flex-col gap-3">
          {results.map((result, index) => (
            <li key={`${result.magnet}-${index}`} className="rounded-xl bg-card p-4">
              <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div className="min-w-0">
                  <p className="break-words font-medium">{result.title}</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {formatPublished(result.published)} · {result.size || 'Unknown size'}
                  </p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {result.seeders} seeders · {result.leechers} leechers · {result.downloads} downloads
                  </p>
                  {(result.trusted || result.remake) && (
                    <p className="mt-2 text-xs text-accent">
                      {[result.trusted && 'Trusted', result.remake && 'Remake'].filter(Boolean).join(' · ')}
                    </p>
                  )}
                </div>
                <button
                  type="button"
                  onClick={() => void download(result, index)}
                  disabled={starting !== null}
                  className="min-h-11 shrink-0 cursor-pointer rounded-lg bg-secondary px-4 text-sm text-on-secondary disabled:cursor-not-allowed disabled:opacity-50"
                >
                  {starting === index ? 'Adding…' : 'Download'}
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}

function formatPublished(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return 'Unknown date'
  }
  return dateFormatter.format(date)
}
