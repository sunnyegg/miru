import {memo} from 'react'
import type {MouseEvent} from 'react'
import type {NyaaResultView} from '../lib/types'
import {Badge} from '@/components/ui/badge'
import {Button} from '@/components/ui/button'
import {Card} from '@/components/ui/card'
import {cn} from '@/lib/utils'

export type TorrentCardHandlers = {
  download: (result: NyaaResultView) => void
  toggleSelected: (result: NyaaResultView) => void
}

export type TorrentCardHandlersRef = {
  current: TorrentCardHandlers
}

type Props = {
  result: NyaaResultView
  starting: boolean
  busy: boolean
  selected: boolean
  selectMode: boolean
  handlers: TorrentCardHandlersRef
}

const dateFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: 'medium',
  timeStyle: 'short',
})

const selectionLabelMax = 80

export const TorrentResultCard = memo(function TorrentResultCard({
  result,
  starting,
  busy,
  selected,
  selectMode,
  handlers,
}: Props) {
  const publishedDate = new Date(result.published)
  const publishedLabel = Number.isNaN(publishedDate.getTime())
    ? 'Unknown date'
    : dateFormatter.format(publishedDate)
  const hasPeerCounts =
    result.seeders > 0 || result.leechers > 0 || result.downloads > 0
  const selectionLabel =
    result.title.length > selectionLabelMax
      ? `${result.title.slice(0, selectionLabelMax)}…`
      : result.title

  // The card is a pointer shortcut for the checkbox, so it only toggles in
  // selection mode and when the click is not really aimed at a control or text.
  function handleCardClick(event: MouseEvent<HTMLDivElement>) {
    if (!selectMode || busy) {
      return
    }
    if (window.getSelection()?.toString()) {
      return
    }
    const target = event.target
    if (
      target instanceof Element &&
      target.closest('button, input, a, label')
    ) {
      return
    }
    handlers.current.toggleSelected(result)
  }

  return (
    <Card
      className={cn(
        'border border-transparent transition-colors',
        selected && 'border-accent border-l-2',
        selectMode && !busy ? 'cursor-pointer' : 'cursor-default',
      )}
      onClick={handleCardClick}
    >
      <div className="flex gap-3">
        <div className="flex min-w-0 flex-1 flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div className="min-w-0">
            <p className="wrap-break-word font-medium">{result.title}</p>
            <p className="mt-1 text-xs text-muted-foreground">
              {publishedLabel} · {result.size || 'Unknown size'}
            </p>
            {hasPeerCounts && (
              <p className="mt-1 text-xs text-muted-foreground">
                {result.seeders} seeders · {result.leechers} leechers ·{' '}
                {result.downloads} downloads
              </p>
            )}
            {(result.trusted || result.remake) && (
              <p className="mt-2">
                {result.trusted && <Badge className="mr-2">Trusted</Badge>}
                {result.remake && <Badge>Remake</Badge>}
              </p>
            )}
          </div>
          {!selectMode && (
            <Button
              type="button"
              variant="secondary"
              onClick={() => handlers.current.download(result)}
              disabled={busy}
            >
              {starting ? 'Adding…' : 'Download'}
            </Button>
          )}
        </div>
        {selectMode && (
          <label className="flex min-h-11 min-w-11 shrink-0 cursor-pointer items-start justify-center pt-0.5 has-[input:disabled]:cursor-not-allowed">
            <input
              type="checkbox"
              className="size-4 cursor-pointer accent-primary disabled:cursor-not-allowed disabled:opacity-50"
              checked={selected}
              disabled={busy}
              aria-label={`Select ${selectionLabel}`}
              onChange={() => handlers.current.toggleSelected(result)}
            />
          </label>
        )}
      </div>
    </Card>
  )
})
