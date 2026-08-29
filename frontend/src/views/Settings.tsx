import {useEffect, useState} from 'react'
import {
  AnilistStatus,
  DetectMpv,
  GetSettings,
  LogoutAnilist,
  OpenAnilistLogin,
  PickDownloadDir,
  PickMpvPath,
  SaveAnilistSettings,
  SaveDesktopSettings,
  SaveDownloadSettings,
  SaveNetworkSettings,
  SavePlaybackSettings,
  SaveRSSPollSettings,
  SaveUpdateChannel,
  TestNetworkConnection,
} from '../../wailsjs/go/main/App'
import {SettingsAboutPanel} from '../components/settings/SettingsAboutPanel'
import {SettingsAnilistPanel} from '../components/settings/SettingsAnilistPanel'
import {SettingsDesktopPanel} from '../components/settings/SettingsDesktopPanel'
import {SettingsDownloadsPanel} from '../components/settings/SettingsDownloadsPanel'
import {SettingsNetworkPanel} from '../components/settings/SettingsNetworkPanel'
import {SettingsPlaybackPanel} from '../components/settings/SettingsPlaybackPanel'
import {errorMessage} from '../lib/format'
import type {
  AnilistStatus as Status,
  SettingsView as SettingsForm,
  UpdateInfo,
} from '../lib/types'
import {useSettingsStore, type SettingsTab} from '../stores/settingsStore'
import {cn} from '@/lib/utils'
import {Alert, AlertAction, AlertDescription} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'

const empty: SettingsForm = {
  mpvPath: '',
  anime4kEnabled: false,
  anime4kShadersReady: false,
  downloadDir: '',
  syncThreshold: 85,
  downloadRateLimit: 0,
  uploadRateLimit: 0,
  maxConcurrentDownloads: 1,
  seedRatio: 0.5,
  networkMode: 'system',
  socks5Address: '127.0.0.1:1080',
  httpProxyUrl: 'http://127.0.0.1:8080',
  updateChannel: 'stable',
  rssPollIntervalMinutes: 30,
  discordRpcEnabled: false,
  downloadNotifications: true,
  rssAutoDownload: false,
  rssAutoDownloadLibraryOnly: true,
  closeToTray: false,
}

type Props = {
  notice: (msg: string, isError?: boolean) => void
  appVersion: string
  update: UpdateInfo | null
  checkingUpdate: boolean
  applyingUpdate: boolean
  onCheckUpdate: () => void
  onApplyUpdate: () => void
  onOpenRelease: () => void
}

type SettingsSection =
  | 'desktop'
  | 'playback'
  | 'downloads'
  | 'network'
  | 'anilist'
  | 'updates'

const settingsTabs: {id: SettingsTab; label: string}[] = [
  {id: 'desktop', label: 'Desktop'},
  {id: 'playback', label: 'Playback'},
  {id: 'downloads', label: 'Downloads'},
  {id: 'network', label: 'Network'},
  {id: 'anilist', label: 'AniList'},
  {id: 'about', label: 'About'},
]

