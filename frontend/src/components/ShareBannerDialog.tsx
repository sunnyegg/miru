import {useEffect, useState} from 'react'
import {IconChevronDown, IconChevronUp, IconHelp} from './Icons'
import {
  SHARE_BANNER_MAX_POSTERS,
  buildShareBannerPng,
  shareBannerSize,
  uint8ToBase64,
  type ShareBannerLayout,
  type ShareBannerOrientation,
  type ShareBannerPoster,
} from '../lib/shareBanner'
import {errorMessage} from '../lib/format'
import {Button} from '@/components/ui/button'
import {Dialog} from '@/components/ui/dialog'
import {Label} from '@/components/ui/label'
import {NativeSelect, NativeSelectOption} from '@/components/ui/native-select'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/components/ui/tooltip'
import {cn} from '@/lib/utils'

export type ShareBannerSession = {
  username: string
  categoryLabel: string
  categoryCount: number
  defaultFilename: string
  posters: ShareBannerPoster[]
}

type Props = {
  open: boolean
  session: ShareBannerSession | null
  saving: boolean
  onClose: () => void
  onSave: (pngBase64: string, defaultFilename: string) => void
  onError: (message: string) => void
}

function defaultPosterCount(total: number): number {
  if (total <= 0) {
    return 0
  }
  return Math.min(6, total, SHARE_BANNER_MAX_POSTERS)
}

function posterCountOptions(total: number): number[] {
  const capped = Math.min(total, SHARE_BANNER_MAX_POSTERS)
  const options = [4, 6, 8, 10, 12].filter((count) => count <= capped)
  if (capped > 0 && !options.includes(capped)) {
    options.push(capped)
  }
  if (options.length === 0 && capped > 0) {
    options.push(capped)
  }
  return options
}

function movePoster(
  posters: ShareBannerPoster[],
  index: number,
  direction: -1 | 1,
): ShareBannerPoster[] {
  const targetIndex = index + direction
  if (targetIndex < 0 || targetIndex >= posters.length) {
    return posters
  }
  const next = posters.slice()
  const [moved] = next.splice(index, 1)
  next.splice(targetIndex, 0, moved)
  return next
}

