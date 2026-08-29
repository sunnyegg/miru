import {cn} from '@/lib/utils'

type Props = {
  title: string
  coverImage: string
  caption: string
  subcaption?: string | null
  accentCaption?: boolean
  active?: boolean
  size?: 'shelf' | 'grid'
  onClick: () => void
}

export function LibraryPosterCard({
  title,
  coverImage,
  caption,
  subcaption,
  accentCaption = false,
  active = false,
  size = 'grid',
  onClick,
}: Props) {
  const widthClass = size === 'shelf' ? 'w-44 shrink-0 sm:w-48' : 'w-full'

  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={cn(
        'group relative cursor-pointer overflow-hidden text-left transition-transform duration-200 motion-reduce:transition-none',
        'hover:scale-[1.02] motion-reduce:hover:scale-100',
        active && 'ring-2 ring-accent ring-offset-2 ring-offset-background',
        widthClass,
      )}
    >
      <div className="relative aspect-[2/3] w-full bg-muted">
        {coverImage ? (
          <img
            src={coverImage}
            alt=""
            referrerPolicy="no-referrer"
            className="size-full object-cover"
          />
        ) : (
          <span className="flex size-full items-end p-3 text-xs text-muted-foreground">
            {title}
          </span>
        )}
        <div
          className="absolute inset-x-0 bottom-0 min-h-[45%] bg-gradient-to-t from-background via-background/95 to-transparent"
          aria-hidden="true"
        />
        <div className="absolute inset-x-0 bottom-0 px-3 pb-3">
          <span className="block truncate text-base font-semibold text-foreground [text-shadow:0_1px_3px_rgba(0,0,0,0.85)]">
            {title}
          </span>
          <span
            className={cn(
              'mt-0.5 block truncate text-xs [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]',
              accentCaption ? 'text-accent' : 'text-foreground/80',
            )}
          >
            {caption}
          </span>
          {subcaption && (
            <span className="mt-0.5 block truncate text-xs text-foreground/70 opacity-0 transition-opacity duration-200 [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] group-hover:opacity-100 motion-reduce:opacity-100 motion-reduce:transition-none">
              {subcaption}
            </span>
          )}
        </div>
      </div>
    </button>
  )
}
