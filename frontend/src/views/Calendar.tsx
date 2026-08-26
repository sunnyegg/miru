import {useEffect, useMemo, useState} from 'react'
import {ListAiringSchedule} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {AiringScheduleView} from '../lib/types'

type Props = {
  notice: (msg: string, isError?: boolean) => void
}

const dayFormatter = new Intl.DateTimeFormat(undefined, {
  weekday: 'long',
  month: 'long',
  day: 'numeric',
})

const timeFormatter = new Intl.DateTimeFormat(undefined, {
  hour: 'numeric',
  minute: '2-digit',
})

function startOfMonday(date: Date): Date {
  const start = new Date(date)
  start.setHours(0, 0, 0, 0)
  const day = start.getDay()
  const daysSinceMonday = (day + 6) % 7
  start.setDate(start.getDate() - daysSinceMonday)
  return start
}

function dateKey(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function CalendarView({notice}: Props) {
  const [weekOffset, setWeekOffset] = useState(0)
  const [schedules, setSchedules] = useState<AiringScheduleView[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const weekStart = useMemo(() => {
    const start = startOfMonday(new Date())
    start.setDate(start.getDate() + weekOffset * 7)
    return start
  }, [weekOffset])

  const days = useMemo(() => {
    return Array.from({length: 7}, (_, index) => {
      const day = new Date(weekStart)
      day.setDate(day.getDate() + index)
      return day
    })
  }, [weekStart])

  const schedulesByDay = useMemo(() => {
    const grouped = new Map<string, AiringScheduleView[]>()
    for (const schedule of schedules) {
      const key = dateKey(new Date(schedule.airingAt * 1000))
      const current = grouped.get(key) ?? []
      current.push(schedule)
      grouped.set(key, current)
    }
    return grouped
  }, [schedules])

  async function loadSchedules() {
    setLoading(true)
    setError('')
    try {
      const start = Math.floor(weekStart.getTime() / 1000)
      const end = Math.floor(days[6].getTime() / 1000) + 24 * 60 * 60
      const result = await ListAiringSchedule(Math.max(0, start - 1), end)
      setSchedules(result ?? [])
    } catch (err) {
      const message = errorMessage(err)
      setError(message)
      notice(message, true)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadSchedules()
  }, [weekStart, days])

  function resetToToday() {
    setWeekOffset(0)
  }

  return (
    <section className="flex h-full flex-col gap-6">
      <header className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h2 className="text-2xl font-semibold">Airing Calendar</h2>
          <p className="mt-1 text-sm text-muted-foreground">Upcoming episodes from AniList, shown in your local time.</p>
        </div>
        <div className="flex flex-wrap gap-2" aria-label="Calendar navigation">
          <button
            type="button"
            onClick={() => setWeekOffset((offset) => offset - 1)}
            className="min-h-11 cursor-pointer bg-muted px-3 text-sm text-foreground hover:bg-secondary"
          >
            Previous
          </button>
          <button
            type="button"
            onClick={resetToToday}
            disabled={weekOffset === 0}
            className="min-h-11 cursor-pointer bg-secondary px-3 text-sm text-on-secondary disabled:cursor-not-allowed disabled:opacity-50"
          >
            Today
          </button>
          <button
            type="button"
            onClick={() => setWeekOffset((offset) => offset + 1)}
            className="min-h-11 cursor-pointer bg-muted px-3 text-sm text-foreground hover:bg-secondary"
          >
            Next
          </button>
        </div>
      </header>

      <p className="text-sm font-medium text-muted-foreground">
        {days[0].toLocaleDateString(undefined, {month: 'short', day: 'numeric'})}
        {' – '}
        {days[6].toLocaleDateString(undefined, {month: 'short', day: 'numeric', year: 'numeric'})}
      </p>

      {loading ? (
        <p className="border border-border/40 bg-card p-8 text-sm text-muted-foreground" role="status">
          Loading airing schedule…
        </p>
      ) : error ? (
        <div className="border border-destructive/60 bg-card p-8" role="alert">
          <p className="text-sm text-destructive">{error}</p>
          <button
            type="button"
            onClick={() => void loadSchedules()}
            className="mt-4 min-h-11 cursor-pointer bg-secondary px-4 text-sm text-on-secondary"
          >
            Try again
          </button>
        </div>
      ) : (
        <div className="grid gap-4 lg:grid-cols-2">
          {days.map((day) => {
            const entries = schedulesByDay.get(dateKey(day)) ?? []
            return (
              <section key={dateKey(day)} className="bg-card p-4">
                <h3 className="font-medium">{dayFormatter.format(day)}</h3>
                {entries.length === 0 ? (
                  <p className="mt-4 text-sm text-muted-foreground">No episodes scheduled.</p>
                ) : (
                  <ul className="mt-3 flex flex-col gap-3">
                    {entries.map((schedule) => (
                      <li key={schedule.id} className="flex items-center gap-3">
                        {schedule.coverImage ? (
                          <img src={schedule.coverImage} alt="" width={40} height={56} className="h-14 w-10 object-cover" />
                        ) : (
                          <span className="h-14 w-10 bg-muted" aria-hidden="true" />
                        )}
                        <div className="min-w-0 flex-1">
                          <p className="truncate text-sm font-medium">{schedule.titleEnglish || schedule.titleRomaji}</p>
                          <p className="text-xs text-muted-foreground">
                            Episode {schedule.episode} · {timeFormatter.format(new Date(schedule.airingAt * 1000))}
                          </p>
                        </div>
                      </li>
                    ))}
                  </ul>
                )}
              </section>
            )
          })}
        </div>
      )}
    </section>
  )
}
