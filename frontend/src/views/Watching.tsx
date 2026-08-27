import {useEffect, useState} from 'react'
import {ListCurrentlyWatching} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {WatchingEntryView} from '../lib/types'
import {Alert, AlertAction, AlertDescription} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'
import {Card} from '@/components/ui/card'

type Props = {
  notice: (msg: string, isError?: boolean) => void
  refreshKey: number
  onSettings: () => void
}

export function WatchingView({notice, refreshKey, onSettings}: Props) {
  const [entries, setEntries] = useState<WatchingEntryView[]>([])
  const [loading, setLoading] = useState(true)
  const [notConnected, setNotConnected] = useState(false)
  const [error, setError] = useState('')

  async function load() {
    setLoading(true)
    setNotConnected(false)
    setError('')
    try {
      const result = await ListCurrentlyWatching()
      setEntries(result ?? [])
    } catch (err) {
      const message = errorMessage(err)
      if (message === 'AniList not connected') {
        setNotConnected(true)
      } else {
        setError(message)
        notice(message, true)
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [refreshKey])

  return (
    <section className="flex h-full flex-col gap-6">
      <header>
        <h2 className="text-2xl font-semibold">Currently Watching</h2>
        <p className="mt-1 text-sm text-muted-foreground">Your AniList anime list and watch progress.</p>
      </header>

      {loading ? (
        <Card className="border border-border/40 p-8" role="status">
          Loading your list…
        </Card>
      ) : notConnected ? (
        <Card>
          <h3 className="font-medium">Connect AniList to see your list</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Sign in with AniList from Settings, then return here to load your currently watching anime.
          </p>
          <Button type="button" variant="secondary" className="mt-4" onClick={onSettings}>
            Open Settings
          </Button>
        </Card>
      ) : error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
          <AlertAction>
            <Button type="button" variant="secondary" onClick={() => void load()}>
              Try again
            </Button>
          </AlertAction>
        </Alert>
      ) : entries.length === 0 ? (
        <p className="border border-dashed border-border/40 p-8 text-sm text-muted-foreground">
          Nothing on your Currently Watching list.
        </p>
      ) : (
        <ul className="grid gap-3 lg:grid-cols-2">
          {entries.map((entry) => {
            const title = entry.titleEnglish || entry.titleRomaji
            const total = entry.totalEpisodes > 0 ? entry.totalEpisodes : '?'
            return (
              <li key={entry.mediaId}>
                <Card className="flex-row items-center gap-4 p-3">
                  {entry.coverImage ? (
                    <img src={entry.coverImage} alt="" width={48} height={64} className="h-16 w-12 object-cover" />
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
                </Card>
              </li>
            )
          })}
        </ul>
      )}
    </section>
  )
}
