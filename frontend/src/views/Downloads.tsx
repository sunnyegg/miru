import {useEffect, useState} from 'react'
import {
  CancelDownload,
  DownloadStatus,
  OpenDownloadFolder,
  StartMagnet,
  StartTorrentFile,
} from '../../wailsjs/go/main/App'
import {errorMessage, formatBytes} from '../lib/format'
import type {DownloadView} from '../lib/types'
import {IconFolder} from '../components/Icons'

type Props = {
  notice: (msg: string, isError?: boolean) => void
  job: DownloadView | null
  onJob: (job: DownloadView | null) => void
}

export function DownloadsView({notice, job, onJob}: Props) {
  const [magnet, setMagnet] = useState('')
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    void DownloadStatus()
      .then((current) => onJob(current))
      .catch((err) => notice(errorMessage(err), true))
  }, [])

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
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBusy(false)
    }
  }

  const downloading = job?.status === 'DOWNLOADING'

  return (
    <section className="flex h-full flex-col gap-6">
      <header>
        <h2 className="text-2xl font-semibold">Downloads</h2>
        <p className="mt-1 text-sm text-muted-foreground">One active torrent at a time. Upload stops when the file is complete.</p>
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
            disabled={busy || !magnet.trim() || downloading}
            className="min-h-11 cursor-pointer rounded-lg bg-accent px-4 text-sm font-medium text-on-accent disabled:cursor-not-allowed disabled:opacity-50"
          >
            Add magnet
          </button>
          <button
            type="button"
            onClick={() => void startFile()}
            disabled={busy || downloading}
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

      {job ? (
        <div className="rounded-xl bg-card p-4">
          <div className="flex items-center justify-between gap-3">
            <div>
              <p className="font-medium">{job.name || 'Torrent'}</p>
              <p className="text-sm text-muted-foreground">
                {job.status} · {formatBytes(job.bytesCompleted)} / {formatBytes(job.bytesTotal)}
              </p>
            </div>
            {downloading && (
              <button
                type="button"
                onClick={() => void cancel()}
                disabled={busy}
                className="min-h-11 cursor-pointer rounded-lg bg-destructive px-4 text-sm text-on-destructive disabled:opacity-50"
              >
                Cancel
              </button>
            )}
          </div>
          <div
            className="mt-3 h-2 overflow-hidden rounded-full bg-muted"
            role="progressbar"
            aria-valuemin={0}
            aria-valuemax={100}
            aria-valuenow={Math.round(job.percent)}
          >
            <div className="h-full bg-accent" style={{width: `${Math.min(100, Math.max(0, job.percent))}%`}} />
          </div>
          {job.error && <p className="mt-2 text-sm text-destructive">{job.error}</p>}
        </div>
      ) : (
        <p className="text-sm text-muted-foreground">No torrent jobs yet.</p>
      )}
    </section>
  )
}
