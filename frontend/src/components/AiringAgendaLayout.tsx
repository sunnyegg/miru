import {useEffect, useRef, useState} from 'react'
import {
  dateKey,
  dayFormatter,
  isToday,
  scheduleTitle,
  timeFormatter,
} from '../lib/calendar'
import type {AiringScheduleView} from '../lib/types'
import {AiringEpisodePoster} from './AiringEpisodePoster'
import {AiringScheduleDialog} from './AiringScheduleDialog'
import {AiringTodayBadge} from './AiringTodayBadge'
import {Skeleton} from '@/components/ui/skeleton'
import {cn} from '@/lib/utils'

function AgendaSkeleton() {
  return (
    <div className="flex flex-col gap-8" aria-busy="true" aria-label="Loading airing schedule">
      {Array.from({length: 2}, (_, sectionIndex) => (
        <div key={sectionIndex}>
          <Skeleton className="h-4 w-44 animate-pulse" />
          <ul className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {Array.from({length: 6}, (_, index) => (
              <li key={index} className="flex items-start gap-3">
                <Skeleton className="h-36 w-24 shrink-0 animate-pulse" />
                <div className="flex flex-1 flex-col gap-2">
                  <Skeleton className="h-3 w-12 animate-pulse" />
                  <Skeleton className="h-3 w-full animate-pulse" />
                  <Skeleton className="h-3 w-1/3 animate-pulse" />
                </div>
              </li>
            ))}
          </ul>
        </div>
      ))}
    </div>
  )
}

type Props = {
  days: Date[]
  schedulesByDay: Map<string, AiringScheduleView[]>
  loading: boolean
  scrollToTodayRequest: number
}

function easeInOutCubic(progress: number): number {
  return progress < 0.5
    ? 4 * progress * progress * progress
    : 1 - Math.pow(-2 * progress + 2, 3) / 2
}

function scrollMainToTodaySection(element: HTMLElement): () => void {
  const scrollContainer = element.closest('main')
  if (!scrollContainer) {
    element.scrollIntoView({behavior: 'smooth', block: 'start'})
    return () => {}
  }

  const container = scrollContainer
  const scrollOffset = 24
  const durationMs = window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 0 : 800

  const containerTop = container.getBoundingClientRect().top
  const elementTop = element.getBoundingClientRect().top
  const targetScrollTop = container.scrollTop + (elementTop - containerTop) - scrollOffset
  const startScrollTop = container.scrollTop
  const scrollDistance = targetScrollTop - startScrollTop

  if (Math.abs(scrollDistance) < 8) {
    return () => {}
  }

  if (durationMs <= 0) {
    container.scrollTop = targetScrollTop
    return () => {}
  }

  let animationFrame = 0
  const startedAt = performance.now()

  function animateScroll(now: number) {
    const elapsed = now - startedAt
    const progress = Math.min(elapsed / durationMs, 1)
    container.scrollTop = startScrollTop + scrollDistance * easeInOutCubic(progress)
    if (progress < 1) {
      animationFrame = requestAnimationFrame(animateScroll)
    }
  }

  animationFrame = requestAnimationFrame(animateScroll)

  return () => {
    if (animationFrame !== 0) {
      cancelAnimationFrame(animationFrame)
    }
  }
}

export function AiringAgendaLayout({days, schedulesByDay, loading, scrollToTodayRequest}: Props) {
  const todaySectionRef = useRef<HTMLElement>(null)
  const [selectedSchedule, setSelectedSchedule] = useState<AiringScheduleView | null>(null)

  useEffect(() => {
    if (loading) {
      return
    }
    const todayVisible = days.some((day) => isToday(day))
    if (!todayVisible || !todaySectionRef.current) {
      return
    }

    let cancelScrollAnimation = () => {}
    let paintFrame = 0
    const layoutFrame = requestAnimationFrame(() => {
      paintFrame = requestAnimationFrame(() => {
        if (!todaySectionRef.current) {
          return
        }
        cancelScrollAnimation = scrollMainToTodaySection(todaySectionRef.current)
      })
    })

    return () => {
      cancelAnimationFrame(layoutFrame)
      if (paintFrame !== 0) {
        cancelAnimationFrame(paintFrame)
      }
      cancelScrollAnimation()
    }
  }, [loading, days, scrollToTodayRequest])

  if (loading) {
    return <AgendaSkeleton />
  }

  return (
    <>
      <div className="flex flex-col gap-8">
      {days.map((day) => {
        const entries = schedulesByDay.get(dateKey(day)) ?? []
        const today = isToday(day)

        return (
          <section
            key={dateKey(day)}
            ref={today ? todaySectionRef : undefined}
            className={cn(
              'scroll-mt-6',
              today && 'border-l-2 border-l-accent bg-muted/40 py-4 pl-4',
            )}
          >
            <div
              className={cn(
                'flex flex-wrap items-baseline gap-2 border-b pb-2',
                today ? 'border-b-accent' : 'border-b-border',
              )}
            >
              <h3 className={cn('text-sm font-medium', today && 'text-accent')}>
                {dayFormatter.format(day)}
              </h3>
              {today && <AiringTodayBadge />}
            </div>
            {entries.length === 0 ? (
              <p className="mt-4 text-sm text-muted-foreground">No episodes scheduled.</p>
            ) : (
              <ul className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
                {entries.map((schedule) => (
                  <li key={schedule.id}>
                    <button
                      type="button"
                      className="flex w-full items-start gap-3 text-left transition-opacity hover:opacity-80 motion-reduce:transition-none"
                      onClick={() => setSelectedSchedule(schedule)}
                    >
                      <AiringEpisodePoster coverImage={schedule.coverImage} size="xxlarge" />
                      <div className="min-w-0 flex-1">
                        <time
                          className="tabular-nums text-xs text-muted-foreground"
                          dateTime={new Date(schedule.airingAt * 1000).toISOString()}
                        >
                          {timeFormatter.format(new Date(schedule.airingAt * 1000))}
                        </time>
                        <p className="mt-0.5 truncate text-sm font-medium">{scheduleTitle(schedule)}</p>
                        <p className="text-xs text-muted-foreground">Episode {schedule.episode}</p>
                      </div>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </section>
        )
      })}
      </div>
      <AiringScheduleDialog
        schedule={selectedSchedule}
        onClose={() => setSelectedSchedule(null)}
      />
    </>
  )
}
