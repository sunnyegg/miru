import {FetchShareImage} from '../../wailsjs/go/main/App'

export type ShareBannerPoster = {
  mediaId: number
  title: string
  coverImage: string
}

export type ShareBannerLayout = 'row' | 'column'
export type ShareBannerOrientation = 'horizontal' | 'vertical'

export type ShareBannerInput = {
  username: string
  categoryLabel: string
  categoryCount: number
  posters: ShareBannerPoster[]
  layout: ShareBannerLayout
  orientation: ShareBannerOrientation
}

/** Soft floor: posters smaller than this stop being readable as a mosaic. */
export const SHARE_BANNER_MIN_CELL_WIDTH = 64
const titleMinCellWidth = 100

const posterAspect = 2 / 3 // width / height — AniList cover shape
const coverBase64Cache = new Map<string, string>()

const colorBackground = '#111111'
const colorCard = '#181818'
const colorBorder = '#2a2a2a'
const colorText = '#f2f2f2'
const colorMuted = '#a1a1a1'
const colorPrimary = '#a78bfa'
const colorFooter = '#777777'

export function shareBannerSize(orientation: ShareBannerOrientation): {
  width: number
  height: number
} {
  if (orientation === 'vertical') {
    return {width: 1080, height: 1920}
  }
  return {width: 1200, height: 630}
}

type BannerMetrics = {
  canvasWidth: number
  canvasHeight: number
  paddingX: number
  paddingTop: number
  footerReserve: number
  brandSize: number
  nameSize: number
  countSize: number
  gap: number
  horizontalSplit: boolean
  horizontalLeftWidth: number
  horizontalColumnGap: number
  posterAreaX: number
  posterTop: number
  areaWidth: number
  areaHeight: number
  textMaxWidth: number
}

function bannerMetrics(orientation: ShareBannerOrientation): BannerMetrics {
  const {width: canvasWidth, height: canvasHeight} =
    shareBannerSize(orientation)
  const isVertical = orientation === 'vertical'
  const paddingX = isVertical ? 64 : 56
  const paddingTop = isVertical ? 72 : 44
  const footerReserve = isVertical ? 88 : 64
  const brandSize = isVertical ? 22 : 18
  const nameSize = isVertical ? 64 : 48
  const countSize = isVertical ? 34 : 26
  const gap = isVertical ? 14 : 10
  const horizontalSplit = !isVertical
  const horizontalLeftWidth = 340
  const horizontalColumnGap = 32

  // Vertical text block height (matches draw order in buildShareBannerPng).
  const nameY = paddingTop + brandSize + (isVertical ? 56 : 44)
  const countY = nameY + (isVertical ? 52 : 40)
  const posterTop = isVertical ? countY + 56 : paddingTop
  const posterAreaX = horizontalSplit
    ? paddingX + horizontalLeftWidth + horizontalColumnGap
    : paddingX
  const posterBottom = canvasHeight - footerReserve
  const areaHeight = posterBottom - posterTop
  const areaWidth = horizontalSplit
    ? canvasWidth - paddingX - posterAreaX
    : canvasWidth - paddingX * 2
  const textMaxWidth = horizontalSplit
    ? horizontalLeftWidth
    : canvasWidth - paddingX * 2

  return {
    canvasWidth,
    canvasHeight,
    paddingX,
    paddingTop,
    footerReserve,
    brandSize,
    nameSize,
    countSize,
    gap,
    horizontalSplit,
    horizontalLeftWidth,
    horizontalColumnGap,
    posterAreaX,
    posterTop,
    areaWidth,
    areaHeight,
    textMaxWidth,
  }
}

