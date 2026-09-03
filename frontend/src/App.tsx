import {useEffect, useRef, useState} from 'react'
import {BrowserOpenURL, EventsOff, EventsOn} from '../wailsjs/runtime/runtime'
import {
  ApplyUpdate,
  AppVersion,
  CheckForUpdate,
  GetSettings,
  InitError,
  SaveLastSeenVersion,
} from '../wailsjs/go/main/App'
import {errorMessage} from './lib/format'
import {Sidebar} from './components/Sidebar'
import {Splash} from './components/Splash'
import {ChangelogDialog} from './components/ChangelogDialog'
import {CloseToTrayDialog} from './components/CloseToTrayDialog'
import {UpdateProgressBar} from './components/UpdateProgressBar'
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
import type {
  DownloadView,
  PlaybackEvent,
  SyncEvent,
  UpdateInfo,
  UpdateProgress,
} from './lib/types'
import {useDownloadStore} from './stores/downloadStore'
import {useFeedStore} from './stores/feedStore'
import {useLibraryStore} from './stores/libraryStore'
import {useNavigationStore} from './stores/navigationStore'
import {usePlaybackStore} from './stores/playbackStore'
import {useSearchStore} from './stores/searchStore'
import {useSettingsStore} from './stores/settingsStore'
import {useWatchingStore} from './stores/watchingStore'

