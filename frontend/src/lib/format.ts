export function formatBytes(n: number): string {
  if (!n || n <= 0) {
    return '0 B'
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = n
  let i = 0
  while (value >= 1024 && i < units.length - 1) {
    value /= 1024
    i += 1
  }
  return `${value.toFixed(value >= 10 || i === 0 ? 0 : 1)} ${units[i]}`
}

export function formatSpeed(bytesPerSecond: number): string {
  return `${formatBytes(bytesPerSecond)}/s`
}

const sizeUnitPowers: Record<string, number> = {
  B: 0,
  KB: 1,
  KIB: 1,
  MB: 2,
  MIB: 2,
  GB: 3,
  GIB: 3,
  TB: 4,
  TIB: 4,
}

// parseSize reads an indexer's free-form size text such as "1.4 GiB" or "800MB".
// Every unit is 1024-based so a summed batch total never surprises; anything it
// cannot read returns null.
export function parseSize(input: string): number | null {
  const match = /^(\d+(?:[.,]\d+)?)\s*([a-z]+)$/i.exec(input.trim())
  if (!match) {
    return null
  }
  const power = sizeUnitPowers[match[2].toUpperCase()]
  if (power === undefined) {
    return null
  }
  const amount = Number.parseFloat(match[1].replace(',', '.'))
  if (!Number.isFinite(amount)) {
    return null
  }
  return amount * 1024 ** power
}

// formatAge renders how old a cached AniList response is for the stale banner.
export function formatAge(seconds: number): string {
  const safe = Number.isFinite(seconds) ? Math.max(0, seconds) : 0
  if (safe < 60) {
    return 'just now'
  }
  const minutes = Math.round(safe / 60)
  if (minutes < 60) {
    return `${minutes} minute${minutes === 1 ? '' : 's'} ago`
  }
  const hours = Math.round(minutes / 60)
  if (hours < 24) {
    return `${hours} hour${hours === 1 ? '' : 's'} ago`
  }
  const days = Math.round(hours / 24)
  return `${days} day${days === 1 ? '' : 's'} ago`
}

export function errorMessage(err: unknown): string {
  const message =
    err instanceof Error
      ? err.message
      : typeof err === 'string'
        ? err
        : String(err)

  if (
    /\bhttp(?:\s+status)?\s+429\b|\b429\s+too many requests\b/i.test(message)
  ) {
    return 'Too many requests. Please wait a moment, then try again.'
  }

  return message
}