/** How many posters still read at min cell size for this orientation + layout. */
export function shareBannerPosterCapacity(
  orientation: ShareBannerOrientation,
  layout: ShareBannerLayout,
): number {
  const metrics = bannerMetrics(orientation)
  const minCellHeight = SHARE_BANNER_MIN_CELL_WIDTH / posterAspect
  const maxColumns = Math.max(
    1,
    Math.floor(
      (metrics.areaWidth + metrics.gap) /
        (SHARE_BANNER_MIN_CELL_WIDTH + metrics.gap),
    ),
  )
  const maxRows = Math.max(
    1,
    Math.floor(
      (metrics.areaHeight + metrics.gap) / (minCellHeight + metrics.gap),
    ),
  )
  if (layout === 'row') {
    return maxColumns
  }
  return maxColumns * maxRows
}

export function uint8ToBase64(bytes: Uint8Array): string {
  const chunkSize = 0x8000
  let binary = ''
  for (let offset = 0; offset < bytes.length; offset += chunkSize) {
    const chunk = bytes.subarray(offset, offset + chunkSize)
    binary += String.fromCharCode(...chunk)
  }
  return btoa(binary)
}

function base64ToUint8Array(value: string): Uint8Array {
  const binary = atob(value)
  const bytes = new Uint8Array(binary.length)
  for (let index = 0; index < binary.length; index++) {
    bytes[index] = binary.charCodeAt(index)
  }
  return bytes
}

function bytesToImageBitmap(base64: string): Promise<ImageBitmap> {
  const bytes = base64ToUint8Array(base64)
  const copy = new Uint8Array(bytes.byteLength)
  copy.set(bytes)
  return createImageBitmap(new Blob([copy]))
}

async function loadPosterBitmap(
  coverImage: string,
): Promise<ImageBitmap | null> {
  if (!coverImage) {
    return null
  }
  try {
    let base64 = coverBase64Cache.get(coverImage)
    if (!base64) {
      base64 = await FetchShareImage(coverImage)
      if (!base64) {
        return null
      }
      coverBase64Cache.set(coverImage, base64)
    }
    return await bytesToImageBitmap(base64)
  } catch {
    return null
  }
}

function drawCover(
  context: CanvasRenderingContext2D,
  bitmap: ImageBitmap,
  x: number,
  y: number,
  width: number,
  height: number,
) {
  const scale = Math.max(width / bitmap.width, height / bitmap.height)
  const drawWidth = bitmap.width * scale
  const drawHeight = bitmap.height * scale
  const drawX = x + (width - drawWidth) / 2
  const drawY = y + (height - drawHeight) / 2
  context.save()
  context.beginPath()
  context.rect(x, y, width, height)
  context.clip()
  context.drawImage(bitmap, drawX, drawY, drawWidth, drawHeight)
  context.restore()
}

function truncateText(
  context: CanvasRenderingContext2D,
  text: string,
  maxWidth: number,
): string {
  if (context.measureText(text).width <= maxWidth) {
    return text
  }
  let truncated = text
  while (
    truncated.length > 0 &&
    context.measureText(`${truncated}…`).width > maxWidth
  ) {
    truncated = truncated.slice(0, -1)
  }
  return truncated.length > 0 ? `${truncated}…` : '…'
}

type PosterGrid = {
  columns: number
  rows: number
  cellWidth: number
  cellHeight: number
  originX: number
  originY: number
}

