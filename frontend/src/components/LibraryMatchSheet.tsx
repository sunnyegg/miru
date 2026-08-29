import type {AnimeView, EpisodeView} from '../lib/types'
import {Button} from '@/components/ui/button'
import {Dialog} from '@/components/ui/dialog'
import {Input} from '@/components/ui/input'

export type LibraryMatchPicker = {
  episode: EpisodeView
  candidates: AnimeView[]
  query: string
}

type Props = {
  picker: LibraryMatchPicker
  searching: boolean
  bindingAnimeId: number | null
  onQueryChange: (query: string) => void
  onSearch: () => void
  onBind: (anilistId: number) => void
  onSkip: () => void
}

export function LibraryMatchSheet({
  picker,
  searching,
  bindingAnimeId,
  onQueryChange,
  onSearch,
  onBind,
  onSkip,
}: Props) {
  const englishTitle = picker.episode.animeTitle?.trim() || picker.episode.displayTitle
  const descriptionCopy = `Match the local file "${englishTitle}" to an AniList title. Pick the right one, or skip to leave the file unmatched.`

  return (
    <Dialog.Root
      open
      onOpenChange={(open) => {
        if (!open) {
          onSkip()
        }
      }}
    >
      <Dialog.Portal>
        <Dialog.Backdrop />
        <Dialog.Viewport>
          <Dialog.Panel>
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <Dialog.Title>Match AniList title</Dialog.Title>
                <Dialog.Description>{descriptionCopy}</Dialog.Description>
                <p className="mt-1 wrap-break-word text-sm text-muted-foreground">
                  {picker.episode.displayTitle}
                </p>
              </div>
              <Button type="button" variant="ghost" onClick={onSkip}>
                Skip
              </Button>
            </div>
            <div className="mt-3 flex flex-wrap gap-2">
              <label className="sr-only" htmlFor="anilist-search">Search AniList</label>
              <Input
                id="anilist-search"
                value={picker.query}
                onChange={(event) => onQueryChange(event.target.value)}
                className="min-w-0 flex-1 basis-56 border-border"
              />
              <Button
                type="button"
                variant="secondary"
                onClick={onSearch}
                disabled={searching || !picker.query.trim()}
              >
                {searching ? 'Searching…' : 'Search'}
              </Button>
            </div>
            <ul className="mt-3 flex flex-col gap-1">
              {picker.candidates.map((anime) => {
                const title = anime.titleEnglish || anime.titleRomaji
                const isBinding = bindingAnimeId === anime.id
                return (
                  <li key={anime.id}>
                    <div className="flex min-h-11 items-center gap-3 bg-muted px-3 transition-colors duration-200 hover:bg-secondary motion-reduce:transition-none">
                      {anime.coverImage ? (
                        <img
                          src={anime.coverImage}
                          alt=""
                          width={40}
                          height={56}
                          referrerPolicy="no-referrer"
                          className="shrink-0 object-cover"
                          style={{width: 40, height: 56}}
                        />
                      ) : (
                        <span className="shrink-0 bg-muted" style={{width: 40, height: 56}} />
                      )}
                      <span className="min-w-0 flex-1 py-2">
                        <span className="block truncate text-sm font-medium" title={title}>
                          {title}
                        </span>
                        <span className="block truncate text-xs text-muted-foreground" title={anime.titleRomaji}>
                          {anime.titleRomaji}
                        </span>
                      </span>
                      <Button
                        type="button"
                        onClick={() => onBind(anime.id)}
                        disabled={bindingAnimeId !== null}
                        aria-busy={isBinding}
                        aria-label={`Bind ${title} to ${englishTitle}`}
                        style={{gap: 8, paddingInline: 12}}
                      >
                        {isBinding ? 'Binding…' : 'Bind'}
                      </Button>
                    </div>
                  </li>
                )
              })}
              {picker.candidates.length === 0 && !searching && (
                <li className="text-sm text-muted-foreground">No matches. Search with a cleaner title.</li>
              )}
            </ul>
          </Dialog.Panel>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
