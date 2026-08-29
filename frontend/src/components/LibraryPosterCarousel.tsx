import {Children, useEffect, useRef, useState, type ReactNode} from 'react'
import {IconBack, IconChevronRight} from './Icons'
import {Button} from '@/components/ui/button'
import {cn} from '@/lib/utils'

const SHELF_GAP_PX = 20

type Props = {
  children: ReactNode
  ariaLabel: string
  ariaBusy?: boolean
}

type PageMetrics = {
  viewportWidth: number
  itemWidth: number
  itemCount: number
  itemsPerPage: number
  maxPage: number
}

function computePageMetrics(
  viewport: HTMLElement | null,
  track: HTMLElement | null,
): PageMetrics {
  const itemCount = track ? track.children.length : 0
  const viewportWidth = viewport?.clientWidth ?? 0
  const firstItem = track?.firstElementChild as HTMLElement | null
  const itemWidth = firstItem?.offsetWidth ?? 0

  if (viewportWidth <= 0 || itemWidth <= 0 || itemCount === 0) {
    return {
      viewportWidth,
      itemWidth,
      itemCount,
      itemsPerPage: 1,
      maxPage: 0,
    }
  }

  const itemsPerPage = Math.max(
    1,
    Math.floor((viewportWidth + SHELF_GAP_PX) / (itemWidth + SHELF_GAP_PX)),
  )
  const maxPage = Math.max(0, Math.ceil(itemCount / itemsPerPage) - 1)

  return {viewportWidth, itemWidth, itemCount, itemsPerPage, maxPage}
}

export function LibraryPosterCarousel({
  children,
  ariaLabel,
  ariaBusy = false,
}: Props) {
  const viewportRef = useRef<HTMLDivElement>(null)
  const trackRef = useRef<HTMLUListElement>(null)
  const [pageIndex, setPageIndex] = useState(0)
  const [pageMetrics, setPageMetrics] = useState<PageMetrics>(() =>
    computePageMetrics(null, null),
  )

  const itemCount = Children.count(children)

  useEffect(() => {
    setPageIndex(0)
  }, [itemCount])

  useEffect(() => {
    const viewport = viewportRef.current
    const track = trackRef.current
    if (!viewport || !track) {
      return
    }

    function updateMetrics() {
      const nextMetrics = computePageMetrics(viewport, track)
      setPageMetrics(nextMetrics)
      setPageIndex((currentPage) => Math.min(currentPage, nextMetrics.maxPage))
    }

    updateMetrics()

    const resizeObserver = new ResizeObserver(updateMetrics)
    resizeObserver.observe(viewport)
    resizeObserver.observe(track)

    return () => {
      resizeObserver.disconnect()
    }
  }, [itemCount])

  const {itemWidth, itemsPerPage, maxPage} = pageMetrics
  const pageStride = itemsPerPage * (itemWidth + SHELF_GAP_PX)
  const offset = pageIndex * pageStride
  const showNavigation = maxPage > 0

  const atStart = pageIndex === 0
  const atEnd = pageIndex >= maxPage

  function goToPreviousPage() {
    setPageIndex((currentPage) => Math.max(0, currentPage - 1))
  }

  function goToNextPage() {
    setPageIndex((currentPage) => Math.min(maxPage, currentPage + 1))
  }

  return (
    <div className="group relative p-1">
      {showNavigation && (
        <Button
          type="button"
          variant="ghost"
          size="icon"
          aria-label="Previous"
          disabled={atStart}
          onClick={goToPreviousPage}
          className={cn(
            'absolute top-1/2 left-1 z-10 -translate-y-1/2 bg-background/80',
            'opacity-0 transition-opacity duration-200 motion-reduce:opacity-100 motion-reduce:transition-none',
            atStart
              ? 'group-hover:opacity-30 group-focus-within:opacity-30'
              : 'group-hover:opacity-100 group-focus-within:opacity-100',
          )}
        >
          <IconBack className="size-5" />
        </Button>
      )}
      <div ref={viewportRef} className="overflow-hidden">
        <ul
          ref={trackRef}
          aria-label={ariaLabel}
          aria-busy={ariaBusy || undefined}
          className="flex gap-5 transition-transform duration-300 motion-reduce:transition-none"
          style={{transform: `translateX(-${offset}px)`}}
        >
          {children}
        </ul>
      </div>
      {showNavigation && (
        <Button
          type="button"
          variant="ghost"
          size="icon"
          aria-label="Next"
          disabled={atEnd}
          onClick={goToNextPage}
          className={cn(
            'absolute top-1/2 right-1 z-10 -translate-y-1/2 bg-background/80',
            'opacity-0 transition-opacity duration-200 motion-reduce:opacity-100 motion-reduce:transition-none',
            atEnd
              ? 'group-hover:opacity-30 group-focus-within:opacity-30'
              : 'group-hover:opacity-100 group-focus-within:opacity-100',
          )}
        >
          <IconChevronRight className="size-5" />
        </Button>
      )}
    </div>
  )
}
