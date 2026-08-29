import {useEffect, useState, type ReactNode} from 'react'
import {
  AnilistStatus,
  DetectMpv,
  GetSettings,
  LogoutAnilist,
  OpenAnilistLogin,
  PickDownloadDir,
  PickMpvPath,
  SaveAnilistSettings,
  SaveDownloadSettings,
  SaveNetworkSettings,
  SavePlaybackSettings,
  SaveRSSPollSettings,
  SaveUpdateChannel,
  TestNetworkConnection,
} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {AnilistStatus as Status, SettingsView, UpdateInfo} from '../lib/types'
import {Alert, AlertAction, AlertDescription} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'
import {Card} from '@/components/ui/card'
import {Input} from '@/components/ui/input'
import {Label} from '@/components/ui/label'
import {NativeSelect, NativeSelectOption} from '@/components/ui/native-select'

const empty: SettingsView = {
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
  discordAppId: '',
  downloadNotifications: true,
  rssAutoDownload: false,
  rssAutoDownloadLibraryOnly: true,
}

type Props = {
  notice: (msg: string, isError?: boolean) => void
  refreshKey: number
  appVersion: string
  update: UpdateInfo | null
  checkingUpdate: boolean
  applyingUpdate: boolean
  onCheckUpdate: () => void
  onApplyUpdate: () => void
  onOpenRelease: () => void
}

type SettingsSection = 'playback' | 'downloads' | 'network' | 'anilist' | 'updates'

