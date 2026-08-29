import {create} from 'zustand'
import {ListAiringSchedule} from '../../wailsjs/go/main/App'
import {buildWeekDays, startOfMonday} from '../lib/calendar'
import {errorMessage} from '../lib/format'
import type {AiringScheduleView} from '../lib/types'

type CalendarState = {
  weekOffset: number
  scrollToTodayRequest: number
  schedules: AiringScheduleView[]
  loading: boolean
  error: string
  setWeekOffset: (offset: number | ((current: number) => number)) => void
  goToToday: () => void
  loadSchedules: () => Promise<void>
}

function weekStartForOffset(weekOffset: number): Date {
  const start = startOfMonday(new Date())
  start.setDate(start.getDate() + weekOffset * 7)
  return start
}

export const useCalendarStore = create<CalendarState>((set, get) => ({
  weekOffset: 0,
  scrollToTodayRequest: 0,
  schedules: [],
  loading: true,
  error: '',

  setWeekOffset: (offset) => {
    set((state) => ({
      weekOffset:
        typeof offset === 'function' ? offset(state.weekOffset) : offset,
    }))
  },

  goToToday: () => {
    set((state) => ({
      weekOffset: 0,
      scrollToTodayRequest: state.scrollToTodayRequest + 1,
    }))
  },

  loadSchedules: async () => {
    set({loading: true, error: ''})
    try {
      const weekStart = weekStartForOffset(get().weekOffset)
      const days = buildWeekDays(weekStart)
      const start = Math.floor(weekStart.getTime() / 1000)
      const end = Math.floor(days[6].getTime() / 1000) + 24 * 60 * 60
      const result = await ListAiringSchedule(Math.max(0, start - 1), end)
      set({schedules: result ?? []})
    } catch (err) {
      set({error: errorMessage(err)})
    } finally {
      set({loading: false})
    }
  },
}))
