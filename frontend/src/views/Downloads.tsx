import {useEffect, useState} from 'react'
import {
  CancelDownload,
  FinishDownload,
  InspectTorrent,
  OpenDownloadFolder,
  PauseDownload,
  PickTorrentFile,
  RemoveDownload,
  ResumeDownload,
  ResumeSeeding,
  StartTorrent,
} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import {groupDownloads, type DownloadGroup} from '../lib/downloadGroups'
import type {TorrentContentsView, TorrentFileView} from '../lib/types'
import {useDownloadStore} from '../stores/downloadStore'
import {AddTorrentDialog} from '../components/AddTorrentDialog'
import {DeleteDownloadDialog} from '../components/DeleteDownloadDialog'
import {DownloadJobCard} from '../components/DownloadJobCard'
import {TorrentFileSheet} from '../components/TorrentFileSheet'
import {IconFolder} from '../components/Icons'
import {Alert, AlertAction, AlertDescription} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'
import {cn} from '@/lib/utils'

type Props = {
  notice: (msg: string, isError?: boolean) => void
}

type PickerState = {
  source: string
  contents: TorrentContentsView
  loading: boolean
  error: string
}

const downloadTabs: {group: DownloadGroup; label: string}[] = [
  {group: 'downloading', label: 'Downloading'},
  {group: 'seeding', label: 'Seeding'},
  {group: 'completed', label: 'Completed'},
]

export function DownloadsView({notice}: Props) {
  const jobs = useDownloadStore((state) => state.jobs)
  const activeTab = useDownloadStore((state) => state.activeTab)
  const setActiveTab = useDownloadStore((state) => state.setActiveTab)
  const loadHistory = useDownloadStore((state) => state.loadHistory)

  const [addDialogOpen, setAddDialogOpen] = useState(false)
  const [busy, setBusy] = useState(false)
  const [pendingDeleteId, setPendingDeleteId] = useState<number | null>(null)
  const [deleteFilesConfirmId, setDeleteFilesConfirmId] = useState<
    number | null
  >(null)
  const [loadError, setLoadError] = useState('')
  const [picker, setPicker] = useState<PickerState | null>(null)
  const [confirming, setConfirming] = useState(false)

  const grouped = groupDownloads(jobs)
  const hasJobs = jobs.length > 0
  const tabJobs = grouped[activeTab]
  const deleteFilesConfirmJob =
    deleteFilesConfirmId === null
      ? null
      : (jobs.find((job) => job.id === deleteFilesConfirmId) ?? null)

  useEffect(() => {
    void refreshHistory()
    // Mount-only load; refreshHistory is recreated each render but does not depend on it.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function refreshHistory() {
    setLoadError('')
    try {
      await loadHistory()
    } catch (err) {
      setLoadError(errorMessage(err))
    }
  }

  async function openPicker(source: string) {
    setPicker({
      source,
      contents: {name: '', bytesTotal: 0, files: []},
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
          contents: contents ?? {name: '', bytesTotal: 0, files: []},
          loading: false,
          error: '',
        }
      })
    } catch (err) {
      const message = errorMessage(err)
      setPicker((current) => {
        if (!current || current.source !== source) {
          return current
        }
        return {...current, loading: false, error: message}
      })
    }
  }

  async function startMagnet(source: string) {
    const trimmed = source.trim()
    if (!trimmed) {
      return
    }
    setBusy(true)
    try {
      await openPicker(trimmed)
      setAddDialogOpen(false)
    } finally {
      setBusy(false)
    }
  }

  async function startFile() {
    setBusy(true)
    try {
      const path = await PickTorrentFile()
      if (!path) {
        return
      }
      await openPicker(path)
      setAddDialogOpen(false)
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBusy(false)
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
      await refreshHistory()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setConfirming(false)
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
      setDeleteFilesConfirmId(null)
      await refreshHistory()
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setBusy(false)
    }
  }

  function confirmDeleteFiles(id: number) {
    setDeleteFilesConfirmId(id)
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
      <header className="flex flex-col gap-4">
        <div>
          <h2 className="text-2xl font-semibold">Downloads</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Choose which files to keep before a torrent starts. Extra torrents
            wait in queue.
          </p>
        </div>
        <div className="flex flex-wrap justify-end gap-2">
          <Button type="button" onClick={() => setAddDialogOpen(true)}>
            Add torrent
          </Button>
          <Button
            type="button"
            variant="secondary"
            onClick={() => void openFolder()}
          >
            <IconFolder className="h-4 w-4" />
            Open folder
          </Button>
        </div>
        {hasJobs && (
          <div
            className="flex flex-wrap gap-2"
            role="group"
            aria-label="Download categories"
          >
            {downloadTabs.map(({group, label}) => {
              const active = activeTab === group
              return (
                <Button
                  key={group}
                  type="button"
                  aria-pressed={active}
                  variant="ghost"
                  onClick={() => setActiveTab(group)}
                  className={cn(
                    'border-b-2 px-4 motion-reduce:transition-none',
                    active
                      ? 'border-accent bg-muted text-foreground hover:text-foreground'
                      : 'border-transparent text-muted-foreground hover:bg-muted hover:text-foreground',
                  )}
                >
                  {label} · {grouped[group].length}
                </Button>
              )
            })}
          </div>
        )}
      </header>

      {loadError ? (
        <Alert variant="destructive">
          <AlertDescription>{loadError}</AlertDescription>
          <AlertAction>
            <Button
              type="button"
              variant="secondary"
              onClick={() => void refreshHistory()}
            >
              Try again
            </Button>
          </AlertAction>
        </Alert>
      ) : hasJobs ? (
        tabJobs.length > 0 ? (
          <div className="flex flex-col gap-3">
            {tabJobs.map((item) => (
              <DownloadJobCard
                key={item.id}
                item={item}
                group={activeTab}
                busy={busy}
                confirmingDelete={pendingDeleteId === item.id}
                onPendingDelete={setPendingDeleteId}
                onClearPendingDelete={() => setPendingDeleteId(null)}
                onCancel={(id) => void cancel(id)}
                onPause={(id) => void pause(id)}
                onResume={(id) => void resume(id)}
                onRemove={(id, deleteFiles) => void remove(id, deleteFiles)}
                onConfirmDeleteFiles={confirmDeleteFiles}
                onResumeSeeding={(id) => void resumeSeeding(id)}
                onFinish={(id) => void finish(id)}
              />
            ))}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">
            No torrents in this category.
          </p>
        )
      ) : (
        <p className="text-sm text-muted-foreground">
          No torrent jobs yet. Add a torrent to start.
        </p>
      )}

      {deleteFilesConfirmJob && (
        <DeleteDownloadDialog
          open={deleteFilesConfirmId !== null}
          torrentName={deleteFilesConfirmJob.name || 'Torrent'}
          busy={busy}
          onOpenChange={(open) => {
            if (!open) {
              setDeleteFilesConfirmId(null)
            }
          }}
          onConfirm={() => void remove(deleteFilesConfirmJob.id, true)}
        />
      )}

      {addDialogOpen && (
        <AddTorrentDialog
          open={addDialogOpen}
          onOpenChange={setAddDialogOpen}
          busy={busy}
          onStartMagnet={startMagnet}
          onStartFile={startFile}
        />
      )}

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
    </section>
  )
}
