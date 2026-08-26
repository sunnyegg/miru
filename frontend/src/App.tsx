import {useEffect, useState} from 'react'
import {EventsOff, EventsOn} from '../wailsjs/runtime/runtime'
import {InitError} from '../wailsjs/go/main/App'
import {Sidebar} from './components/Sidebar'
import {LibraryView} from './views/Library'
import {WatchingView} from './views/Watching'
import {DownloadsView} from './views/Downloads'
import {CalendarView} from './views/Calendar'
import {SettingsView} from './views/Settings'
import type {DownloadView, PlaybackEvent, SyncEvent, TabId} from './lib/types'

export default function App() {
  const [tab, setTab] = useState<TabId>('library')
  const [notice, setNotice] = useState<{text: string; error: boolean} | null>(null)
  const [initError, setInitError] = useState('')
  const [jobs, setJobs] = useState<DownloadView[]>([])
  const [libraryKey, setLibraryKey] = useState(0)
  const [authKey, setAuthKey] = useState(0)
  const [playing, setPlaying] = useState<PlaybackEvent | null>(null)

  function showNotice(text: string, error = false) {
    setNotice({text, error})
  }

  useEffect(() => {
    void InitError().then(setInitError)

    EventsOn('torrent:progress', (payload: DownloadView) => {
      setJobs((current) => {
        const exists = current.some((job) => job.id === payload.id)
        if (!exists) {
          return [payload, ...current]
        }
        return current.map((job) => job.id === payload.id ? payload : job)
      })
    })
    EventsOn('library:changed', () => setLibraryKey((n) => n + 1))
    EventsOn('mpv:progress', (payload: PlaybackEvent) => setPlaying(payload))
    EventsOn('mpv:ended', () => setPlaying(null))
    EventsOn('sync:result', (payload: SyncEvent) => {
      showNotice(payload.message, !payload.ok)
    })
    EventsOn('anilist:connected', () => {
      setAuthKey((n) => n + 1)
      showNotice('AniList connected')
    })

    return () => {
      EventsOff('torrent:progress')
      EventsOff('library:changed')
      EventsOff('mpv:progress')
      EventsOff('mpv:ended')
      EventsOff('sync:result')
      EventsOff('anilist:connected')
    }
  }, [])

  useEffect(() => {
    if (!notice) {
      return
    }
    const id = window.setTimeout(() => setNotice(null), 4000)
    return () => window.clearTimeout(id)
  }, [notice])

  return (
    <div className="flex h-full bg-background text-foreground">
      <Sidebar current={tab} onChange={setTab} />
      <div className="flex min-w-0 flex-1 flex-col">
        {initError && (
          <div className="bg-destructive px-4 py-2 text-sm text-on-destructive" role="alert">
            {initError}
          </div>
        )}
        {playing && (
          <div className="bg-secondary px-4 py-2 text-sm text-on-secondary" role="status">
            Playing · {Math.round(playing.percent)}%
          </div>
        )}
        <main className="min-h-0 flex-1 overflow-auto p-6">
          {tab === 'library' && <LibraryView notice={showNotice} refreshKey={libraryKey} />}
          {tab === 'watching' && <WatchingView notice={showNotice} refreshKey={authKey} onSettings={() => setTab('settings')} />}
          {tab === 'downloads' && <DownloadsView notice={showNotice} jobs={jobs} onJobs={setJobs} />}
          {tab === 'calendar' && <CalendarView notice={showNotice} />}
          {tab === 'settings' && <SettingsView notice={showNotice} refreshKey={authKey} />}
        </main>
      </div>
      {notice && (
        <div
          role="status"
          className={`fixed bottom-4 right-4 z-50 max-w-sm rounded-lg px-4 py-3 text-sm shadow-lg ${
            notice.error ? 'bg-destructive text-on-destructive' : 'bg-secondary text-on-secondary'
          }`}
        >
          {notice.text}
        </div>
      )}
    </div>
  )
}
