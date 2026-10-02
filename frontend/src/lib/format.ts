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
