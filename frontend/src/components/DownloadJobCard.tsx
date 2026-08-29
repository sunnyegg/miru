import {formatBytes, formatSpeed} from '../lib/format'
import type {DownloadGroup} from '../lib/downloadGroups'
import type {DownloadView} from '../lib/types'
import {Badge} from '@/components/ui/badge'
import {Button} from '@/components/ui/button'
import {Card} from '@/components/ui/card'
import {Progress} from '@/components/ui/progress'

type Props = {
  item: DownloadView
  group: DownloadGroup
  busy: boolean
  confirmingDelete: boolean
  onPendingDelete: (id: number) => void
  onClearPendingDelete: () => void
  onCancel: (id: number) => void
  onPause: (id: number) => void
  onResume: (id: number) => void
  onRemove: (id: number, deleteFiles: boolean) => void
  onConfirmDeleteFiles: (id: number) => void
  onResumeSeeding: (id: number) => void
  onFinish: (id: number) => void
}

function torrentStatusLabel(status: string): string {
  switch (status) {
    case 'DOWNLOADING':
      return 'Downloading'
    case 'PAUSED':
      return 'Paused'
    case 'SEEDING':
      return 'Seeding'
    case 'QUEUED':
      return 'Queued'
    case 'FAILED':
      return 'Failed'
    case 'COMPLETED':
      return 'Completed'
    case 'CANCELLED':
      return 'Cancelled'
    default:
      return status
  }
}

function torrentErrorLabel(error: string): string {
  if (
    /no such host|connection refused|i\/o timeout|network is unreachable|proxy/i.test(
      error,
    )
  ) {
    return `Network problem: ${error}`
  }
  return error
}

function completedStatusBadge(status: string) {
  if (status === 'FAILED') {
    return <Badge variant="destructive">Failed</Badge>
  }
  if (status === 'CANCELLED') {
    return <Badge variant="secondary">Cancelled</Badge>
  }
  if (status === 'COMPLETED') {
    return <Badge variant="secondary">Completed</Badge>
  }
  return null
}

export function DownloadJobCard({
  item,
  group,
  busy,
  confirmingDelete,
  onPendingDelete,
  onClearPendingDelete,
  onCancel,
  onPause,
  onResume,
  onRemove,
  onConfirmDeleteFiles,
  onResumeSeeding,
  onFinish,
}: Props) {
  const isDownloading = item.status === 'DOWNLOADING'
  const isPaused = item.status === 'PAUSED'
  const isSeeding = item.status === 'SEEDING'
  const isQueued = item.status === 'QUEUED'
  const isLive = Boolean(item.live)
  const files = item.files ?? []

  let actions = (
    <div className="flex flex-wrap gap-2">
      <Button
        type="button"
        variant="destructive"
        onClick={() => onPendingDelete(item.id)}
        disabled={busy}
      >
        Delete
      </Button>
    </div>
  )

  if (isQueued) {
    actions = (
      <div className="flex flex-wrap gap-2">
        <Button
          type="button"
          variant="destructive"
          onClick={() => onCancel(item.id)}
          disabled={busy}
        >
          Cancel
        </Button>
      </div>
    )
  } else if (isLive && isPaused) {
    actions = (
      <div className="flex flex-wrap gap-2">
        <Button type="button" onClick={() => onResume(item.id)} disabled={busy}>
          Resume
        </Button>
        <Button
          type="button"
          variant="destructive"
          onClick={() => onCancel(item.id)}
          disabled={busy}
        >
          Cancel
        </Button>
      </div>
    )
  } else if (isLive) {
    actions = (
      <div className="flex flex-wrap gap-2">
        <Button
          type="button"
          variant="secondary"
          onClick={() => onPause(item.id)}
          disabled={busy}
        >
          Pause
        </Button>
        {isSeeding && (
          <Button
            type="button"
            onClick={() => onFinish(item.id)}
            disabled={busy}
          >
            Stop seeding
          </Button>
        )}
        <Button
          type="button"
          variant="destructive"
          onClick={() => onCancel(item.id)}
          disabled={busy}
        >
          Cancel
        </Button>
      </div>
    )
  } else if (isSeeding) {
    actions = (
      <div className="flex flex-wrap gap-2">
        <Button
          type="button"
          onClick={() => onResumeSeeding(item.id)}
          disabled={busy}
        >
          Resume
        </Button>
        <Button
          type="button"
          variant="secondary"
          onClick={() => onFinish(item.id)}
          disabled={busy}
        >
          Stop seeding
        </Button>
      </div>
    )
  } else if (confirmingDelete) {
    actions = (
      <div className="flex flex-wrap gap-2">
        <Button
          type="button"
          variant="secondary"
          onClick={() => onRemove(item.id, false)}
          disabled={busy}
        >
          Remove from list
        </Button>
        <Button
          type="button"
          variant="destructive"
          onClick={() => onConfirmDeleteFiles(item.id)}
          disabled={busy}
        >
          Remove and delete files
        </Button>
        <Button
          type="button"
          variant="ghost"
          onClick={onClearPendingDelete}
          disabled={busy}
        >
          Back
        </Button>
      </div>
    )
  }

  const statusLine =
    group === 'completed'
      ? `${formatBytes(item.bytesCompleted)} / ${formatBytes(item.bytesTotal)}`
      : [
          item.percent >= 100 ? 'Completed' : torrentStatusLabel(item.status),
          `${formatBytes(item.bytesCompleted)} / ${formatBytes(item.bytesTotal)}`,
          isDownloading ? formatSpeed(item.speedBytesPerSecond) : '',
          isSeeding
            ? `${formatSpeed(item.uploadSpeedBytesPerSecond)} upload · ${Math.round(item.uploadRatio * 100)}% uploaded`
            : '',
        ]
          .filter(Boolean)
          .join(' · ')

  return (
    <Card>
      <div className="flex items-center justify-between gap-3">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <p className="font-medium">{item.name || 'Torrent'}</p>
            {group === 'completed' && completedStatusBadge(item.status)}
          </div>
          <p
            className={`tabular-nums text-sm ${item.percent >= 100 ? 'text-foreground' : 'text-muted-foreground'}`}
          >
            {statusLine}
          </p>
        </div>
        {actions}
      </div>
      {!isQueued && (
        <Progress
          className="mt-3"
          value={Math.min(100, Math.max(0, item.percent))}
        />
      )}
      {files.length > 0 && (
        <ul className="mt-3 max-h-40 overflow-y-auto border-t border-border/40 pt-3">
          {files.map((file) => {
            const filePercent =
              file.length > 0
                ? Math.min(100, (file.bytesCompleted / file.length) * 100)
                : 0
            return (
              <li
                key={file.path}
                className="flex min-h-11 items-center justify-between gap-3 px-1 text-sm"
              >
                <span className="min-w-0 break-all">{file.path}</span>
                <span className="shrink-0 tabular-nums text-xs text-muted-foreground">
                  {file.bytesCompleted > 0
                    ? `${formatBytes(file.bytesCompleted)} / ${formatBytes(file.length)}`
                    : formatBytes(file.length)}
                  {isLive && file.length > 0
                    ? ` · ${Math.round(filePercent)}%`
                    : ''}
                </span>
              </li>
            )
          })}
        </ul>
      )}
      {item.error && (
        <p className="mt-2 text-sm text-destructive">
          {torrentErrorLabel(item.error)}
        </p>
      )}
    </Card>
  )
}
