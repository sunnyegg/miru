import {useEffect, useMemo, useState} from 'react'
import {ListAiringSchedule} from '../../wailsjs/go/main/App'
import {AiringAgendaLayout} from '../components/AiringAgendaLayout'
import {buildWeekDays, groupSchedulesByDay, startOfMonday} from '../lib/calendar'
import {errorMessage} from '../lib/format'
import type {AiringScheduleView} from '../lib/types'
import {Alert, AlertAction, AlertDescription} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'

export function CalendarView() {
  const [weekOffset, setWeekOffset] = useState(0)
  const [scrollToTodayRequest, setScrollToTodayRequest] = useState(0)
  const [schedules, setSchedules] = useState<AiringScheduleView[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const weekStart = useMemo(() => {
    const start = startOfMonday(new Date())
    start.setDate(start.getDate() + weekOffset * 7)
    return start
  }, [weekOffset])

  const days = useMemo(() => buildWeekDays(weekStart), [weekStart])

  const schedulesByDay = useMemo(() => groupSchedulesByDay(schedules), [schedules])

  async function loadSchedules() {
    setLoading(true)
    setError('')
    try {
      const start = Math.floor(weekStart.getTime() / 1000)
      const end = Math.floor(days[6].getTime() / 1000) + 24 * 60 * 60
      const result = await ListAiringSchedule(Math.max(0, start - 1), end)
      setSchedules(result ?? [])
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadSchedules()
    // loadSchedules closes over current weekStart/days; rely on deps to refire.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [weekStart, days])

  return (
    <section className="flex h-full flex-col gap-6">
      <header className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h2 className="text-2xl font-semibold">Airing</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Upcoming episodes from AniList, shown in your local time.
          </p>
        </div>
        <div className="flex flex-wrap gap-2" aria-label="Calendar navigation">
          <Button type="button" variant="muted" onClick={() => setWeekOffset((offset) => offset - 1)}>
            Previous
          </Button>
          <Button
            type="button"
            variant="secondary"
            onClick={() => {
              setWeekOffset(0)
              setScrollToTodayRequest((request) => request + 1)
            }}
          >
            Today
          </Button>
          <Button type="button" variant="muted" onClick={() => setWeekOffset((offset) => offset + 1)}>
            Next
          </Button>
        </div>
      </header>

      <p className="text-sm font-medium text-muted-foreground">
        {days[0].toLocaleDateString(undefined, {month: 'short', day: 'numeric'})}
        {' – '}
        {days[6].toLocaleDateString(undefined, {month: 'short', day: 'numeric', year: 'numeric'})}
      </p>

      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
          <AlertAction>
            <Button type="button" variant="secondary" onClick={() => void loadSchedules()}>
              Try again
            </Button>
          </AlertAction>
        </Alert>
      ) : (
        <AiringAgendaLayout
          days={days}
          schedulesByDay={schedulesByDay}
          loading={loading}
          scrollToTodayRequest={scrollToTodayRequest}
        />
      )}
    </section>
  )
}
