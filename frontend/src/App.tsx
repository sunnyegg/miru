import {useEffect, useState} from 'react'
import {EventsOff, EventsOn} from '../wailsjs/runtime/runtime'
import {InitError} from '../wailsjs/go/main/App'
import {errorMessage} from './lib/format'
import {Sidebar} from './components/Sidebar'
import {LibraryView} from './views/Library'
import {WatchingView} from './views/Watching'
import {SearchView} from './views/Search'
import {DownloadsView} from './views/Downloads'
import {CalendarView} from './views/Calendar'
import {SettingsView} from './views/Settings'
import {Alert} from '@/components/ui/alert'
import {toast} from '@/components/ui/toast'
import type {DownloadView, PlaybackEvent, SyncEvent, TabId} from './lib/types'

export default function App() {
  const [tab, setTab] = useState<TabId>('library')
  const [initError, setInitError] = useState('')
  const [jobs, setJobs] = useState<DownloadView[]>([])
  const [libraryKey, setLibraryKey] = useState(0)
  const [authKey, setAuthKey] = useState(0)
  const [playing, setPlaying] = useState<PlaybackEvent | null>(null)

  function showNotice(text: string, error = false) {
    toast.add({
      title: text,
      type: error ? 'error' : undefined,
      timeout: 4000,
    })
  }

  async function loadInitError() {
    try {
      setInitError(await InitError())
    } catch (err) {
      setInitError(errorMessage(err))
    }
  }

  useEffect(() => {
    void loadInitError()

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
      if (payload.ok) {
        setLibraryKey((count) => count + 1)
      }
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

  return (
    <div className="flex h-full bg-background text-foreground">
      <Sidebar current={tab} onChange={setTab} />
      <div className="flex min-w-0 flex-1 flex-col bg-background">
        {initError && (
          <Alert className="border-0 bg-destructive px-4 py-2 text-sm text-destructive-foreground">
            {initError}
          </Alert>
        )}
        {playing && (
          <div className="border-b border-border bg-bezel px-4 py-2 text-sm" role="status">
            Playing · {Math.round(playing.percent)}%
          </div>
        )}
        <main className={`min-h-0 flex-1 ${tab === 'library' ? 'overflow-hidden' : 'overflow-auto p-6'}`}>
          {tab === 'library' && <LibraryView notice={showNotice} refreshKey={libraryKey} playing={playing} />}
          {tab === 'watching' && <WatchingView refreshKey={authKey} onSettings={() => setTab('settings')} />}
          {tab === 'search' && <SearchView notice={showNotice} onDownloads={() => setTab('downloads')} />}
          {tab === 'downloads' && <DownloadsView notice={showNotice} jobs={jobs} onJobs={setJobs} />}
          {tab === 'calendar' && <CalendarView />}
          {tab === 'settings' && <SettingsView notice={showNotice} refreshKey={authKey} />}
        </main>
      </div>
    </div>
  )
}
