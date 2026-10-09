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

export const SHARE_BANNER_MAX_POSTERS = 12

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
  const columnChoices: number[] = []
  if (layout === 'row') {
    columnChoices.push(posterCount)
  } else {
    const maxColumns = Math.min(posterCount, 4)
    for (let columns = 1; columns <= maxColumns; columns++) {
      columnChoices.push(columns)
    }
  }

  let best: (PosterGrid & {score: number}) | null = null

  for (const columns of columnChoices) {
    const rows = Math.ceil(posterCount / columns)
    const maxCellWidth = (areaWidth - gap * (columns - 1)) / columns
    const maxCellHeight = (areaHeight - gap * (rows - 1)) / rows
    if (maxCellWidth <= 0 || maxCellHeight <= 0) {
      continue
    }

    let cellWidth = maxCellWidth
    let cellHeight = cellWidth / posterAspect
    if (cellHeight > maxCellHeight) {
      cellHeight = maxCellHeight
      cellWidth = cellHeight * posterAspect
    }

    const gridWidth = columns * cellWidth + gap * (columns - 1)
    const gridHeight = rows * cellHeight + gap * (rows - 1)
    // Prefer larger posters; slight bias to more rows in column mode.
    const score =
      cellWidth * cellHeight * (layout === 'column' ? 1 + rows * 0.02 : 1)

    if (!best || score > best.score) {
      best = {
        columns,
        rows,
        cellWidth,
        cellHeight,
        originX: (areaWidth - gridWidth) / 2,
        originY: (areaHeight - gridHeight) / 2,
        score,
      }
    }
  }

  if (!best) {
    return {
      columns: 1,
      rows: 1,
      cellWidth: areaWidth,
      cellHeight: areaHeight,
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
  const {width: canvasWidth, height: canvasHeight} = shareBannerSize(
    input.orientation,
  )
  const canvas = document.createElement('canvas')
  canvas.width = canvasWidth
  canvas.height = canvasHeight
  const context = canvas.getContext('2d')
  if (!context) {
    throw new Error('Could not create banner canvas')
  }

  if (document.fonts?.ready) {
    await document.fonts.ready
  }

  const posters = input.posters.slice(0, SHARE_BANNER_MAX_POSTERS)
  const bitmaps = await Promise.all(
    posters.map((poster) => loadPosterBitmap(poster.coverImage)),
  )

  const isVertical = input.orientation === 'vertical'
  const paddingX = isVertical ? 64 : 56
  const paddingTop = isVertical ? 72 : 44
  const footerReserve = isVertical ? 88 : 64
  const brandSize = isVertical ? 22 : 18
  const nameSize = isVertical ? 64 : 48
  const countSize = isVertical ? 34 : 26
  const gap = isVertical ? 18 : 12

  context.fillStyle = colorBackground
  context.fillRect(0, 0, canvasWidth, canvasHeight)

  context.fillStyle = colorMuted
  context.font = `600 ${brandSize}px "Source Sans 3 Variable", "Source Sans 3", sans-serif`
  context.fillText('MIRU', paddingX, paddingTop + brandSize)

  const displayName = input.username.trim() || 'AniList user'
  const nameY = paddingTop + brandSize + (isVertical ? 56 : 44)
  context.fillStyle = colorText
  context.font = `700 ${nameSize}px "Source Sans 3 Variable", "Source Sans 3", sans-serif`
  context.fillText(
    truncateText(context, displayName, canvasWidth - paddingX * 2),
    paddingX,
    nameY,
  )

  const countY = nameY + (isVertical ? 52 : 40)
  const countDigits = String(input.categoryCount)
  context.font = `700 ${countSize}px "Source Sans 3 Variable", "Source Sans 3", sans-serif`
  context.fillStyle = colorPrimary
  context.fillText(countDigits, paddingX, countY)
  const countWidth = context.measureText(countDigits).width
  context.fillStyle = colorText
  context.font = `500 ${countSize}px "Source Sans 3 Variable", "Source Sans 3", sans-serif`
  context.fillText(` ${input.categoryLabel}`, paddingX + countWidth, countY)

  context.fillStyle = colorPrimary
  context.fillRect(paddingX, countY + 16, 48, 3)

  const posterTop = countY + (isVertical ? 56 : 40)
  const posterBottom = canvasHeight - footerReserve
  const areaHeight = posterBottom - posterTop
  const areaWidth = canvasWidth - paddingX * 2

  if (posters.length === 0) {
    context.strokeStyle = colorBorder
    context.strokeRect(paddingX, posterTop, areaWidth, areaHeight)
    context.fillStyle = colorMuted
    context.font =
      '500 22px "Source Sans 3 Variable", "Source Sans 3", sans-serif'
    context.fillText('No posters', paddingX + 24, posterTop + areaHeight / 2)
  } else {
    const grid = computePosterGrid(
      posters.length,
      areaWidth,
      areaHeight,
      gap,
      input.layout,
    )
    const titleHeight = Math.min(44, Math.max(26, grid.cellHeight * 0.18))
    const titleFontSize =
      grid.cellWidth < 110 ? 12 : grid.cellWidth < 160 ? 14 : 16

    posters.forEach((poster, index) => {
      const column = index % grid.columns
      const row = Math.floor(index / grid.columns)
      const x = paddingX + grid.originX + column * (grid.cellWidth + gap)
      const y = posterTop + grid.originY + row * (grid.cellHeight + gap)

      context.fillStyle = colorCard
      context.fillRect(x, y, grid.cellWidth, grid.cellHeight)

      const bitmap = bitmaps[index]
      if (bitmap) {
        drawCover(context, bitmap, x, y, grid.cellWidth, grid.cellHeight)
        bitmap.close()
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
  context.font = `500 ${isVertical ? 22 : 18}px "Source Sans 3 Variable", "Source Sans 3", sans-serif`
  context.fillText(
    'My AniList on Miru',
    paddingX,
    canvasHeight - (isVertical ? 40 : 28),
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