function computePosterGrid(
  posterCount: number,
  areaWidth: number,
  areaHeight: number,
  gap: number,
  layout: ShareBannerLayout,
): PosterGrid {
  const maxColumnsByWidth = Math.max(
    1,
    Math.floor((areaWidth + gap) / (SHARE_BANNER_MIN_CELL_WIDTH + gap)),
  )
  const columnChoices: number[] = []
  if (layout === 'row') {
    columnChoices.push(Math.min(posterCount, maxColumnsByWidth))
  } else {
    const maxColumns = Math.min(posterCount, maxColumnsByWidth)
    for (let columns = 1; columns <= maxColumns; columns++) {
      columnChoices.push(columns)
    }
  }

  let best: (PosterGrid & {score: number}) | null = null

  for (const columns of columnChoices) {
    const rows = Math.ceil(posterCount / columns)
    const maxCellWidth = (areaWidth - gap * (columns - 1)) / columns
    const maxCellHeight = (areaHeight - gap * (rows - 1)) / rows
    if (maxCellWidth < SHARE_BANNER_MIN_CELL_WIDTH || maxCellHeight <= 0) {
      continue
    }

    let cellWidth = maxCellWidth
    let cellHeight = cellWidth / posterAspect
    if (cellHeight > maxCellHeight) {
      cellHeight = maxCellHeight
      cellWidth = cellHeight * posterAspect
    }
    if (cellWidth < SHARE_BANNER_MIN_CELL_WIDTH) {
      continue
    }

    const gridWidth = columns * cellWidth + gap * (columns - 1)
    const gridHeight = rows * cellHeight + gap * (rows - 1)
    const fillRatio =
      (gridWidth * gridHeight) / Math.max(1, areaWidth * areaHeight)
    // Prefer filling the plane; then larger covers.
    const score = fillRatio * 10_000 + cellWidth * cellHeight

    if (!best || score > best.score) {
      best = {
        columns,
        rows,
        cellWidth,
        cellHeight,
        originX: 0,
        originY: 0,
        score,
      }
    }
  }

  if (!best) {
    return {
      columns: 1,
      rows: 1,
      cellWidth: Math.min(areaWidth, SHARE_BANNER_MIN_CELL_WIDTH),
      cellHeight: Math.min(
        areaHeight,
        SHARE_BANNER_MIN_CELL_WIDTH / posterAspect,
      ),
      originX: 0,
      originY: 0,
    }
  }

  return {
    columns: best.columns,
    rows: best.rows,
    cellWidth: best.cellWidth,
    cellHeight: best.cellHeight,
    originX: best.originX,
    originY: best.originY,
  }
}

