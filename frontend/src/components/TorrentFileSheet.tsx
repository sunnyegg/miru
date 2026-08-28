import {useEffect, useState} from 'react'
import {formatBytes} from '../lib/format'
import type {TorrentFileView} from '../lib/types'
import {Button} from '@/components/ui/button'
import {Dialog} from '@/components/ui/dialog'
import {Skeleton} from '@/components/ui/skeleton'

type Props = {
  name: string
  bytesTotal: number
  files: TorrentFileView[]
  loading: boolean
  error: string
  confirming: boolean
  onClose: () => void
  onConfirm: (files: TorrentFileView[]) => void
}

export function TorrentFileSheet({
  name,
  bytesTotal,
  files,
  loading,
  error,
  confirming,
  onClose,
  onConfirm,
}: Props) {
  const [selectedPaths, setSelectedPaths] = useState<string[]>([])

  useEffect(() => {
    setSelectedPaths(files.map((file) => file.path))
  }, [files])

  const selectedFiles = files.filter((file) => selectedPaths.includes(file.path))
  const selectedBytes = selectedFiles.reduce((sum, file) => sum + file.length, 0)
  const hasVideos = files.some((file) => file.isVideo)
  const canDownload = selectedFiles.length > 0 && !loading && !confirming && !error

  function togglePath(path: string) {
    setSelectedPaths((current) => {
      if (current.includes(path)) {
        return current.filter((item) => item !== path)
      }
      return [...current, path]
    })
  }

  return (
    <Dialog.Root
      open
      onOpenChange={(open) => {
        if (!open) {
          onClose()
        }
      }}
    >
      <Dialog.Portal>
        <Dialog.Backdrop />
        <Dialog.Viewport>
          <Dialog.Panel aria-labelledby="torrent-files-title">
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <Dialog.Title id="torrent-files-title">Choose files</Dialog.Title>
                <p className="mt-1 truncate text-sm text-muted-foreground">
                  {name || 'Torrent'}
                </p>
                <p className="mt-1 text-xs text-muted-foreground">
                  {loading
                    ? 'Reading torrent contents…'
                    : `${selectedFiles.length} of ${files.length} files · ${formatBytes(selectedBytes)} of ${formatBytes(bytesTotal)}`}
                </p>
              </div>
              <Button type="button" variant="ghost" onClick={onClose} disabled={confirming}>
                Cancel
              </Button>
            </div>

            {error ? (
              <p className="mt-4 text-sm text-destructive">{error}</p>
            ) : loading ? (
              <ul className="mt-4 flex flex-col gap-1" aria-busy="true" aria-label="Loading torrent files">
                {Array.from({length: 4}, (_, index) => (
                  <li key={index} className="flex min-h-11 items-center gap-3 bg-muted px-3">
                    <Skeleton className="size-4 shrink-0 animate-pulse" />
                    <Skeleton className="h-4 w-2/3 animate-pulse" />
                  </li>
                ))}
              </ul>
            ) : (
              <>
                <div className="mt-4 flex flex-wrap gap-2">
                  <Button
                    type="button"
                    variant="ghost"
                    onClick={() => setSelectedPaths(files.map((file) => file.path))}
                  >
                    All
                  </Button>
                  <Button type="button" variant="ghost" onClick={() => setSelectedPaths([])}>
                    None
                  </Button>
                  {hasVideos && (
                    <Button
                      type="button"
                      variant="ghost"
                      onClick={() => setSelectedPaths(files.filter((file) => file.isVideo).map((file) => file.path))}
                    >
                      Videos
                    </Button>
                  )}
                </div>
                <ul className="mt-3 max-h-80 overflow-y-auto">
                  {files.map((file) => {
                    const checked = selectedPaths.includes(file.path)
                    return (
                      <li key={file.path}>
                        <label className="flex min-h-11 cursor-pointer items-center gap-3 px-3 hover:bg-muted">
                          <input
                            type="checkbox"
                            className="size-4 shrink-0 accent-primary"
                            checked={checked}
                            onChange={() => togglePath(file.path)}
                          />
                          <span className="min-w-0 flex-1 break-all text-sm">{file.path}</span>
                          <span className="shrink-0 tabular-nums text-xs text-muted-foreground">
                            {formatBytes(file.length)}
                          </span>
                        </label>
                      </li>
                    )
                  })}
                </ul>
              </>
            )}

            <div className="mt-4 flex justify-end gap-2">
              <Button type="button" variant="ghost" onClick={onClose} disabled={confirming}>
                Cancel
              </Button>
              <Button
                type="button"
                onClick={() => onConfirm(selectedFiles)}
                disabled={!canDownload}
                aria-busy={confirming}
              >
                {confirming ? 'Adding…' : 'Download'}
              </Button>
            </div>
          </Dialog.Panel>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
