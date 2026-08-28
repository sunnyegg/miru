import {useEffect, useRef, useState} from 'react'
import {SearchNyaa, SearchTokyoToshokan, StartTorrentURL} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {NyaaResultView} from '../lib/types'
import {Alert, AlertAction, AlertDescription} from '@/components/ui/alert'
import {Badge} from '@/components/ui/badge'
import {Button} from '@/components/ui/button'
import {Card} from '@/components/ui/card'
import {Input} from '@/components/ui/input'
import {NativeSelect, NativeSelectOption} from '@/components/ui/native-select'
import {Skeleton} from '@/components/ui/skeleton'

type Props = {
  notice: (msg: string, isError?: boolean) => void
  onDownloads: () => void
  prefillQuery?: string
  onPrefillConsumed?: () => void
}

type SearchSource = 'nyaa' | 'tokyotosho'

const PAGE_SIZE = 10

const dateFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: 'medium',
  timeStyle: 'short',
})

export function SearchView({notice, onDownloads, prefillQuery, onPrefillConsumed}: Props) {
  const [query, setQuery] = useState('')
  const [source, setSource] = useState<SearchSource>('nyaa')
  const [submittedQuery, setSubmittedQuery] = useState('')
  const [results, setResults] = useState<NyaaResultView[]>([])
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [starting, setStarting] = useState<number | null>(null)
  const resultsScrollRef = useRef<HTMLDivElement>(null)
  const prefillHandled = useRef('')

  const sourceLabel = source === 'tokyotosho' ? 'Tokyo Toshokan' : 'Nyaa'
  const pageStart = (page - 1) * PAGE_SIZE
  const pageResults = results.slice(pageStart, pageStart + PAGE_SIZE)
  const lastPage = Math.max(1, Math.ceil(results.length / PAGE_SIZE))
  const showPager = results.length > PAGE_SIZE

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
      setPage(1)
      resultsScrollRef.current?.scrollTo({top: 0})
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

  async function download(result: NyaaResultView, resultIndex: number) {
    setStarting(resultIndex)
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

  function goToPage(nextPage: number) {
    setPage(nextPage)
    resultsScrollRef.current?.scrollTo({top: 0})
  }

  useEffect(() => {
    const trimmed = prefillQuery?.trim()
    if (!trimmed || prefillHandled.current === trimmed) {
      return
    }
    prefillHandled.current = trimmed
    if (!query.trim()) {
      setQuery(trimmed)
    }
    void search(trimmed)
    onPrefillConsumed?.()
    // Fires only when prefillQuery changes; query/search/onPrefillConsumed are read but not triggers.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [prefillQuery])

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
          className="min-w-0 flex-1"
        />
        <Button type="submit" disabled={loading}>
          {loading ? 'Searching…' : 'Search'}
        </Button>
      </form>

      <div ref={resultsScrollRef} className="min-h-0 flex-1 overflow-auto">
        {loading ? (
          <ul className="flex flex-col gap-2" aria-busy="true" aria-label={`Loading ${sourceLabel} results`}>
            {Array.from({length: 5}, (_, i) => (
              <li key={i} className="flex items-center gap-3 bg-card p-3">
                <Skeleton className="h-12 w-12 shrink-0 animate-pulse" />
                <div className="flex flex-1 flex-col gap-2">
                  <Skeleton className="h-4 w-2/3 animate-pulse" />
                  <Skeleton className="h-3 w-1/3 animate-pulse" />
                </div>
              </li>
            ))}
          </ul>
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
            {pageResults.map((result, indexOnPage) => {
              const resultIndex = pageStart + indexOnPage
              const publishedDate = new Date(result.published)
              const publishedLabel = Number.isNaN(publishedDate.getTime())
                ? 'Unknown date'
                : dateFormatter.format(publishedDate)
              const hasPeerCounts = result.seeders > 0 || result.leechers > 0 || result.downloads > 0
              return (
                <li key={`${result.magnet || result.link}-${resultIndex}`}>
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
                        onClick={() => void download(result, resultIndex)}
                        disabled={starting !== null}
                      >
                        {starting === resultIndex ? 'Adding…' : 'Download'}
                      </Button>
                    </div>
                  </Card>
                </li>
              )
            })}
          </ul>
        )}
      </div>

      {showPager && !loading && !error && (
        <div className="flex shrink-0 flex-wrap items-center justify-between gap-2">
          <Button
            type="button"
            variant="muted"
            disabled={page <= 1}
            onClick={() => goToPage(page - 1)}
          >
            Previous
          </Button>
          <p className="text-sm text-muted-foreground">
            {pageStart + 1}–{pageStart + pageResults.length} of {results.length}
          </p>
          <Button
            type="button"
            variant="muted"
            disabled={page >= lastPage}
            onClick={() => goToPage(page + 1)}
          >
            Next
          </Button>
        </div>
      )}
    </section>
  )
}
