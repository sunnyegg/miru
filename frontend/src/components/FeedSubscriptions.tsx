import {useEffect, useState} from 'react'
import {
  AddRSSFeed,
  InspectTorrent,
  ListRSSFeedItems,
  ListRSSFeeds,
  MarkAllRSSFeedItemsSeen,
  MarkRSSFeedItemsSeen,
  PollRSSFeedsNow,
  RemoveRSSFeed,
  SetRSSFeedEnabled,
  StartTorrent,
} from '../../wailsjs/go/main/App'
import {EventsOff, EventsOn} from '../../wailsjs/runtime/runtime'
import {errorMessage} from '../lib/format'
import type {RSSFeedItemView, RSSFeedView, TorrentContentsView, TorrentFileView} from '../lib/types'
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
  onDownloads: () => void
}

const dateFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: 'medium',
  timeStyle: 'short',
})

export function FeedSubscriptions({notice, onDownloads}: Props) {
  const [feeds, setFeeds] = useState<RSSFeedView[]>([])
  const [items, setItems] = useState<RSSFeedItemView[]>([])
  const [feedURL, setFeedURL] = useState('')
  const [feedTitle, setFeedTitle] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [adding, setAdding] = useState(false)
  const [polling, setPolling] = useState(false)
  const [starting, setStarting] = useState<number | null>(null)
  const [showNewOnly, setShowNewOnly] = useState(true)
  const [picker, setPicker] = useState<{
    source: string
    contents: TorrentContentsView
    loading: boolean
    error: string
  } | null>(null)
  const [confirming, setConfirming] = useState(false)

  async function reload() {
    setError('')
    try {
      const [loadedFeeds, loadedItems] = await Promise.all([
        ListRSSFeeds(),
        ListRSSFeedItems(showNewOnly),
      ])
      setFeeds(loadedFeeds ?? [])
      setItems(loadedItems ?? [])
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void reload()
    EventsOn('feeds:updated', () => {
      void reload()
    })
    return () => {
      EventsOff('feeds:updated')
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [showNewOnly])

  async function addFeed() {
    const trimmedURL = feedURL.trim()
    if (!trimmedURL) {
      notice('Enter an RSS feed URL', true)
      return
    }
    setAdding(true)
    try {
      await AddRSSFeed(trimmedURL, feedTitle.trim())
      setFeedURL('')
      setFeedTitle('')
      notice('Feed subscribed')
      await reload()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setAdding(false)
    }
  }

  async function removeFeed(feed: RSSFeedView) {
    try {
      await RemoveRSSFeed(feed.id)
      notice('Feed removed')
      await reload()
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  async function toggleFeed(feed: RSSFeedView) {
    try {
      await SetRSSFeedEnabled(feed.id, !feed.enabled)
      await reload()
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  async function pollNow() {
    setPolling(true)
    try {
      await PollRSSFeedsNow()
      notice('Feeds polled')
      await reload()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setPolling(false)
    }
  }

  async function markAllSeen() {
    try {
      await MarkAllRSSFeedItemsSeen()
      await reload()
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  async function markSeen(item: RSSFeedItemView) {
    try {
      await MarkRSSFeedItemsSeen([item.id])
      await reload()
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  async function download(item: RSSFeedItemView, itemIndex: number) {
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
      onDownloads()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setConfirming(false)
    }
  }

  const totalNew = feeds.reduce((sum, feed) => sum + feed.newCount, 0)

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-6 overflow-hidden">
      <form
        className="flex shrink-0 flex-wrap gap-2 bg-card p-4"
        onSubmit={(event) => {
          event.preventDefault()
          void addFeed()
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
        <Button type="button" variant="muted" disabled={polling} onClick={() => void pollNow()}>
          {polling ? 'Polling…' : 'Poll now'}
        </Button>
        <Button
          type="button"
          variant="muted"
          disabled={totalNew === 0}
          onClick={() => void markAllSeen()}
        >
          Mark all seen
        </Button>
        <Button
          type="button"
          variant={showNewOnly ? 'secondary' : 'muted'}
          onClick={() => setShowNewOnly((current) => !current)}
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
                              onClick={() => void toggleFeed(feed)}
                            >
                              {feed.enabled ? 'Pause' : 'Resume'}
                            </Button>
                            <Button
                              type="button"
                              variant="ghost"
                              className="text-destructive hover:text-destructive"
                              onClick={() => void removeFeed(feed)}
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
                                onClick={() => void markSeen(item)}
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
