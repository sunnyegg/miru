import {useEffect, useMemo, useState} from 'react'
import {InspectTorrent, StartTorrent} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {TorrentContentsView, TorrentFileView} from '../lib/types'
import {useFeedStore} from '../stores/feedStore'
import {useNavigationStore} from '../stores/navigationStore'
import {TorrentFileSheet} from './TorrentFileSheet'
import {Alert, AlertDescription} from '@/components/ui/alert'
import {Badge} from '@/components/ui/badge'
import {Button} from '@/components/ui/button'
import {Card} from '@/components/ui/card'
import {Input} from '@/components/ui/input'
import {Label} from '@/components/ui/label'
import {Skeleton} from '@/components/ui/skeleton'

type Props = {
  notice: (msg: string, isError?: boolean) => void
}

const dateFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: 'medium',
  timeStyle: 'short',
})

export function FeedSubscriptions({notice}: Props) {
  const goToDownloads = useNavigationStore((state) => state.setTab)
  const feeds = useFeedStore((state) => state.feeds)
  const items = useFeedStore((state) => state.items)
  const showNewOnly = useFeedStore((state) => state.showNewOnly)
  const loading = useFeedStore((state) => state.loading)
  const error = useFeedStore((state) => state.error)
  const setShowNewOnly = useFeedStore((state) => state.setShowNewOnly)
  const reload = useFeedStore((state) => state.reload)
  const addFeed = useFeedStore((state) => state.addFeed)
  const removeFeed = useFeedStore((state) => state.removeFeed)
  const toggleFeed = useFeedStore((state) => state.toggleFeed)
  const pollNow = useFeedStore((state) => state.pollNow)
  const markAllSeen = useFeedStore((state) => state.markAllSeen)
  const markSeen = useFeedStore((state) => state.markSeen)

  const [feedURL, setFeedURL] = useState('')
  const [feedTitle, setFeedTitle] = useState('')
  const [adding, setAdding] = useState(false)
  const [polling, setPolling] = useState(false)
  const [starting, setStarting] = useState<number | null>(null)
  const [picker, setPicker] = useState<{
    source: string
    contents: TorrentContentsView
    loading: boolean
    error: string
  } | null>(null)
  const [confirming, setConfirming] = useState(false)

  useEffect(() => {
    void reload()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function submitFeed() {
    setAdding(true)
    try {
      const added = await addFeed(feedURL, feedTitle, notice)
      if (added) {
        setFeedURL('')
        setFeedTitle('')
      }
    } finally {
      setAdding(false)
    }
  }

  async function pollFeedsNow() {
    setPolling(true)
    try {
      await pollNow(notice)
    } finally {
      setPolling(false)
    }
  }

  async function download(item: typeof items[number], itemIndex: number) {
    const source = item.magnet || item.link
    if (!source) {
      notice('This item has no torrent link', true)
      return
    }
    setStarting(itemIndex)
    setPicker({
      source,
      contents: {name: item.title, bytesTotal: 0, files: []},
      loading: true,
      error: '',
    })
    try {
      const contents = await InspectTorrent(source)
      setPicker((current) => {
        if (!current || current.source !== source) {
          return current
        }
        return {
          source,
          contents: contents ?? {name: item.title, bytesTotal: 0, files: []},
          loading: false,
          error: '',
        }
      })
    } catch (err) {
      setPicker((current) => {
        if (!current || current.source !== source) {
          return current
        }
        return {
          source,
          contents: {name: item.title, bytesTotal: 0, files: []},
          loading: false,
          error: errorMessage(err),
        }
      })
    } finally {
      setStarting(null)
    }
  }

  async function confirmPicker(files: TorrentFileView[]) {
    if (!picker) {
      return
    }
    setConfirming(true)
    try {
      await StartTorrent(picker.source, files)
      setPicker(null)
      notice('Download added')
      goToDownloads('downloads')
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setConfirming(false)
    }
  }

  const totalNew = useMemo(
    () => feeds.reduce((sum, feed) => sum + feed.newCount, 0),
    [feeds],
  )

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-6 overflow-hidden">
      <form
        className="flex shrink-0 flex-wrap gap-2 bg-card p-4"
        onSubmit={(event) => {
          event.preventDefault()
          void submitFeed()
        }}
      >
        <div className="min-w-0 flex-1 space-y-2">
          <Label htmlFor="feed-url">RSS feed URL</Label>
          <Input
            id="feed-url"
            value={feedURL}
            onChange={(event) => setFeedURL(event.target.value)}
            placeholder="https://…/feed.xml or Nyaa/Tokyo Toshokan RSS"
            className="bg-background"
          />
          <p className="text-xs text-muted-foreground">
            Any http/https RSS feed works — Nyaa, Tokyo Toshokan, or a fansub site feed URL.
          </p>
        </div>
        <div className="w-full min-w-48 space-y-2 sm:w-auto sm:flex-1">
          <Label htmlFor="feed-title">Title (optional)</Label>
          <Input
            id="feed-title"
            value={feedTitle}
            onChange={(event) => setFeedTitle(event.target.value)}
            placeholder="Fansub name"
            className="bg-background"
          />
        </div>
        <div className="flex w-full items-end sm:w-auto">
          <Button type="submit" disabled={adding}>
            {adding ? 'Adding…' : 'Subscribe'}
          </Button>
        </div>
      </form>

      <div className="flex shrink-0 flex-wrap items-center gap-2">
        <Button type="button" variant="muted" disabled={polling} onClick={() => void pollFeedsNow()}>
          {polling ? 'Polling…' : 'Poll now'}
        </Button>
        <Button
          type="button"
          variant="muted"
          disabled={totalNew === 0}
          onClick={() => void markAllSeen(notice)}
        >
          Mark all seen
        </Button>
        <Button
          type="button"
          variant={showNewOnly ? 'secondary' : 'muted'}
          onClick={() => void setShowNewOnly(!showNewOnly)}
        >
          {showNewOnly ? 'Showing new' : 'Showing all'}
        </Button>
        {totalNew > 0 && <Badge>{totalNew} new</Badge>}
      </div>

      <div className="min-h-0 flex-1 overflow-auto">
        {loading ? (
          <ul className="flex flex-col gap-2" aria-busy="true">
            {Array.from({length: 4}, (_, index) => (
              <li key={index} className="bg-card p-3">
                <Skeleton className="h-4 w-2/3 animate-pulse" />
              </li>
            ))}
          </ul>
        ) : error ? (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : (
          <div className="flex flex-col gap-6">
            <section>
              <h3 className="text-sm font-medium">Subscriptions</h3>
              {feeds.length === 0 ? (
                <p className="mt-2 text-sm text-muted-foreground">
                  Add Nyaa, Tokyo Toshokan, or fansub RSS URLs to poll in the background.
                </p>
              ) : (
                <ul className="mt-3 flex flex-col gap-2">
                  {feeds.map((feed) => (
                    <li key={feed.id}>
                      <Card>
                        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                          <div className="min-w-0">
                            <p className="font-medium">{feed.title || feed.url}</p>
                            <p className="mt-1 break-all text-xs text-muted-foreground">{feed.url}</p>
                            <p className="mt-1 text-xs text-muted-foreground">
                              {feed.lastPolled
                                ? `Last polled ${formatDate(feed.lastPolled)}`
                                : 'Not polled yet'}
                              {feed.newCount > 0 && ` · ${feed.newCount} new`}
                            </p>
                          </div>
                          <div className="flex shrink-0 flex-wrap gap-2">
                            <Button
                              type="button"
                              variant="muted"
                              onClick={() => void toggleFeed(feed, notice)}
                            >
                              {feed.enabled ? 'Pause' : 'Resume'}
                            </Button>
                            <Button
                              type="button"
                              variant="ghost"
                              className="text-destructive hover:text-destructive"
                              onClick={() => void removeFeed(feed, notice)}
                            >
                              Remove
                            </Button>
                          </div>
                        </div>
                      </Card>
                    </li>
                  ))}
                </ul>
              )}
            </section>

            <section>
              <h3 className="text-sm font-medium">Items</h3>
              {items.length === 0 ? (
                <p className="mt-2 text-sm text-muted-foreground">
                  {showNewOnly ? 'No new torrents from your feeds.' : 'No items stored yet.'}
                </p>
              ) : (
                <ul className="mt-3 flex flex-col gap-3">
                  {items.map((item, itemIndex) => (
                    <li key={item.id}>
                      <Card>
                        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                          <div className="min-w-0">
                            <p className="break-words font-medium">{item.title}</p>
                            <p className="mt-1 text-xs text-muted-foreground">
                              {item.feedTitle} · {formatDate(item.published)}
                            </p>
                            {item.isNew && <Badge className="mt-2">New</Badge>}
                          </div>
                          <div className="flex shrink-0 flex-wrap gap-2">
                            {item.isNew && (
                              <Button
                                type="button"
                                variant="muted"
                                onClick={() => void markSeen(item, notice)}
                              >
                                Mark seen
                              </Button>
                            )}
                            <Button
                              type="button"
                              variant="secondary"
                              onClick={() => void download(item, itemIndex)}
                              disabled={starting !== null}
                            >
                              {starting === itemIndex ? 'Adding…' : 'Download'}
                            </Button>
                          </div>
                        </div>
                      </Card>
                    </li>
                  ))}
                </ul>
              )}
            </section>
          </div>
        )}
      </div>

      {picker && (
        <TorrentFileSheet
          name={picker.contents.name}
          bytesTotal={picker.contents.bytesTotal}
          files={picker.contents.files ?? []}
          loading={picker.loading}
          error={picker.error}
          confirming={confirming}
          onClose={() => setPicker(null)}
          onConfirm={(files) => void confirmPicker(files)}
        />
      )}
    </div>
  )
}

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return 'Unknown date'
  }
  return dateFormatter.format(date)
}
