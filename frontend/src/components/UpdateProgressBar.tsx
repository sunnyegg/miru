import {Progress} from '@/components/ui/progress'
import {formatBytes} from '../lib/format'
import type {UpdateProgress} from '../lib/types'

type Props = {
  progress: UpdateProgress
  className?: string
}

export function UpdateProgressBar({progress, className}: Props) {
  const known = progress.total > 0
  const value = known
    ? Math.min(100, Math.round((progress.downloaded / progress.total) * 100))
    : null
  const label =
    progress.phase === 'installing'
      ? 'Installing…'
      : known
        ? `${formatBytes(progress.downloaded)} / ${formatBytes(progress.total)}`
        : `${formatBytes(progress.downloaded)} downloaded`
  return (
    <div className={className ?? 'flex w-full flex-col gap-1'}>
      <div className="flex items-center justify-between gap-2 text-xs text-muted-foreground">
        <span>{label}</span>
        {known && <span className="tabular-nums">{value}%</span>}
      </div>
      <Progress value={value} />
    </div>
  )
}
