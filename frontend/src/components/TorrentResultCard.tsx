import {memo} from 'react'
import type {NyaaResultView} from '../lib/types'
import {Badge} from '@/components/ui/badge'
import {Button} from '@/components/ui/button'
import {Card} from '@/components/ui/card'

export type TorrentDownloadHandlerRef = {
  current: (result: NyaaResultView, resultIndex: number) => void
}

type Props = {
  result: NyaaResultView
  resultIndex: number
  starting: boolean
  busy: boolean
  onDownload: TorrentDownloadHandlerRef
}

const dateFormatter = new Intl.DateTimeFormat(undefined, {
  dateStyle: 'medium',
  timeStyle: 'short',
})

export const TorrentResultCard = memo(function TorrentResultCard({
  result,
  resultIndex,
  starting,
  busy,
  onDownload,
}: Props) {
  const publishedDate = new Date(result.published)
  const publishedLabel = Number.isNaN(publishedDate.getTime())
    ? 'Unknown date'
    : dateFormatter.format(publishedDate)
  const hasPeerCounts =
    result.seeders > 0 || result.leechers > 0 || result.downloads > 0

  return (
    <Card>
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
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
        <Button
          type="button"
          variant="secondary"
          onClick={() => onDownload.current(result, resultIndex)}
          disabled={busy}
        >
          {starting ? 'Adding…' : 'Download'}
        </Button>
      </div>
    </Card>
  )
})