export async function buildShareBannerPng(
  input: ShareBannerInput,
): Promise<Uint8Array> {
  const capacity = shareBannerPosterCapacity(input.orientation, input.layout)
  const posters = input.posters.slice(0, capacity)
  const metrics = bannerMetrics(input.orientation)

  const canvas = document.createElement('canvas')
  canvas.width = metrics.canvasWidth
  canvas.height = metrics.canvasHeight
  const context = canvas.getContext('2d')
  if (!context) {
    throw new Error('Could not create banner canvas')
  }

  if (document.fonts?.ready) {
    await document.fonts.ready
  }

  const bitmaps = await Promise.all(
    posters.map((poster) => loadPosterBitmap(poster.coverImage)),
  )

  context.fillStyle = colorBackground
  context.fillRect(0, 0, metrics.canvasWidth, metrics.canvasHeight)

  context.fillStyle = colorMuted
  context.font = `600 ${metrics.brandSize}px "Source Sans 3 Variable", "Source Sans 3", sans-serif`
  context.fillText(
    'MIRU',
    metrics.paddingX,
    metrics.paddingTop + metrics.brandSize,
  )

  const displayName = input.username.trim() || 'AniList user'
  const nameY =
    metrics.paddingTop + metrics.brandSize + (metrics.horizontalSplit ? 44 : 56)
  context.fillStyle = colorText
  context.font = `700 ${metrics.nameSize}px "Source Sans 3 Variable", "Source Sans 3", sans-serif`
  context.fillText(
    truncateText(context, displayName, metrics.textMaxWidth),
    metrics.paddingX,
    nameY,
  )

  const countY = nameY + (metrics.horizontalSplit ? 40 : 52)
  const countDigits = String(input.categoryCount)
  context.font = `700 ${metrics.countSize}px "Source Sans 3 Variable", "Source Sans 3", sans-serif`
  context.fillStyle = colorPrimary
  context.fillText(countDigits, metrics.paddingX, countY)
  const countWidth = context.measureText(countDigits).width
  context.fillStyle = colorText
  context.font = `500 ${metrics.countSize}px "Source Sans 3 Variable", "Source Sans 3", sans-serif`
  context.fillText(
    ` ${input.categoryLabel}`,
    metrics.paddingX + countWidth,
    countY,
  )

  context.fillStyle = colorPrimary
  context.fillRect(metrics.paddingX, countY + 16, 48, 3)

  const hiddenPosterCount = Math.max(0, input.categoryCount - posters.length)
  if (hiddenPosterCount > 0 && metrics.horizontalSplit) {
    context.fillStyle = colorMuted
    context.font =
      '500 18px "Source Sans 3 Variable", "Source Sans 3", sans-serif'
    context.fillText(
      `+${hiddenPosterCount} more`,
      metrics.paddingX,
      countY + 52,
    )
  }

  if (posters.length === 0) {
    context.strokeStyle = colorBorder
    context.strokeRect(
      metrics.posterAreaX,
      metrics.posterTop,
      metrics.areaWidth,
      metrics.areaHeight,
    )
    context.fillStyle = colorMuted
    context.font =
      '500 22px "Source Sans 3 Variable", "Source Sans 3", sans-serif'
    context.fillText(
      'No posters',
      metrics.posterAreaX + 24,
      metrics.posterTop + metrics.areaHeight / 2,
    )
  } else {
    const grid = computePosterGrid(
      posters.length,
      metrics.areaWidth,
      metrics.areaHeight,
      metrics.gap,
      input.layout,
    )
    const showTitles = grid.cellWidth >= titleMinCellWidth
    const titleHeight = showTitles
      ? Math.min(44, Math.max(26, grid.cellHeight * 0.18))
      : 0
    const titleFontSize =
      grid.cellWidth < 110 ? 12 : grid.cellWidth < 160 ? 14 : 16

    posters.forEach((poster, index) => {
      const column = index % grid.columns
      const row = Math.floor(index / grid.columns)
      const x =
        metrics.posterAreaX +
        grid.originX +
        column * (grid.cellWidth + metrics.gap)
      const y =
        metrics.posterTop + grid.originY + row * (grid.cellHeight + metrics.gap)

      context.fillStyle = colorCard
      context.fillRect(x, y, grid.cellWidth, grid.cellHeight)

      const bitmap = bitmaps[index]
      if (bitmap) {
        drawCover(context, bitmap, x, y, grid.cellWidth, grid.cellHeight)
        bitmap.close()
      }

      if (!showTitles) {
        return
      }

      const gradient = context.createLinearGradient(
        0,
        y + grid.cellHeight - titleHeight,
        0,
        y + grid.cellHeight,
      )
      gradient.addColorStop(0, 'rgba(0,0,0,0)')
      gradient.addColorStop(1, 'rgba(0,0,0,0.82)')
      context.fillStyle = gradient
      context.fillRect(
        x,
        y + grid.cellHeight - titleHeight,
        grid.cellWidth,
        titleHeight,
      )

      context.fillStyle = colorText
      context.font = `600 ${titleFontSize}px "Source Sans 3 Variable", "Source Sans 3", sans-serif`
      const title = truncateText(context, poster.title, grid.cellWidth - 16)
      context.fillText(title, x + 8, y + grid.cellHeight - 10)
    })
  }

  context.fillStyle = colorFooter
  context.font = `500 ${metrics.horizontalSplit ? 18 : 22}px "Source Sans 3 Variable", "Source Sans 3", sans-serif`
  context.fillText(
    'My AniList on Miru',
    metrics.paddingX,
    metrics.canvasHeight - (metrics.horizontalSplit ? 28 : 40),
  )

  const blob = await new Promise<Blob>((resolve, reject) => {
    canvas.toBlob((result) => {
      if (!result) {
        reject(new Error('Could not encode banner PNG'))
        return
      }
      resolve(result)
    }, 'image/png')
  })
  return new Uint8Array(await blob.arrayBuffer())
}
