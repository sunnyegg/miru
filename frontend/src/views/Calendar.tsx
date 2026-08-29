import {useEffect, useMemo} from 'react'
import {AiringAgendaLayout} from '../components/AiringAgendaLayout'
import {
  buildWeekDays,
  groupSchedulesByDay,
  startOfMonday,
} from '../lib/calendar'
import {useCalendarStore} from '../stores/calendarStore'
import {Alert, AlertAction, AlertDescription} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'

export function CalendarView() {
  const weekOffset = useCalendarStore((state) => state.weekOffset)
  const scrollToTodayRequest = useCalendarStore(
    (state) => state.scrollToTodayRequest,
  )
  const schedules = useCalendarStore((state) => state.schedules)
  const loading = useCalendarStore((state) => state.loading)
  const error = useCalendarStore((state) => state.error)
  const setWeekOffset = useCalendarStore((state) => state.setWeekOffset)
  const goToToday = useCalendarStore((state) => state.goToToday)
  const loadSchedules = useCalendarStore((state) => state.loadSchedules)

  const weekStart = useMemo(() => {
    const start = startOfMonday(new Date())
    start.setDate(start.getDate() + weekOffset * 7)
    return start
  }, [weekOffset])

  const days = useMemo(() => buildWeekDays(weekStart), [weekStart])

  const schedulesByDay = useMemo(
    () => groupSchedulesByDay(schedules),
    [schedules],
  )

  useEffect(() => {
    void loadSchedules()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [weekOffset])

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
          <Button
            type="button"
            variant="muted"
            onClick={() => setWeekOffset((offset) => offset - 1)}
          >
            Previous
          </Button>
          <Button type="button" variant="secondary" onClick={() => goToToday()}>
            Today
          </Button>
          <Button
            type="button"
            variant="muted"
            onClick={() => setWeekOffset((offset) => offset + 1)}
          >
            Next
          </Button>
        </div>
      </header>

      <p className="text-sm font-medium text-muted-foreground">
        {days[0].toLocaleDateString(undefined, {
          month: 'short',
          day: 'numeric',
        })}
        {' – '}
        {days[6].toLocaleDateString(undefined, {
          month: 'short',
          day: 'numeric',
          year: 'numeric',
        })}
      </p>

      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
          <AlertAction>
            <Button
              type="button"
              variant="secondary"
              onClick={() => void loadSchedules()}
            >
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
