import {useEffect, useState} from 'react'
import {BrowserOpenURL, EventsOff, EventsOn} from '../wailsjs/runtime/runtime'
import {ApplyUpdate, AppVersion, CheckForUpdate, InitError} from '../wailsjs/go/main/App'
import {errorMessage} from './lib/format'
import {Sidebar} from './components/Sidebar'
import {Splash} from './components/Splash'
import {CloseToTrayDialog} from './components/CloseToTrayDialog'
import {LibraryView} from './views/Library'
import {WatchingView} from './views/Watching'
import {SearchView} from './views/Search'
import {DownloadsView} from './views/Downloads'
import {CalendarView} from './views/Calendar'
import {SettingsView} from './views/Settings'
import {Alert} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'
import {TooltipProvider} from '@/components/ui/tooltip'
import {toast} from '@/components/ui/toast'
import type {DownloadView, PlaybackEvent, SyncEvent, TabId, UpdateInfo} from './lib/types'

export default function App() {
  const [tab, setTab] = useState<TabId>('library')
  const [initError, setInitError] = useState('')
  const [jobs, setJobs] = useState<DownloadView[]>([])
  const [libraryKey, setLibraryKey] = useState(0)
  const [authKey, setAuthKey] = useState(0)
  const [playing, setPlaying] = useState<PlaybackEvent | null>(null)
  const [lastPlayback, setLastPlayback] = useState<PlaybackEvent | null>(null)
  const [searchPrefill, setSearchPrefill] = useState('')
  const [bootDone, setBootDone] = useState(false)
  const [appVersion, setAppVersion] = useState('')
  const [update, setUpdate] = useState<UpdateInfo | null>(null)
  const [showUpdateBanner, setShowUpdateBanner] = useState(true)
  const [checkingUpdate, setCheckingUpdate] = useState(false)
  const [applyingUpdate, setApplyingUpdate] = useState(false)
  const [closePromptOpen, setClosePromptOpen] = useState(false)

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

  function friendlyInitError(raw: string): string {
    if (!raw) {
      return ''
    }
    if (raw.startsWith('resolve dirs:')) {
      return 'Miru could not create its data folder. Check the file system permissions and try again.'
    }
    if (raw.startsWith('open database:')) {
      return 'Miru could not open its database. Check the file system permissions and try again.'
    }
    return raw
  }

  async function checkUpdate(manual: boolean) {
    setCheckingUpdate(true)
    try {
      const info = await CheckForUpdate()
      setUpdate(info)
      if (!manual) {
        return
      }
      if (info.available) {
        showNotice(`Miru ${info.latest} is available`)
        return
      }
      if (info.current === 'dev') {
        showNotice('Updates are disabled in development builds')
        return
      }
      showNotice('You are on the latest version')
    } catch (err) {
      if (manual) {
        showNotice(errorMessage(err), true)
      }
    } finally {
      setCheckingUpdate(false)
    }
  }

  async function applyUpdate() {
    setApplyingUpdate(true)
    try {
      await ApplyUpdate()
      showNotice('Restarting…')
    } catch (err) {
      showNotice(errorMessage(err), true)
      setApplyingUpdate(false)
    }
  }

  function openReleasePage() {
    if (update?.releaseUrl) {
      BrowserOpenURL(update.releaseUrl)
    }
  }

  async function loadVersion() {
    try {
      setAppVersion(await AppVersion())
    } catch {
      setAppVersion('')
    }
  }

  useEffect(() => {
    void loadInitError()
    void loadVersion()
    void (async () => {
      setCheckingUpdate(true)
      try {
        setUpdate(await CheckForUpdate())
      } catch {
        // startup check is silent
      } finally {
        setCheckingUpdate(false)
      }
    })()

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
    EventsOn('mpv:progress', (payload: PlaybackEvent) => {
      setPlaying(payload)
      setLastPlayback(payload)
    })
    EventsOn('mpv:ended', () => {
      setPlaying(null)
    })
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
    EventsOn('rss:auto_queued', (payload: {count: number}) => {
      const count = payload?.count ?? 1
      const label = count === 1 ? '1 RSS item' : `${count} RSS items`
      showNotice(`Auto-added ${label} for download`)
    })
    EventsOn('window:close-prompt', () => {
      setClosePromptOpen(true)
    })

    return () => {
      EventsOff('torrent:progress')
      EventsOff('library:changed')
      EventsOff('mpv:progress')
      EventsOff('mpv:ended')
      EventsOff('sync:result')
      EventsOff('anilist:connected')
      EventsOff('rss:auto_queued')
      EventsOff('window:close-prompt')
    }
  }, [])

  function openSearchForTorrent(query: string) {
    setSearchPrefill(query)
    setTab('search')
  }

  return (
    <TooltipProvider delay={300}>
    <div className="flex h-full bg-background text-foreground">
      <Sidebar current={tab} onChange={setTab} />
      <div className="relative flex min-w-0 flex-1 flex-col bg-background">
        {initError && (
          <Alert className="border-0 bg-destructive px-4 py-2 text-sm text-destructive-foreground">
            {friendlyInitError(initError)}
          </Alert>
        )}
        {update?.available && showUpdateBanner && (
          <Alert className="flex flex-wrap items-center gap-2 border-0 border-b border-border bg-card px-4 py-2 text-sm text-foreground">
            <span className="min-w-0 flex-1">
              Miru {update.latest} is available (you have {update.current}).
            </span>
            <Button type="button" disabled={applyingUpdate} onClick={() => void applyUpdate()}>
              {applyingUpdate ? 'Updating…' : 'Update'}
            </Button>
            <Button type="button" variant="ghost" onClick={() => setShowUpdateBanner(false)}>
              Later
            </Button>
          </Alert>
        )}
        {playing && (
          <div className="border-b border-border bg-bezel px-4 py-2 text-sm" role="status">
            Playing · {Math.round(playing.percent)}%
          </div>
        )}
        <main className={`min-h-0 flex-1 ${tab === 'library' ? 'overflow-hidden' : 'overflow-auto p-6'}`}>
          {tab === 'library' && (
            <LibraryView
              notice={showNotice}
              refreshKey={libraryKey}
              authKey={authKey}
              playing={playing}
              lastPlayback={lastPlayback}
              onFindTorrent={openSearchForTorrent}
              onReady={() => setBootDone(true)}
            />
          )}
          {tab === 'watching' && (
            <WatchingView
              refreshKey={authKey}
              notice={showNotice}
              onSettings={() => setTab('settings')}
            />
          )}
          {tab === 'search' && (
            <SearchView
              notice={showNotice}
              onDownloads={() => setTab('downloads')}
              prefillQuery={searchPrefill}
              onPrefillConsumed={() => setSearchPrefill('')}
            />
          )}
          {tab === 'downloads' && <DownloadsView notice={showNotice} jobs={jobs} onJobs={setJobs} />}
          {tab === 'calendar' && <CalendarView />}
          {tab === 'settings' && (
            <SettingsView
              notice={showNotice}
              refreshKey={authKey}
              appVersion={appVersion}
              update={update}
              checkingUpdate={checkingUpdate}
              applyingUpdate={applyingUpdate}
              onCheckUpdate={() => void checkUpdate(true)}
              onApplyUpdate={() => void applyUpdate()}
              onOpenRelease={openReleasePage}
            />
          )}
        </main>
        {tab === 'library' && !bootDone && !initError && (
          <div
            className="pointer-events-none absolute inset-0 z-40 flex items-center justify-center bg-background transition-opacity duration-200 motion-reduce:transition-none"
            aria-hidden="false"
          >
            <Splash />
          </div>
        )}
      </div>
      <CloseToTrayDialog
        open={closePromptOpen}
        onOpenChange={setClosePromptOpen}
        notice={showNotice}
      />
    </div>
    </TooltipProvider>
  )
}
