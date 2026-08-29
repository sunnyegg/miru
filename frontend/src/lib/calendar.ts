import type {AiringScheduleView} from './types'

export const dayFormatter = new Intl.DateTimeFormat(undefined, {
  weekday: 'long',
  month: 'long',
  day: 'numeric',
})

export const timeFormatter = new Intl.DateTimeFormat(undefined, {
  hour: 'numeric',
  minute: '2-digit',
})

export function startOfMonday(date: Date): Date {
  const start = new Date(date)
  start.setHours(0, 0, 0, 0)
  const day = start.getDay()
  const daysSinceMonday = (day + 6) % 7
  start.setDate(start.getDate() - daysSinceMonday)
  return start
}

export function dateKey(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function isToday(date: Date): boolean {
  return dateKey(date) === dateKey(new Date())
}

export function buildWeekDays(weekStart: Date): Date[] {
  return Array.from({length: 7}, (_, index) => {
    const day = new Date(weekStart)
    day.setDate(day.getDate() + index)
    return day
  })
}

export function scheduleTitle(schedule: AiringScheduleView): string {
  return schedule.titleEnglish || schedule.titleRomaji
}

export function groupSchedulesByDay(schedules: AiringScheduleView[]): Map<string, AiringScheduleView[]> {
  const grouped = new Map<string, AiringScheduleView[]>()
  for (const schedule of schedules) {
    const key = dateKey(new Date(schedule.airingAt * 1000))
    const current = grouped.get(key) ?? []
    current.push(schedule)
    grouped.set(key, current)
  }
  for (const entries of grouped.values()) {
    entries.sort((left, right) => left.airingAt - right.airingAt)
  }
  return grouped
}
