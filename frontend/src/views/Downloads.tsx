import {useEffect, useState} from 'react'
import {
  CancelDownload,
  DownloadHistory,
  FinishDownload,
  OpenDownloadFolder,
  PauseDownload,
  RemoveDownload,
  ResumeDownload,
  ResumeSeeding,
  StartMagnet,
  StartTorrentFile,
} from '../../wailsjs/go/main/App'
import {errorMessage, formatBytes, formatSpeed} from '../lib/format'
import type {DownloadView} from '../lib/types'
import {IconFolder} from '../components/Icons'
import {Alert, AlertAction, AlertDescription} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'
import {Card} from '@/components/ui/card'
import {Label} from '@/components/ui/label'
import {Progress} from '@/components/ui/progress'
import {Textarea} from '@/components/ui/textarea'

type Props = {
  notice: (msg: string, isError?: boolean) => void
  jobs: DownloadView[]
  onJobs: (jobs: DownloadView[]) => void
}

export function DownloadsView({notice, jobs, onJobs}: Props) {
  const [magnet, setMagnet] = useState('')
  const [busy, setBusy] = useState(false)
  const [pendingDeleteId, setPendingDeleteId] = useState<number | null>(null)
  const [loadError, setLoadError] = useState('')

  useEffect(() => {
    void refreshHistory()
    // Mount-only load; refreshHistory is recreated each render but does not depend on it.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function refreshHistory() {
    setLoadError('')
    try {
      onJobs((await DownloadHistory()) ?? [])
    } catch (err) {
      setLoadError(errorMessage(err))
    }
  }

  async function startMagnet() {
    setBusy(true)
    try {
      await StartMagnet(magnet.trim())
      setMagnet('')
      await refreshHistory()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBusy(false)
    }
  }

  async function startFile() {
    setBusy(true)
    try {
      await StartTorrentFile()
      await refreshHistory()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBusy(false)
    }
  }

  async function cancel(id: number) {
    setBusy(true)
    try {
      await CancelDownload(id)
      await refreshHistory()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBusy(false)
    }
  }

  async function pause(id: number) {
    setBusy(true)
    try {
      await PauseDownload(id)
      await refreshHistory()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBusy(false)
    }
  }

  async function resume(id: number) {
    setBusy(true)
    try {
      await ResumeDownload(id)
      await refreshHistory()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBusy(false)
    }
  }

  async function remove(id: number, deleteFiles: boolean) {
    setBusy(true)
    try {
      await RemoveDownload(id, deleteFiles)
      setPendingDeleteId(null)
      await refreshHistory()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBusy(false)
    }
  }

  async function resumeSeeding(id: number) {
    setBusy(true)
    try {
      await ResumeSeeding(id)
      await refreshHistory()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBusy(false)
    }
  }

  async function finish(id: number) {
    setBusy(true)
    try {
      await FinishDownload(id)
      await refreshHistory()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBusy(false)
    }
  }

  async function openFolder() {
    try {
      await OpenDownloadFolder()
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  return (
    <section className="flex h-full flex-col gap-6">
      <header>
        <h2 className="text-2xl font-semibold">Downloads</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Up to your max concurrent setting can download at once. Extra torrents wait in queue. Completed files seed until 0.5x upload ratio.
        </p>
      </header>

      <form
        className="flex flex-col gap-3 bg-card p-4"
        onSubmit={(e) => {
          e.preventDefault()
          void startMagnet()
        }}
      >
        <Label htmlFor="magnet">Magnet link</Label>
        <Textarea
          id="magnet"
          value={magnet}
          onChange={(e) => setMagnet(e.target.value)}
          rows={3}
          placeholder="magnet:?xt=urn:btih:..."
        />
        <div className="flex flex-wrap gap-2">
          <Button type="submit" disabled={busy || !magnet.trim()}>
            Add magnet
          </Button>
          <Button type="button" variant="secondary" onClick={() => void startFile()} disabled={busy}>
            Open .torrent
          </Button>
          <Button type="button" variant="ghost" onClick={() => void openFolder()}>
            <IconFolder className="h-4 w-4" />
            Open folder
          </Button>
        </div>
      </form>

      {loadError ? (
        <Alert variant="destructive">
          <AlertDescription>{loadError}</AlertDescription>
          <AlertAction>
            <Button type="button" variant="secondary" onClick={() => void refreshHistory()}>
              Try again
            </Button>
          </AlertAction>
        </Alert>
      ) : jobs.length > 0 ? (
        <div className="flex flex-col gap-3">
          {jobs.map((item) => {
            const isDownloading = item.status === 'DOWNLOADING'
            const isPaused = item.status === 'PAUSED'
            const isSeeding = item.status === 'SEEDING'
            const isQueued = item.status === 'QUEUED'
            const isLive = Boolean(item.live)
            const confirmingDelete = pendingDeleteId === item.id

            let actions = (
              <div className="flex flex-wrap gap-2">
                <Button type="button" variant="destructive" onClick={() => setPendingDeleteId(item.id)} disabled={busy}>
                  Delete
                </Button>
              </div>
            )
            if (isQueued) {
              actions = (
                <div className="flex flex-wrap gap-2">
                  <Button type="button" variant="destructive" onClick={() => void cancel(item.id)} disabled={busy}>
                    Cancel
                  </Button>
                </div>
              )
            } else if (isLive && isPaused) {
              actions = (
                <div className="flex flex-wrap gap-2">
                  <Button type="button" onClick={() => void resume(item.id)} disabled={busy}>
                    Resume
                  </Button>
                  <Button type="button" variant="destructive" onClick={() => void cancel(item.id)} disabled={busy}>
                    Cancel
                  </Button>
                </div>
              )
            } else if (isLive) {
              actions = (
                <div className="flex flex-wrap gap-2">
                  <Button type="button" variant="secondary" onClick={() => void pause(item.id)} disabled={busy}>
                    Pause
                  </Button>
                  {isSeeding && (
                    <Button type="button" onClick={() => void finish(item.id)} disabled={busy}>
                      Finish
                    </Button>
                  )}
                  <Button type="button" variant="destructive" onClick={() => void cancel(item.id)} disabled={busy}>
                    Cancel
                  </Button>
                </div>
              )
            } else if (isSeeding) {
              actions = (
                <div className="flex flex-wrap gap-2">
                  <Button type="button" onClick={() => void resumeSeeding(item.id)} disabled={busy}>
                    Resume
                  </Button>
                  <Button type="button" variant="secondary" onClick={() => void finish(item.id)} disabled={busy}>
                    Finish
                  </Button>
                </div>
              )
            } else if (confirmingDelete) {
              actions = (
                <div className="flex flex-wrap gap-2">
                  <Button type="button" variant="secondary" onClick={() => void remove(item.id, false)} disabled={busy}>
                    From list
                  </Button>
                  <Button type="button" variant="destructive" onClick={() => void remove(item.id, true)} disabled={busy}>
                    List + files
                  </Button>
                  <Button type="button" variant="ghost" onClick={() => setPendingDeleteId(null)} disabled={busy}>
                    Back
                  </Button>
                </div>
              )
            }

            return (
              <Card key={item.id}>
                <div className="flex items-center justify-between gap-3">
                  <div className="min-w-0">
                    <p className="font-medium">{item.name || 'Torrent'}</p>
                    <p className={`tabular-nums text-sm ${item.percent >= 100 ? 'text-foreground' : 'text-muted-foreground'}`}>
                      {item.percent >= 100 ? 'Completed' : item.status} · {formatBytes(item.bytesCompleted)} / {formatBytes(item.bytesTotal)}
                      {isDownloading && ` · ${formatSpeed(item.speedBytesPerSecond)}`}
                      {isSeeding && ` · ${formatSpeed(item.uploadSpeedBytesPerSecond)} upload · ${Math.round(item.uploadRatio * 100)}% uploaded`}
                    </p>
                  </div>
                  {actions}
                </div>
                {!isQueued && (
                  <Progress className="mt-3" value={Math.min(100, Math.max(0, item.percent))} />
                )}
                {item.error && <p className="mt-2 text-sm text-destructive">{item.error}</p>}
              </Card>
            )
          })}
        </div>
      ) : (
        <p className="text-sm text-muted-foreground">No torrent jobs yet. Add a magnet or open a .torrent file to start.</p>
      )}
    </section>
  )
}
