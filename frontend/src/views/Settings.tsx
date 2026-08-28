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
  TestNetworkConnection,
} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {AnilistStatus as Status, SettingsView} from '../lib/types'
import {Alert, AlertAction, AlertDescription} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'
import {Card} from '@/components/ui/card'
import {Input} from '@/components/ui/input'
import {Label} from '@/components/ui/label'
import {NativeSelect, NativeSelectOption} from '@/components/ui/native-select'

const empty: SettingsView = {
  mpvPath: '',
  downloadDir: '',
  syncThreshold: 85,
  downloadRateLimit: 0,
  uploadRateLimit: 0,
  maxConcurrentDownloads: 1,
  networkMode: 'system',
  socks5Address: '127.0.0.1:1080',
}

type Props = {
  notice: (msg: string, isError?: boolean) => void
  refreshKey: number
}

type SettingsSection = 'playback' | 'downloads' | 'network' | 'anilist'

export function SettingsView({notice, refreshKey}: Props) {
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
        downloadDir: settings?.downloadDir ?? '',
        syncThreshold: settings?.syncThreshold || 85,
        downloadRateLimit: bytesToKb(settings?.downloadRateLimit ?? 0),
        uploadRateLimit: bytesToKb(settings?.uploadRateLimit ?? 0),
        maxConcurrentDownloads: settings?.maxConcurrentDownloads ?? 1,
        networkMode: settings?.networkMode ?? 'system',
        socks5Address: settings?.socks5Address ?? '127.0.0.1:1080',
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
      await TestNetworkConnection(form.networkMode, form.socks5Address)
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
        <p className="mt-1 text-sm text-muted-foreground">Playback, downloads, networking, and AniList.</p>
      </header>

      {loadError ? (
        <Alert variant="destructive">
          <AlertDescription>{loadError}</AlertDescription>
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
            void saveSection('playback', () => SavePlaybackSettings(form.mpvPath), 'Playback saved')
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
                  className="min-w-0 flex-1 border-border/40 bg-card"
                />
                <Button type="button" variant="muted" onClick={() => void detectMpv()}>
                  Detect
                </Button>
                <Button type="button" variant="muted" onClick={() => void pickMpvPath()}>
                  Browse
                </Button>
              </div>
            </Field>
            <Button type="submit" disabled={saving === 'playback'} className="mt-4 w-fit">
              {saving === 'playback' ? 'Saving…' : 'Save'}
            </Button>
          </Card>
        </form>

        <form
          onSubmit={(e) => {
            e.preventDefault()
            void saveSection(
              'downloads',
              () =>
                SaveDownloadSettings(
                  form.downloadDir,
                  kbToBytes(form.downloadRateLimit),
                  kbToBytes(form.uploadRateLimit),
                  form.maxConcurrentDownloads,
                ),
              'Downloads saved',
            )
          }}
        >
          <Card>
            <h3 className="text-sm font-medium">Downloads</h3>
            <p className="mt-1 text-sm text-muted-foreground">Folder and speed limits for torrents.</p>
            <Field label="Download folder" htmlFor="downloadDir">
              <div className="flex flex-wrap gap-2">
                <Input
                  id="downloadDir"
                  value={form.downloadDir}
                  onChange={(e) => setForm({...form, downloadDir: e.target.value})}
                  className="min-w-0 flex-1 border-border/40 bg-card"
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
                className="w-32 border-border/40 bg-card"
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
                className="w-32 border-border/40 bg-card"
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
                className="w-32 border-border/40 bg-card"
              />
              <p className="mt-1 text-xs text-muted-foreground">Queued torrents start when a slot is free.</p>
            </Field>
            <Button type="submit" disabled={saving === 'downloads'} className="mt-4 w-fit">
              {saving === 'downloads' ? 'Saving…' : 'Save'}
            </Button>
          </Card>
        </form>

        <form
          onSubmit={(e) => {
            e.preventDefault()
            void saveSection(
              'network',
              () => SaveNetworkSettings(form.networkMode, form.socks5Address),
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
            </NativeSelect>
            {form.networkMode === 'socks5' && (
              <>
                <Label htmlFor="socks5Address" className="mt-4 mb-2">SOCKS5 address</Label>
                <Input
                  id="socks5Address"
                  value={form.socks5Address}
                  onChange={(e) => setForm({...form, socks5Address: e.target.value})}
                  placeholder="127.0.0.1:1080"
                  className="border-border/40 bg-card"
                />
                <p className="mt-1 text-xs text-muted-foreground">
                  Torrent traffic uses TCP through this proxy. UDP, DHT, and inbound peers are disabled.
                </p>
              </>
            )}
            <div className="mt-4 flex flex-wrap gap-2">
              <Button type="submit" disabled={saving === 'network'}>
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
                className="w-32 border-border/40 bg-card"
              />
            </Field>
            <Button type="submit" disabled={saving === 'anilist'} className="mt-4 w-fit">
              {saving === 'anilist' ? 'Saving…' : 'Save'}
            </Button>
          </Card>
        </form>
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