export function ShareBannerDialog({
  open,
  session,
  saving,
  onClose,
  onSave,
  onError,
}: Props) {
  const [orderedPosters, setOrderedPosters] = useState<ShareBannerPoster[]>([])
  const [posterCount, setPosterCount] = useState(0)
  const [layout, setLayout] = useState<ShareBannerLayout>('row')
  const [orientation, setOrientation] =
    useState<ShareBannerOrientation>('horizontal')
  const [previewUrl, setPreviewUrl] = useState('')
  const [pngBase64, setPngBase64] = useState('')
  const [imageLoaded, setImageLoaded] = useState(false)
  const [buildingPreview, setBuildingPreview] = useState(false)

  useEffect(() => {
    if (!open || !session) {
      return
    }
    setOrderedPosters(session.posters)
    setPosterCount(defaultPosterCount(session.posters.length))
    setLayout('row')
    setOrientation('horizontal')
  }, [open, session])

  useEffect(() => {
    if (!open || !session || orderedPosters.length === 0 || posterCount === 0) {
      return
    }

    let cancelled = false

    async function rebuildPreview() {
      setBuildingPreview(true)
      setImageLoaded(false)
      try {
        const png = await buildShareBannerPng({
          username: session!.username,
          categoryLabel: session!.categoryLabel,
          categoryCount: session!.categoryCount,
          posters: orderedPosters.slice(0, posterCount),
          layout,
          orientation,
        })
        if (cancelled) {
          return
        }
        const previewBytes = new Uint8Array(png.byteLength)
        previewBytes.set(png)
        const createdUrl = URL.createObjectURL(
          new Blob([previewBytes], {type: 'image/png'}),
        )
        if (cancelled) {
          URL.revokeObjectURL(createdUrl)
          return
        }
        setPreviewUrl((current) => {
          if (current) {
            URL.revokeObjectURL(current)
          }
          return createdUrl
        })
        setPngBase64(uint8ToBase64(png))
      } catch (err) {
        if (!cancelled) {
          onError(errorMessage(err))
        }
      } finally {
        if (!cancelled) {
          setBuildingPreview(false)
        }
      }
    }

    void rebuildPreview()

    return () => {
      cancelled = true
    }
    // onError is a notice sink from the parent; do not rebuild when its identity changes.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, session, orderedPosters, posterCount, layout, orientation])

  useEffect(() => {
    if (open) {
      return
    }
    setPreviewUrl((current) => {
      if (current) {
        URL.revokeObjectURL(current)
      }
      return ''
    })
    setPngBase64('')
    setImageLoaded(false)
  }, [open])

  const countOptions = posterCountOptions(orderedPosters.length)
  const busy = saving || buildingPreview
  const bannerSize = shareBannerSize(orientation)
  const previewAspectClass =
    orientation === 'vertical'
      ? 'mx-auto aspect-[1080/1920] max-h-[55vh] w-auto'
      : 'aspect-[1200/630] w-full'

  return (
    <Dialog.Root
      open={open}
      onOpenChange={(nextOpen) => {
        if (!nextOpen && !saving) {
          onClose()
        }
      }}
    >
      <Dialog.Portal>
        <Dialog.Backdrop />
        <Dialog.Viewport>
          <Dialog.Panel
            className="max-w-3xl p-6"
            aria-labelledby="share-banner-title"
          >
            <Dialog.Title id="share-banner-title">
              Share {session?.categoryLabel ?? ''}
            </Dialog.Title>
            <Dialog.Description>
              Preview the banner, adjust options, then save a PNG.
            </Dialog.Description>

            <div className="relative mt-5 flex justify-center border border-border bg-background p-3">
              {(buildingPreview || !imageLoaded) && (
                <div
                  className={cn(
                    'flex items-center justify-center text-sm text-muted-foreground',
                    previewAspectClass,
                  )}
                >
                  {buildingPreview ? 'Building preview…' : 'Loading preview…'}
                </div>
              )}
              {previewUrl ? (
                <img
                  src={previewUrl}
                  alt={`${session?.categoryLabel ?? ''} share banner preview`}
                  className={
                    imageLoaded && !buildingPreview
                      ? cn(previewAspectClass, 'object-contain')
                      : 'absolute opacity-0'
                  }
                  onLoad={() => setImageLoaded(true)}
                />
              ) : null}
            </div>

            <div className="mt-5 grid grid-cols-1 gap-4 sm:grid-cols-3">
              <div>
                <div className="mb-2 flex h-5 items-center gap-1">
                  <Label htmlFor="share-orientation">Orientation</Label>
                  <Tooltip>
                    <TooltipTrigger
                      render={
                        <button
                          type="button"
                          className="inline-flex size-5 items-center justify-center text-muted-foreground hover:text-foreground"
                          aria-label="Banner image size"
                        />
                      }
                    >
                      <IconHelp className="size-3.5" />
                    </TooltipTrigger>
                    <TooltipContent>
                      {bannerSize.width}×{bannerSize.height}
                    </TooltipContent>
                  </Tooltip>
                </div>
                <NativeSelect
                  id="share-orientation"
                  value={orientation}
                  disabled={busy}
                  onChange={(event) =>
                    setOrientation(event.target.value as ShareBannerOrientation)
                  }
                  className="w-full bg-card"
                >
                  <NativeSelectOption value="horizontal">
                    Horizontal
                  </NativeSelectOption>
                  <NativeSelectOption value="vertical">
                    Vertical
                  </NativeSelectOption>
                </NativeSelect>
              </div>

              <div>
                <div className="mb-2 flex h-5 items-center">
                  <Label htmlFor="share-poster-count">Posters</Label>
                </div>
                <NativeSelect
                  id="share-poster-count"
                  value={String(posterCount)}
                  disabled={busy || countOptions.length === 0}
                  onChange={(event) =>
                    setPosterCount(Number(event.target.value))
                  }
                  className="w-full bg-card"
                >
                  {countOptions.map((count) => (
                    <NativeSelectOption key={count} value={String(count)}>
                      {count} of {orderedPosters.length}
                    </NativeSelectOption>
                  ))}
                </NativeSelect>
              </div>

              <div>
                <div className="mb-2 flex h-5 items-center">
                  <Label htmlFor="share-poster-layout">Layout</Label>
                </div>
                <NativeSelect
                  id="share-poster-layout"
                  value={layout}
                  disabled={busy}
                  onChange={(event) =>
                    setLayout(event.target.value as ShareBannerLayout)
                  }
                  className="w-full bg-card"
                >
                  <NativeSelectOption value="row">Row</NativeSelectOption>
                  <NativeSelectOption value="column">Column</NativeSelectOption>
                </NativeSelect>
              </div>
            </div>

            <p className="mt-5 text-sm font-medium">Poster order</p>
            <ul
              className="mt-2 max-h-56 overflow-y-auto border border-border"
              aria-label="Poster order"
            >
              {orderedPosters.map((poster, index) => {
                const included = index < posterCount
                return (
                  <li
                    key={poster.mediaId}
                    className={cn(
                      'flex h-11 items-center gap-2 border-b border-border px-3 last:border-b-0',
                      !included && 'opacity-50',
                    )}
                  >
                    <span className="w-6 shrink-0 text-xs tabular-nums text-muted-foreground">
                      {index + 1}
                    </span>
                    <span className="min-w-0 flex-1 truncate text-sm">
                      {poster.title}
                    </span>
                    {!included && (
                      <span className="shrink-0 text-xs text-muted-foreground">
                        Hidden
                      </span>
                    )}
                    <div className="flex shrink-0">
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="size-11"
                        aria-label={`Move ${poster.title} up`}
                        disabled={busy || index === 0}
                        onClick={() =>
                          setOrderedPosters((current) =>
                            movePoster(current, index, -1),
                          )
                        }
                      >
                        <IconChevronUp className="size-4" />
                      </Button>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="size-11"
                        aria-label={`Move ${poster.title} down`}
                        disabled={busy || index === orderedPosters.length - 1}
                        onClick={() =>
                          setOrderedPosters((current) =>
                            movePoster(current, index, 1),
                          )
                        }
                      >
                        <IconChevronDown className="size-4" />
                      </Button>
                    </div>
                  </li>
                )
              })}
            </ul>

            <div className="mt-5 flex justify-end gap-2">
              <Button
                type="button"
                variant="secondary"
                className="min-w-28"
                disabled={saving}
                onClick={onClose}
              >
                Cancel
              </Button>
              <Button
                type="button"
                className="min-w-28"
                disabled={busy || !pngBase64 || !session}
                onClick={() => {
                  if (!session || !pngBase64) {
                    return
                  }
                  const baseName = session.defaultFilename.replace(
                    /\.png$/i,
                    '',
                  )
                  onSave(pngBase64, `${baseName}-${orientation}.png`)
                }}
              >
                {saving ? 'Saving…' : 'Save PNG'}
              </Button>
            </div>
          </Dialog.Panel>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
