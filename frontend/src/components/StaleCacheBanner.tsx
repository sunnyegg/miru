import {formatAge} from '../lib/format'
import type {AnilistCacheStale} from '../lib/types'
import {Alert, AlertDescription} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'

type Props = {
  entries: AnilistCacheStale[]
  retrying: boolean
  onRetry: () => void
}

const sourceLabels: Record<string, string> = {
  'search results': 'search results',
  'anime details': 'anime details',
  'your lists': 'your lists',
  'airing schedule': 'the airing schedule',
  'episode progress': 'episode progress',
  'episode thumbnails': 'episode thumbnails',
}

export function StaleCacheBanner({entries, retrying, onRetry}: Props) {
  if (entries.length === 0) {
    return null
  }
  const oldest = entries.reduce((oldestEntry, entry) =>
    entry.ageSeconds > oldestEntry.ageSeconds ? entry : oldestEntry,
  )
  const labels = new Set(
    entries.map((entry) => sourceLabels[entry.source] ?? 'saved data'),
  )
  const subject = labels.size === 1 ? [...labels][0] : 'saved data'

  return (
    <Alert
      role="status"
      className="border-0 border-b border-border bg-card px-4 py-2 text-sm text-muted-foreground"
    >
      <div className="flex flex-wrap items-center gap-2">
        <AlertDescription className="min-w-0 flex-1">
          AniList is unreachable. Showing {subject} from{' '}
          {formatAge(oldest.ageSeconds)}.
        </AlertDescription>
        <Button
          type="button"
          variant="secondary"
          disabled={retrying}
          onClick={onRetry}
        >
          {retrying ? 'Retrying…' : 'Retry'}
        </Button>
      </div>
    </Alert>
  )
}
