import {useEffect, useState} from 'react'
import {
  CancelDownload,
  DownloadHistory,
  OpenDownloadFolder,
  PauseDownload,
  ResumeDownload,
  StartMagnet,
  StartTorrentFile,
} from '../../wailsjs/go/main/App'
import {errorMessage, formatBytes, formatSpeed} from '../lib/format'
import type {DownloadView} from '../lib/types'
import {IconFolder} from '../components/Icons'
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

  useEffect(() => {
    void DownloadHistory()
      .then((current) => onJobs(current ?? []))
      .catch((err) => notice(errorMessage(err), true))
  }, [])

  async function refreshHistory() {
    try {
      onJobs((await DownloadHistory()) ?? [])
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  async function startMagnet() {
    setBusy(true)
    try {
      await StartMagnet(magnet.trim())
      setMagnet('')
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
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBusy(false)
    }
  }

  async function cancel() {
    setBusy(true)
    try {
      await CancelDownload()
      await refreshHistory()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBusy(false)
    }
  }

  async function pause() {
    setBusy(true)
    try {
      await PauseDownload()
      await refreshHistory()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBusy(false)
    }
  }

  async function resume() {
    setBusy(true)
    try {
      await ResumeDownload()
      await refreshHistory()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBusy(false)
    }
  }

  const activeJob = jobs.find((item) =>
    item.status === 'DOWNLOADING' ||
    item.status === 'PAUSED' ||
    item.status === 'SEEDING',
  )

  return (
    <section className="flex h-full flex-col gap-6">
      <header>
        <h2 className="text-2xl font-semibold">Downloads</h2>
        <p className="mt-1 text-sm text-muted-foreground">One active torrent at a time. Completed files seed until 0.5x upload ratio.</p>
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
          <Button type="submit" disabled={busy || !magnet.trim() || Boolean(activeJob)}>
            Add magnet
          </Button>
          <Button type="button" variant="secondary" onClick={() => void startFile()} disabled={busy || Boolean(activeJob)}>
            Open .torrent
          </Button>
          <Button type="button" variant="ghost" onClick={() => void OpenDownloadFolder()}>
            <IconFolder className="h-4 w-4" />
            Open folder
          </Button>
        </div>
      </form>

      {jobs.length > 0 ? (
        <div className="flex flex-col gap-3">
          {jobs.map((item) => {
            const isDownloading = item.status === 'DOWNLOADING'
            const isPaused = item.status === 'PAUSED'
            const isSeeding = item.status === 'SEEDING'
            const isActive = isDownloading || isPaused || isSeeding
            return (
              <Card key={item.id}>
                <div className="flex items-center justify-between gap-3">
                  <div>
                    <p className="font-medium">{item.name || 'Torrent'}</p>
                    <p className="text-sm text-muted-foreground">
                      {item.status} · {formatBytes(item.bytesCompleted)} / {formatBytes(item.bytesTotal)}
                      {isDownloading && ` · ${formatSpeed(item.speedBytesPerSecond)}`}
                      {isSeeding && ` · ${formatSpeed(item.uploadSpeedBytesPerSecond)} upload · ${Math.round(item.uploadRatio * 100)}% uploaded`}
                    </p>
                  </div>
                  {isActive && (
                    <div className="flex flex-wrap gap-2">
                      {isPaused ? (
                        <Button type="button" onClick={() => void resume()} disabled={busy}>
                          Resume
                        </Button>
                      ) : (
                        <Button type="button" variant="secondary" onClick={() => void pause()} disabled={busy}>
                          Pause
                        </Button>
                      )}
                      <Button type="button" variant="destructive" onClick={() => void cancel()} disabled={busy}>
                        Cancel
                      </Button>
                    </div>
                  )}
                </div>
                <Progress className="mt-3" value={Math.min(100, Math.max(0, item.percent))} />
                {item.error && <p className="mt-2 text-sm text-destructive">{item.error}</p>}
              </Card>
            )
          })}
        </div>
      ) : (
        <p className="text-sm text-muted-foreground">No torrent jobs yet.</p>
      )}
    </section>
  )
}