export function SettingsView({
  notice,
  appVersion,
  update,
  checkingUpdate,
  applyingUpdate,
  onCheckUpdate,
  onApplyUpdate,
  onOpenRelease,
}: Props) {
  const activeTab = useSettingsStore((state) => state.activeTab)
  const setActiveTab = useSettingsStore((state) => state.setActiveTab)
  const reloadKey = useSettingsStore((state) => state.reloadKey)

  const [form, setForm] = useState<SettingsForm>(empty)
  const [status, setStatus] = useState<Status>({connected: false, username: ''})
  const [saving, setSaving] = useState<SettingsSection | null>(null)
  const [testingNetwork, setTestingNetwork] = useState(false)
  const [loadError, setLoadError] = useState('')

  async function reload() {
    setLoadError('')
    try {
      const [settings, anilist] = await Promise.all([
        GetSettings(),
        AnilistStatus(),
      ])
      setForm({
        mpvPath: settings?.mpvPath ?? '',
        anime4kEnabled: settings?.anime4kEnabled ?? false,
        anime4kShadersReady: settings?.anime4kShadersReady ?? false,
        downloadDir: settings?.downloadDir ?? '',
        syncThreshold: settings?.syncThreshold || 85,
        downloadRateLimit: bytesToKb(settings?.downloadRateLimit ?? 0),
        uploadRateLimit: bytesToKb(settings?.uploadRateLimit ?? 0),
        maxConcurrentDownloads: settings?.maxConcurrentDownloads ?? 1,
        seedRatio: settings?.seedRatio ?? 0.5,
        networkMode: settings?.networkMode ?? 'system',
        socks5Address: settings?.socks5Address ?? '127.0.0.1:1080',
        httpProxyUrl: settings?.httpProxyUrl ?? 'http://127.0.0.1:8080',
        updateChannel: settings?.updateChannel ?? 'stable',
        rssPollIntervalMinutes: settings?.rssPollIntervalMinutes ?? 30,
        discordRpcEnabled: settings?.discordRpcEnabled ?? false,
        downloadNotifications: settings?.downloadNotifications ?? true,
        rssAutoDownload: settings?.rssAutoDownload ?? false,
        rssAutoDownloadLibraryOnly:
          settings?.rssAutoDownloadLibraryOnly ?? true,
        closeToTray: settings?.closeToTray ?? false,
      })
      setStatus(anilist ?? {connected: false, username: ''})
    } catch (err) {
      setLoadError(errorMessage(err))
    }
  }

  useEffect(() => {
    void reload()
  }, [reloadKey])

  async function saveSection(
    section: SettingsSection,
    run: () => Promise<void>,
    ok: string,
  ) {
    setSaving(section)
    try {
      await run()
      notice(ok)
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setSaving(null)
    }
  }

  async function detectMpv() {
    try {
      const path = await DetectMpv()
      setForm((current) => ({...current, mpvPath: path}))
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  async function pickMpvPath() {
    try {
      const path = await PickMpvPath()
      if (!path) {
        return
      }
      setForm((current) => ({...current, mpvPath: path}))
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  async function pickDownloadDir() {
    try {
      const path = await PickDownloadDir()
      if (!path) {
        return
      }
      setForm((current) => ({...current, downloadDir: path}))
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  async function testNetwork() {
    setTestingNetwork(true)
    try {
      await TestNetworkConnection(
        form.networkMode,
        form.socks5Address,
        form.httpProxyUrl,
      )
      notice('Network connection succeeded')
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setTestingNetwork(false)
    }
  }

  async function openAnilistLogin() {
    try {
      await OpenAnilistLogin()
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  async function saveUpdateChannel(channel: string) {
    setForm((current) => ({...current, updateChannel: channel}))
    setSaving('updates')
    try {
      await SaveUpdateChannel(channel)
      onCheckUpdate()
    } catch (err) {
      notice(errorMessage(err), true)
      await reload()
    } finally {
      setSaving(null)
    }
  }

  async function logoutAnilist() {
    try {
      await LogoutAnilist()
      await reload()
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  return (
    <section className="flex h-full flex-col gap-6">
      <header>
        <h2 className="text-2xl font-semibold">Settings</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Playback, downloads, network, AniList, and updates.
        </p>
      </header>

      {loadError ? (
        <Alert variant="destructive">
          <AlertDescription>
            <span className="block font-medium">
              Settings could not be loaded
            </span>
            <span className="mt-1 block">{loadError}</span>
          </AlertDescription>
          <AlertAction>
            <Button
              type="button"
              variant="secondary"
              onClick={() => void reload()}
            >
              Try again
            </Button>
          </AlertAction>
        </Alert>
      ) : (
        <div className="flex min-h-0 flex-1 gap-6">
          <nav
            className="flex w-44 shrink-0 flex-col gap-1 border-r border-border pr-4"
            aria-label="Settings sections"
          >
            {settingsTabs.map((tab) => {
              const active = activeTab === tab.id
              return (
                <Button
                  key={tab.id}
                  type="button"
                  variant="ghost"
                  aria-current={active ? 'page' : undefined}
                  onClick={() => setActiveTab(tab.id)}
                  className={cn(
                    'w-full justify-start border-l px-3 motion-reduce:transition-none',
                    active
                      ? 'border-accent bg-muted text-foreground hover:text-foreground'
                      : 'border-transparent hover:bg-muted hover:text-foreground',
                  )}
                >
                  {tab.label}
                </Button>
              )
            })}
          </nav>

          <div className="min-w-0 flex-1">
            {activeTab === 'desktop' && (
              <SettingsDesktopPanel
                closeToTray={form.closeToTray}
                saving={saving === 'desktop'}
                onCloseToTrayChange={(value) =>
                  setForm((current) => ({...current, closeToTray: value}))
                }
                onSubmit={(event) => {
                  event.preventDefault()
                  void saveSection(
                    'desktop',
                    () => SaveDesktopSettings(form.closeToTray),
                    'Desktop saved',
                  )
                }}
              />
            )}

            {activeTab === 'playback' && (
              <SettingsPlaybackPanel
                form={form}
                setForm={setForm}
                saving={saving === 'playback'}
                onDetectMpv={() => void detectMpv()}
                onPickMpvPath={() => void pickMpvPath()}
                onSubmit={(event) => {
                  event.preventDefault()
                  void saveSection(
                    'playback',
                    async () => {
                      await SavePlaybackSettings(
                        form.mpvPath,
                        form.anime4kEnabled,
                        form.discordRpcEnabled,
                      )
                      await reload()
                    },
                    'Playback saved',
                  )
                }}
              />
            )}

            {activeTab === 'downloads' && (
              <SettingsDownloadsPanel
                form={form}
                setForm={setForm}
                saving={saving === 'downloads'}
                onPickDownloadDir={() => void pickDownloadDir()}
                onSubmit={(event) => {
                  event.preventDefault()
                  void saveSection(
                    'downloads',
                    async () => {
                      await SaveDownloadSettings(
                        form.downloadDir,
                        kbToBytes(form.downloadRateLimit),
                        kbToBytes(form.uploadRateLimit),
                        form.maxConcurrentDownloads,
                        form.seedRatio,
                        form.downloadNotifications,
                        form.rssAutoDownload,
                        form.rssAutoDownloadLibraryOnly,
                      )
                      await SaveRSSPollSettings(form.rssPollIntervalMinutes)
                    },
                    'Downloads saved',
                  )
                }}
              />
            )}

            {activeTab === 'network' && (
              <SettingsNetworkPanel
                form={form}
                setForm={setForm}
                saving={saving === 'network'}
                testingNetwork={testingNetwork}
                onTestNetwork={() => void testNetwork()}
                onSubmit={(event) => {
                  event.preventDefault()
                  void saveSection(
                    'network',
                    () =>
                      SaveNetworkSettings(
                        form.networkMode,
                        form.socks5Address,
                        form.httpProxyUrl,
                      ),
                    'Network saved',
                  )
                }}
              />
            )}

            {activeTab === 'anilist' && (
              <SettingsAnilistPanel
                form={form}
                setForm={setForm}
                status={status}
                saving={saving === 'anilist'}
                onOpenLogin={() => void openAnilistLogin()}
                onLogout={() => void logoutAnilist()}
                onSubmit={(event) => {
                  event.preventDefault()
                  void saveSection(
                    'anilist',
                    () => SaveAnilistSettings(form.syncThreshold),
                    'AniList saved',
                  )
                }}
              />
            )}

            {activeTab === 'about' && (
              <SettingsAboutPanel
                appVersion={appVersion}
                form={form}
                update={update}
                saving={saving === 'updates'}
                checkingUpdate={checkingUpdate}
                applyingUpdate={applyingUpdate}
                onSaveUpdateChannel={(channel) =>
                  void saveUpdateChannel(channel)
                }
                onCheckUpdate={onCheckUpdate}
                onApplyUpdate={onApplyUpdate}
                onOpenRelease={onOpenRelease}
              />
            )}
          </div>
        </div>
      )}
    </section>
  )
}

function bytesToKb(bytes: number) {
  return bytes / 1024
}

function kbToBytes(kb: number) {
  return Math.max(0, Math.round(kb * 1024))
}
