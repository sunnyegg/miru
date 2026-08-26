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
        className="flex flex-col gap-3 rounded-xl bg-card p-4"
        onSubmit={(e) => {
          e.preventDefault()
          void startMagnet()
        }}
      >
        <label htmlFor="magnet" className="text-sm font-medium">Magnet link</label>
        <textarea
          id="magnet"
          value={magnet}
          onChange={(e) => setMagnet(e.target.value)}
          rows={3}
          className="rounded-lg border border-border/40 bg-muted px-3 py-2 text-sm text-foreground focus-visible:outline-2 focus-visible:outline-ring"
          placeholder="magnet:?xt=urn:btih:..."
        />
        <div className="flex flex-wrap gap-2">
          <button
            type="submit"
            disabled={busy || !magnet.trim() || Boolean(activeJob)}
            className="min-h-11 cursor-pointer rounded-lg bg-accent px-4 text-sm font-medium text-on-accent disabled:cursor-not-allowed disabled:opacity-50"
          >
            Add magnet
          </button>
          <button
            type="button"
            onClick={() => void startFile()}
            disabled={busy || Boolean(activeJob)}
            className="min-h-11 cursor-pointer rounded-lg bg-secondary px-4 text-sm text-on-secondary disabled:cursor-not-allowed disabled:opacity-50"
          >
            Open .torrent
          </button>
          <button
            type="button"
            onClick={() => void OpenDownloadFolder()}
            className="inline-flex min-h-11 cursor-pointer items-center gap-2 rounded-lg px-4 text-sm text-muted-foreground hover:text-foreground"
          >
            <IconFolder className="h-4 w-4" />
            Open folder
          </button>
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
              <div key={item.id} className="rounded-xl bg-card p-4">
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
                        <button
                          type="button"
                          onClick={() => void resume()}
                          disabled={busy}
                          className="min-h-11 cursor-pointer rounded-lg bg-accent px-4 text-sm font-medium text-on-accent disabled:opacity-50"
                        >
                          Resume
                        </button>
                      ) : (
                        <button
                          type="button"
                          onClick={() => void pause()}
                          disabled={busy}
                          className="min-h-11 cursor-pointer rounded-lg bg-secondary px-4 text-sm text-on-secondary disabled:opacity-50"
                        >
                          Pause
                        </button>
                      )}
                      <button
                        type="button"
                        onClick={() => void cancel()}
                        disabled={busy}
                        className="min-h-11 cursor-pointer rounded-lg bg-destructive px-4 text-sm text-on-destructive disabled:opacity-50"
                      >
                        Cancel
                      </button>
                    </div>
                  )}
                </div>
                <div
                  className="mt-3 h-2 overflow-hidden rounded-full bg-muted"
                  role="progressbar"
                  aria-valuemin={0}
                  aria-valuemax={100}
                  aria-valuenow={Math.round(item.percent)}
                >
                  <div className="h-full bg-accent" style={{width: `${Math.min(100, Math.max(0, item.percent))}%`}} />
                </div>
                {item.error && <p className="mt-2 text-sm text-destructive">{item.error}</p>}
              </div>
            )
          })}
        </div>
      ) : (
        <p className="text-sm text-muted-foreground">No torrent jobs yet.</p>
      )}
    </section>
  )
}
