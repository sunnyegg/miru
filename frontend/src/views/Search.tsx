import {useState} from 'react'
import {SearchNyaa, SearchTokyoToshokan, StartTorrentURL} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {NyaaResultView} from '../lib/types'
import {Alert, AlertAction, AlertDescription} from '@/components/ui/alert'
import {Badge} from '@/components/ui/badge'
import {Button} from '@/components/ui/button'
import {Card} from '@/components/ui/card'
import {Input} from '@/components/ui/input'
import {NativeSelect, NativeSelectOption} from '@/components/ui/native-select'

type Props = {
  notice: (msg: string, isError?: boolean) => void
  onDownloads: () => void
}

type SearchSource = 'nyaa' | 'tokyotosho'

const dateFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: 'medium',
  timeStyle: 'short',
})

export function SearchView({notice, onDownloads}: Props) {
  const [query, setQuery] = useState('')
  const [source, setSource] = useState<SearchSource>('nyaa')
  const [submittedQuery, setSubmittedQuery] = useState('')
  const [results, setResults] = useState<NyaaResultView[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [starting, setStarting] = useState<number | null>(null)

  const sourceLabel = source === 'tokyotosho' ? 'Tokyo Toshokan' : 'Nyaa'

  async function search(searchQuery = query, searchSource = source) {
    const trimmed = searchQuery.trim()
    if (!trimmed) {
      setError('Enter an anime title to search.')
      return
    }
    setLoading(true)
    setError('')
    setSubmittedQuery(trimmed)
    try {
      const found = searchSource === 'tokyotosho'
        ? await SearchTokyoToshokan(trimmed)
        : await SearchNyaa(trimmed)
      setResults(found ?? [])
    } catch (err) {
      const message = errorMessage(err)
      setError(message)
      notice(message, true)
    } finally {
      setLoading(false)
    }
  }

  async function changeSource(nextSource: SearchSource) {
    setSource(nextSource)
    if (submittedQuery) {
      await search(submittedQuery, nextSource)
    }
  }

  async function download(result: NyaaResultView, index: number) {
    setStarting(index)
    try {
      await StartTorrentURL(result.link)
      notice('Download added')
      onDownloads()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setStarting(null)
    }
  }

  return (
    <section className="flex h-full min-h-0 flex-col gap-6 overflow-hidden">
      <header className="shrink-0">
        <h2 className="text-2xl font-semibold">Search</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Find English-translated anime torrents from {sourceLabel} RSS results.
        </p>
      </header>

      <form
        className="flex shrink-0 flex-wrap gap-2 bg-card p-4"
        onSubmit={(event) => {
          event.preventDefault()
          void search()
        }}
      >
        <label htmlFor="search-source" className="sr-only">Indexer</label>
        <NativeSelect
          id="search-source"
          className="w-auto min-w-44 shrink-0"
          value={source}
          onChange={(event) => void changeSource(event.target.value as SearchSource)}
        >
          <NativeSelectOption value="nyaa">Nyaa</NativeSelectOption>
          <NativeSelectOption value="tokyotosho">Tokyo Toshokan</NativeSelectOption>
        </NativeSelect>
        <label htmlFor="search-query" className="sr-only">Anime title</label>
        <Input
          id="search-query"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Search anime title"
          className="min-w-0 flex-1 border-border/40"
        />
        <Button type="submit" disabled={loading}>
          {loading ? 'Searching…' : 'Search'}
        </Button>
      </form>

      <div className="min-h-0 flex-1 overflow-auto">
        {loading ? (
          <Card className="border border-border/40 p-8" role="status">
            Loading {sourceLabel} results…
          </Card>
        ) : error ? (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
            {submittedQuery && (
              <AlertAction>
                <Button type="button" variant="secondary" onClick={() => void search(submittedQuery)}>
                  Try again
                </Button>
              </AlertAction>
            )}
          </Alert>
        ) : results.length === 0 ? (
          <p className="border border-dashed border-border/40 p-8 text-sm text-muted-foreground">
            Search for an anime to see available torrents.
          </p>
        ) : (
          <ul className="flex flex-col gap-3">
            {results.map((result, index) => {
              const publishedDate = new Date(result.published)
              const publishedLabel = Number.isNaN(publishedDate.getTime())
                ? 'Unknown date'
                : dateFormatter.format(publishedDate)
              const hasPeerCounts = result.seeders > 0 || result.leechers > 0 || result.downloads > 0
              return (
                <li key={`${result.magnet || result.link}-${index}`}>
                  <Card>
                    <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                      <div className="min-w-0">
                        <p className="break-words font-medium">{result.title}</p>
                        <p className="mt-1 text-xs text-muted-foreground">
                          {publishedLabel} · {result.size || 'Unknown size'}
                        </p>
                        {hasPeerCounts && (
                          <p className="mt-1 text-xs text-muted-foreground">
                            {result.seeders} seeders · {result.leechers} leechers · {result.downloads} downloads
                          </p>
                        )}
                        {(result.trusted || result.remake) && (
                          <p className="mt-2">
                            {result.trusted && <Badge className="mr-2">Trusted</Badge>}
                            {result.remake && <Badge>Remake</Badge>}
                          </p>
                        )}
                      </div>
                      <Button
                        type="button"
                        variant="secondary"
                        onClick={() => void download(result, index)}
                        disabled={starting !== null}
                      >
                        {starting === index ? 'Adding…' : 'Download'}
                      </Button>
                    </div>
                  </Card>
                </li>
              )
            })}
          </ul>
        )}
      </div>
    </section>
  )
}