export default function App() {
  const tab = useNavigationStore((state) => state.tab)
  const setTab = useNavigationStore((state) => state.setTab)
  const playing = usePlaybackStore((state) => state.playing)

  const [initError, setInitError] = useState('')
  const [bootDone, setBootDone] = useState(false)
  const [appVersion, setAppVersion] = useState('')
  const [update, setUpdate] = useState<UpdateInfo | null>(null)
  const [showUpdateBanner, setShowUpdateBanner] = useState(true)
  const [checkingUpdate, setCheckingUpdate] = useState(false)
  const [applyingUpdate, setApplyingUpdate] = useState(false)
  const [updateProgress, setUpdateProgress] = useState<UpdateProgress | null>(
    null,
  )
  const [closePromptOpen, setClosePromptOpen] = useState(false)
  const [changelogOpen, setChangelogOpen] = useState(false)
  const mainRef = useRef<HTMLElement>(null)

  useEffect(() => {
    mainRef.current?.focus()
  }, [tab])

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
    setUpdateProgress({downloaded: 0, total: 0, phase: 'downloading'})
    try {
      await ApplyUpdate()
      showNotice('Restarting…')
    } catch (err) {
      showNotice(errorMessage(err), true)
      setApplyingUpdate(false)
      setUpdateProgress(null)
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

  async function markChangelogSeen(version: string) {
    if (!version) {
      return
    }
    try {
      await SaveLastSeenVersion(version)
    } catch (err) {
      showNotice(errorMessage(err), true)
      throw err
    }
  }

  useEffect(() => {
    void loadInitError()
    void loadVersion()
    void (async () => {
      setCheckingUpdate(true)
      try {
        const [info, settings] = await Promise.all([
          CheckForUpdate(),
          GetSettings().catch(() => null),
        ])
        setUpdate(info)
        const seen = settings?.lastSeenVersion ?? '0'
        if (info.available && info.latest && info.latest !== seen) {
          setChangelogOpen(true)
        }
      } catch {
        // startup check is silent
      } finally {
        setCheckingUpdate(false)
      }
    })()

    EventsOn('torrent:progress', (payload: DownloadView) => {
      useDownloadStore.getState().upsertJob(payload)
    })
    EventsOn('library:changed', () => {
      void useLibraryStore.getState().reload(showNotice)
    })
    EventsOn('mpv:progress', (payload: PlaybackEvent) => {
      usePlaybackStore.getState().trackProgress(payload)
    })
    EventsOn('mpv:ended', () => {
      usePlaybackStore.getState().clearPlaying()
    })
    EventsOn('sync:result', (payload: SyncEvent) => {
      showNotice(payload.message, !payload.ok)
      if (payload.ok) {
        void useLibraryStore.getState().reload(showNotice)
      }
    })
    EventsOn('anilist:connected', () => {
      void useLibraryStore.getState().reloadWatching()
      void useWatchingStore.getState().loadList()
      useSettingsStore.getState().bumpReloadKey()
      showNotice('AniList connected')
    })
    EventsOn('feeds:updated', () => {
      void useFeedStore.getState().reload()
    })
    EventsOn('rss:auto_queued', (payload: {count: number}) => {
      const count = payload?.count ?? 1
      const label = count === 1 ? '1 RSS item' : `${count} RSS items`
      showNotice(`Auto-added ${label} for download`)
    })
    EventsOn('window:close-prompt', () => {
      setClosePromptOpen(true)
    })
    EventsOn('update:progress', (payload: UpdateProgress) => {
      setUpdateProgress(payload)
    })

    return () => {
      EventsOff('torrent:progress')
      EventsOff('library:changed')
      EventsOff('mpv:progress')
      EventsOff('mpv:ended')
      EventsOff('sync:result')
      EventsOff('anilist:connected')
      EventsOff('feeds:updated')
      EventsOff('rss:auto_queued')
      EventsOff('window:close-prompt')
      EventsOff('update:progress')
    }
  }, [])

  function openSearchForTorrent(query: string) {
    void useSearchStore.getState().prefillSearch(query, showNotice)
    setTab('search')
  }

  return (
    <TooltipProvider delay={300}>
      <div className="flex h-full bg-background text-foreground">
        <Sidebar />
        <div className="relative flex min-w-0 flex-1 flex-col bg-background">
          {initError && (
            <Alert className="border-0 bg-destructive px-4 py-2 text-sm text-destructive-foreground">
              {friendlyInitError(initError)}
            </Alert>
          )}
          {update?.available && showUpdateBanner && (
            <Alert className="flex flex-col gap-2 border-0 border-b border-border bg-card px-4 py-2 text-sm text-foreground">
              <div className="flex flex-wrap items-center gap-2">
                <span className="min-w-0 flex-1">
                  Miru {update.latest} is available (you have {update.current}).
                </span>
                <Button
                  type="button"
                  disabled={applyingUpdate}
                  onClick={() => void applyUpdate()}
                >
                  {applyingUpdate ? 'Updating…' : 'Update'}
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  onClick={() => setChangelogOpen(true)}
                >
                  {'What\u2019s new'}
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  onClick={() => setShowUpdateBanner(false)}
                >
                  Later
                </Button>
              </div>
              {applyingUpdate && updateProgress && (
                <UpdateProgressBar progress={updateProgress} />
              )}
            </Alert>
          )}
          {playing && (
            <div
              className="border-b border-border bg-bezel px-4 py-2 text-sm"
              role="status"
            >
              Playing · {Math.round(playing.percent)}%
            </div>
          )}
          <main
            ref={mainRef}
            tabIndex={-1}
            className={`min-h-0 flex-1 outline-none ${tab === 'library' ? 'overflow-hidden' : 'overflow-auto p-6'}`}
          >
            {tab === 'library' && (
              <LibraryView
                notice={showNotice}
                onFindTorrent={openSearchForTorrent}
                onReady={() => setBootDone(true)}
              />
            )}
            {tab === 'watching' && <WatchingView notice={showNotice} />}
            {tab === 'search' && <SearchView notice={showNotice} />}
            {tab === 'downloads' && <DownloadsView notice={showNotice} />}
            {tab === 'calendar' && <CalendarView notice={showNotice} />}
            {tab === 'settings' && (
              <SettingsView
                notice={showNotice}
                appVersion={appVersion}
                update={update}
                updateProgress={updateProgress}
                checkingUpdate={checkingUpdate}
                applyingUpdate={applyingUpdate}
                onCheckUpdate={() => void checkUpdate(true)}
                onApplyUpdate={() => void applyUpdate()}
                onOpenRelease={openReleasePage}
                onOpenChangelog={() => setChangelogOpen(true)}
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
        <ChangelogDialog
          open={changelogOpen}
          onOpenChange={setChangelogOpen}
          version={update?.latest ?? ''}
          notes={update?.notes ?? ''}
          releaseUrl={update?.releaseUrl ?? ''}
          notice={showNotice}
          onDismiss={markChangelogSeen}
        />
      </div>
    </TooltipProvider>
  )
}