export function SettingsView({
  notice,
  refreshKey,
  appVersion,
  update,
  checkingUpdate,
  applyingUpdate,
  onCheckUpdate,
  onApplyUpdate,
  onOpenRelease,
}: Props) {
  const [form, setForm] = useState<SettingsView>(empty)
  const [status, setStatus] = useState<Status>({connected: false, username: ''})
  const [saving, setSaving] = useState<SettingsSection | null>(null)
  const [testingNetwork, setTestingNetwork] = useState(false)
  const [loadError, setLoadError] = useState('')

  async function reload() {
    setLoadError('')
    try {
      const [settings, anilist] = await Promise.all([GetSettings(), AnilistStatus()])
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
        discordAppId: settings?.discordAppId ?? '',
        downloadNotifications: settings?.downloadNotifications ?? true,
        rssAutoDownload: settings?.rssAutoDownload ?? false,
        rssAutoDownloadLibraryOnly: settings?.rssAutoDownloadLibraryOnly ?? true,
      })
      setStatus(anilist ?? {connected: false, username: ''})
    } catch (err) {
      setLoadError(errorMessage(err))
    }
  }

  useEffect(() => {
    void reload()
  }, [refreshKey])

  async function saveSection(section: SettingsSection, run: () => Promise<void>, ok: string) {
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
      await TestNetworkConnection(form.networkMode, form.socks5Address, form.httpProxyUrl)
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
        <p className="mt-1 text-sm text-muted-foreground">Playback, downloads, networking, AniList, and updates.</p>
      </header>

      {loadError ? (
        <Alert variant="destructive">
          <AlertDescription>
            <span className="block font-medium">Settings could not be loaded</span>
            <span className="mt-1 block">{loadError}</span>
          </AlertDescription>
          <AlertAction>
            <Button type="button" variant="secondary" onClick={() => void reload()}>
              Try again
            </Button>
          </AlertAction>
        </Alert>
      ) : (
        <div className="flex max-w-3xl flex-col gap-5">
        <form
          onSubmit={(e) => {
            e.preventDefault()
            void saveSection(
              'playback',
              async () => {
                await SavePlaybackSettings(
                  form.mpvPath,
                  form.anime4kEnabled,
                  form.discordRpcEnabled,
                  form.discordAppId,
                )
                await reload()
              },
              'Playback saved',
            )
          }}
        >
          <Card>
            <h3 className="text-sm font-medium">Playback</h3>
            <p className="mt-1 text-sm text-muted-foreground">MPV binary used to play episodes.</p>
            <Field label="MPV path" htmlFor="mpvPath">
              <div className="flex flex-wrap gap-2">
                <Input
                  id="mpvPath"
                  value={form.mpvPath}
                  onChange={(e) => setForm({...form, mpvPath: e.target.value})}
                  className="min-w-0 flex-1 bg-card"
                />
                <Button type="button" variant="muted" onClick={() => void detectMpv()}>
                  Detect
                </Button>
                <Button type="button" variant="muted" onClick={() => void pickMpvPath()}>
                  Browse
                </Button>
              </div>
            </Field>
            <div className="mt-4">
              <label className="flex min-h-11 cursor-pointer items-center gap-3">
                <input
                  type="checkbox"
                  checked={form.anime4kEnabled}
                  onChange={(e) => setForm({...form, anime4kEnabled: e.target.checked})}
                  className="size-4 accent-primary"
                />
                <span className="text-sm">Enable Anime4K upscaling</span>
              </label>
              <p className="mt-1 text-xs text-muted-foreground">
                Applies Anime4K Mode A shaders when MPV starts. Shaders are cached in your Miru config folder.
              </p>
              {form.anime4kEnabled && !form.anime4kShadersReady && (
                <Alert variant="destructive" className="mt-3">
                  <AlertDescription>
                    Anime4K shaders are not installed yet. Save playback settings to download them.
                  </AlertDescription>
                </Alert>
              )}
              {form.anime4kEnabled && form.anime4kShadersReady && (
                <p className="mt-2 text-xs text-muted-foreground">Anime4K shaders are installed.</p>
              )}
            </div>
            <div className="mt-4 flex items-start gap-3">
              <input
                id="discordRpcEnabled"
                type="checkbox"
                checked={form.discordRpcEnabled}
                onChange={(e) => setForm({...form, discordRpcEnabled: e.target.checked})}
                className="mt-1 size-4 shrink-0 accent-primary"
              />
              <div>
                <Label htmlFor="discordRpcEnabled">Discord Rich Presence</Label>
                <p className="mt-1 text-sm text-muted-foreground">
                  Show the anime you are watching on your Discord profile while MPV is playing.
                </p>
              </div>
            </div>
            {form.discordRpcEnabled && (
              <Field label="Discord application ID" htmlFor="discordAppId">
                <Input
                  id="discordAppId"
                  value={form.discordAppId}
                  onChange={(e) => setForm({...form, discordAppId: e.target.value})}
                  placeholder="From the Discord Developer Portal"
                  className="bg-card"
                />
                <p className="mt-1 text-xs text-muted-foreground">
                  Leave empty to use DISCORD_APP_ID from the build .env file.
                </p>
              </Field>
            )}
            <Button type="submit" disabled={saving === 'playback'} className="mt-4 w-fit">
              {saving === 'playback' ? 'Saving…' : 'Save'}
            </Button>
            <p className="mt-2 text-xs text-muted-foreground">
              Discord Rich Presence requires the Discord desktop app to be running. MPV path changes take effect after restart.
            </p>
          </Card>
        </form>

        <form
          onSubmit={(e) => {
            e.preventDefault()
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
        >
          <Card>
            <h3 className="text-sm font-medium">Downloads</h3>
            <p className="mt-1 text-sm text-muted-foreground">Folder, speed limits, queue, seeding, and RSS auto-download.</p>
            <Field label="Download folder" htmlFor="downloadDir">
              <div className="flex flex-wrap gap-2">
                <Input
                  id="downloadDir"
                  value={form.downloadDir}
                  onChange={(e) => setForm({...form, downloadDir: e.target.value})}
                  className="min-w-0 flex-1 bg-card"
                />
                <Button type="button" variant="muted" onClick={() => void pickDownloadDir()}>
                  Browse
                </Button>
              </div>
            </Field>
            <Field label="Download speed limit (KB/s)" htmlFor="downloadRateLimit">
              <Input
                id="downloadRateLimit"
                type="number"
                min={0}
                step={1}
                value={form.downloadRateLimit}
                onChange={(e) => setForm({...form, downloadRateLimit: Number(e.target.value)})}
                  className="w-32 bg-card"
              />
              <p className="mt-1 text-xs text-muted-foreground">0 = unlimited</p>
            </Field>
            <Field label="Upload speed limit (KB/s)" htmlFor="uploadRateLimit">
              <Input
                id="uploadRateLimit"
                type="number"
                min={0}
                step={1}
                value={form.uploadRateLimit}
                onChange={(e) => setForm({...form, uploadRateLimit: Number(e.target.value)})}
                  className="w-32 bg-card"
              />
              <p className="mt-1 text-xs text-muted-foreground">0 = unlimited</p>
            </Field>
            <Field label="Max concurrent downloads" htmlFor="maxConcurrentDownloads">
              <Input
                id="maxConcurrentDownloads"
                type="number"
                min={1}
                max={8}
                step={1}
                value={form.maxConcurrentDownloads}
                onChange={(e) => setForm({...form, maxConcurrentDownloads: Number(e.target.value)})}
                  className="w-32 bg-card"
              />
              <p className="mt-1 text-xs text-muted-foreground">Queued torrents start when a slot is free.</p>
            </Field>
            <Field label="Seed ratio" htmlFor="seedRatio">
              <Input
                id="seedRatio"
                type="number"
                min={0}
                max={10}
                step={0.1}
                value={form.seedRatio}
                onChange={(e) => setForm({...form, seedRatio: Number(e.target.value)})}
                className="w-32 bg-card"
              />
              <p className="mt-1 text-xs text-muted-foreground">
                Upload ratio before auto-finish (0.5 = half the download size). 0 stops seeding right away.
              </p>
            </Field>
            <Field label="RSS poll interval (minutes)" htmlFor="rssPollIntervalMinutes">
              <Input
                id="rssPollIntervalMinutes"
                type="number"
                min={5}
                max={1440}
                step={5}
                value={form.rssPollIntervalMinutes}
                onChange={(e) => setForm({...form, rssPollIntervalMinutes: Number(e.target.value)})}
                className="w-32 bg-card"
              />
              <p className="mt-1 text-xs text-muted-foreground">
                How often subscribed RSS feeds are checked in the background (5–1440).
              </p>
            </Field>
            <div className="mt-4 flex items-start gap-3">
              <input
                id="rssAutoDownload"
                type="checkbox"
                checked={form.rssAutoDownload}
                onChange={(event) =>
                  setForm((current) => ({...current, rssAutoDownload: event.target.checked}))
                }
                className="mt-1 size-4 shrink-0 accent-primary"
              />
              <div>
                <Label htmlFor="rssAutoDownload">Auto-download new RSS items</Label>
                <p className="mt-1 text-xs text-muted-foreground">
                  Queue torrent downloads when subscribed feeds publish new items with magnet links.
                </p>
              </div>
            </div>
            {form.rssAutoDownload && (
              <div className="mt-4 flex items-start gap-3">
                <input
                  id="rssAutoDownloadLibraryOnly"
                  type="checkbox"
                  checked={form.rssAutoDownloadLibraryOnly}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      rssAutoDownloadLibraryOnly: event.target.checked,
                    }))
                  }
                  className="mt-1 size-4 shrink-0 accent-primary"
                />
                <div>
                  <Label htmlFor="rssAutoDownloadLibraryOnly">Only your library</Label>
                  <p className="mt-1 text-xs text-muted-foreground">
                    Only auto-download when the item title matches an anime in your local library.
                  </p>
                </div>
              </div>
            )}
            <div className="mt-4 flex items-start gap-3">
              <input
                id="downloadNotifications"
                type="checkbox"
                checked={form.downloadNotifications}
                onChange={(event) =>
                  setForm((current) => ({...current, downloadNotifications: event.target.checked}))
                }
                className="mt-1 size-4 shrink-0 accent-primary"
              />
              <div>
                <Label htmlFor="downloadNotifications">Desktop notifications</Label>
                <p className="mt-1 text-xs text-muted-foreground">
                  Show an OS notification when a download you started finishes in the background.
                </p>
              </div>
            </div>
            <Button type="submit" variant="secondary" disabled={saving === 'downloads'} className="mt-4 w-fit">
              {saving === 'downloads' ? 'Saving…' : 'Save'}
            </Button>
          </Card>
        </form>

        <form
          onSubmit={(e) => {
            e.preventDefault()
            void saveSection(
              'network',
              () => SaveNetworkSettings(form.networkMode, form.socks5Address, form.httpProxyUrl),
              'Networking saved',
            )
          }}
        >
          <Card>
            <h3 className="text-sm font-medium">Networking</h3>
            <p className="mt-1 text-sm text-muted-foreground">
              Choose how Miru connects to AniList, Nyaa, and downloads.
            </p>
            <Label htmlFor="networkMode" className="mt-4 mb-2">Connection mode</Label>
            <NativeSelect
              id="networkMode"
              value={form.networkMode}
              onChange={(e) => setForm({...form, networkMode: e.target.value})}
            >
              <NativeSelectOption value="system">System proxy / VPN</NativeSelectOption>
              <NativeSelectOption value="direct">Direct connection</NativeSelectOption>
              <NativeSelectOption value="socks5">SOCKS5 proxy</NativeSelectOption>
              <NativeSelectOption value="http_proxy">HTTP/HTTPS proxy</NativeSelectOption>
            </NativeSelect>
            {form.networkMode === 'socks5' && (
              <>
                <Label htmlFor="socks5Address" className="mt-4 mb-2">SOCKS5 address</Label>
                <Input
                  id="socks5Address"
                  value={form.socks5Address}
                  onChange={(e) => setForm({...form, socks5Address: e.target.value})}
                  placeholder="127.0.0.1:1080"
                  className="bg-card"
                />
                <p className="mt-1 text-xs text-muted-foreground">
                  Torrent traffic uses TCP through this proxy. UDP, DHT, and inbound peers are disabled.
                </p>
              </>
            )}
            {form.networkMode === 'http_proxy' && (
              <>
                <Label htmlFor="httpProxyUrl" className="mt-4 mb-2">Proxy URL</Label>
                <Input
                  id="httpProxyUrl"
                  value={form.httpProxyUrl}
                  onChange={(e) => setForm({...form, httpProxyUrl: e.target.value})}
                  placeholder="http://127.0.0.1:8080"
                  className="bg-card"
                />
                <p className="mt-1 text-xs text-muted-foreground">
                  HTTP and HTTPS traffic routes through this proxy. Use http:// or https:// with host and port.
                </p>
              </>
            )}
            <div className="mt-4 flex flex-wrap gap-2">
              <Button type="submit" variant="secondary" disabled={saving === 'network'}>
                {saving === 'network' ? 'Saving…' : 'Save'}
              </Button>
              <Button type="button" variant="muted" disabled={testingNetwork} onClick={() => void testNetwork()}>
                {testingNetwork ? 'Testing…' : 'Test connection'}
              </Button>
            </div>
          </Card>
        </form>

        <form
          onSubmit={(e) => {
            e.preventDefault()
            void saveSection('anilist', () => SaveAnilistSettings(form.syncThreshold), 'AniList saved')
          }}
        >
          <Card>
            <h3 className="text-sm font-medium">AniList</h3>
            <p className="mt-1 text-sm text-muted-foreground">
              {status.connected ? `Connected as ${status.username}` : 'Not connected. Open login, then authorize in the browser.'}
            </p>
            <div className="mt-3 flex flex-wrap gap-2">
              <Button type="button" variant="secondary" onClick={() => void openAnilistLogin()}>
                Open login
              </Button>
              {status.connected && (
                <Button
                  type="button"
                  variant="ghost"
                  className="text-destructive hover:text-destructive"
                  onClick={() => void logoutAnilist()}
                >
                  Log out
                </Button>
              )}
            </div>
            <Field label="Sync threshold (%)" htmlFor="threshold">
              <Input
                id="threshold"
                type="number"
                min={1}
                max={100}
                value={form.syncThreshold}
                onChange={(e) => setForm({...form, syncThreshold: Number(e.target.value)})}
                  className="w-32 bg-card"
              />
            </Field>
            <Button type="submit" variant="secondary" disabled={saving === 'anilist'} className="mt-4 w-fit">
              {saving === 'anilist' ? 'Saving…' : 'Save'}
            </Button>
          </Card>
        </form>

        <Card>
          <h3 className="text-sm font-medium">About</h3>
          <p className="mt-1 text-sm text-muted-foreground">Version {appVersion || 'dev'}</p>
          <Label htmlFor="updateChannel" className="mt-4 mb-2">Update channel</Label>
          <NativeSelect
            id="updateChannel"
            value={form.updateChannel}
            disabled={saving === 'updates' || checkingUpdate}
            onChange={(e) => void saveUpdateChannel(e.target.value)}
          >
            <NativeSelectOption value="stable">Stable</NativeSelectOption>
            <NativeSelectOption value="prerelease">Prerelease</NativeSelectOption>
          </NativeSelect>
          <p className="mt-1 text-xs text-muted-foreground">
            Prerelease includes alpha and beta builds. Stable ignores them.
          </p>
          {update?.available ? (
            <>
              <p className="mt-3 text-sm">Miru {update.latest} is available.</p>
              <div className="mt-3 flex flex-wrap gap-2">
                <Button type="button" disabled={applyingUpdate} onClick={onApplyUpdate}>
                  {applyingUpdate ? 'Updating…' : 'Update now'}
                </Button>
                <Button type="button" variant="muted" onClick={onOpenRelease}>
                  Open download page
                </Button>
              </div>
            </>
          ) : (
            <Button
              type="button"
              variant="secondary"
              className="mt-4 w-fit"
              disabled={checkingUpdate}
              onClick={onCheckUpdate}
            >
              {checkingUpdate ? 'Checking…' : 'Check for updates'}
            </Button>
          )}
        </Card>
        </div>
      )}
    </section>
  )
}

function Field({label, htmlFor, children}: {label: string; htmlFor: string; children: ReactNode}) {
  return (
    <div className="mt-4">
      <Label htmlFor={htmlFor} className="mb-2">{label}</Label>
      {children}
    </div>
  )
}

function bytesToKb(bytes: number) {
  return bytes / 1024
}

function kbToBytes(kb: number) {
  return Math.max(0, Math.round(kb * 1024))
}
