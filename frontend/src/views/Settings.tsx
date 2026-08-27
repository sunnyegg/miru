import {useEffect, useState, type ReactNode} from 'react'
import {
  AnilistStatus,
  DetectMpv,
  GetSettings,
  LogoutAnilist,
  OpenAnilistLogin,
  PickDownloadDir,
  PickMpvPath,
  SaveSettings,
  TestNetworkConnection,
} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {AnilistStatus as Status, SettingsView} from '../lib/types'
import {Button} from '@/components/ui/button'
import {Card} from '@/components/ui/card'
import {Input} from '@/components/ui/input'
import {Label} from '@/components/ui/label'
import {NativeSelect, NativeSelectOption} from '@/components/ui/native-select'

const empty: SettingsView = {
  mpvPath: '',
  downloadDir: '',
  syncThreshold: 85,
  anilistClientId: '',
  downloadRateLimit: 0,
  uploadRateLimit: 0,
  networkMode: 'system',
  socks5Address: '127.0.0.1:1080',
}

type Props = {
  notice: (msg: string, isError?: boolean) => void
  refreshKey: number
}

export function SettingsView({notice, refreshKey}: Props) {
  const [form, setForm] = useState<SettingsView>(empty)
  const [status, setStatus] = useState<Status>({connected: false, username: ''})
  const [saving, setSaving] = useState(false)
  const [testingNetwork, setTestingNetwork] = useState(false)

  async function reload() {
    try {
      const [settings, anilist] = await Promise.all([GetSettings(), AnilistStatus()])
      setForm({
        mpvPath: settings?.mpvPath ?? '',
        downloadDir: settings?.downloadDir ?? '',
        syncThreshold: settings?.syncThreshold || 85,
        anilistClientId: settings?.anilistClientId ?? '',
        downloadRateLimit: bytesToKb(settings?.downloadRateLimit ?? 0),
        uploadRateLimit: bytesToKb(settings?.uploadRateLimit ?? 0),
        networkMode: settings?.networkMode ?? 'system',
        socks5Address: settings?.socks5Address ?? '127.0.0.1:1080',
      })
      setStatus(anilist ?? {connected: false, username: ''})
    } catch (err) {
      notice(errorMessage(err), true)
    }
  }

  useEffect(() => {
    void reload()
  }, [refreshKey])

  async function save() {
    setSaving(true)
    try {
      await SaveSettings({
        ...form,
        downloadRateLimit: kbToBytes(form.downloadRateLimit),
        uploadRateLimit: kbToBytes(form.uploadRateLimit),
      })
      notice('Settings saved')
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setSaving(false)
    }
  }

  return (
    <section className="flex h-full flex-col gap-6">
      <header>
        <h2 className="text-2xl font-semibold">Settings</h2>
        <p className="mt-1 text-sm text-muted-foreground">Paths, AniList login, and sync threshold.</p>
      </header>

      <form
        className="flex max-w-3xl flex-col gap-5"
        onSubmit={(e) => {
          e.preventDefault()
          void save()
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
          <Button
            type="button"
            variant="muted"
            className="mt-4"
            disabled={testingNetwork}
            onClick={() => {
              setTestingNetwork(true)
              void TestNetworkConnection(form.networkMode, form.socks5Address)
                .then(() => notice('Network connection succeeded'))
                .catch((err) => notice(errorMessage(err), true))
                .finally(() => setTestingNetwork(false))
            }}
          >
            {testingNetwork ? 'Testing…' : 'Test connection'}
          </Button>
        </Card>

        <Field label="MPV path" htmlFor="mpvPath">
          <div className="flex flex-wrap gap-2">
            <Input
              id="mpvPath"
              value={form.mpvPath}
              onChange={(e) => setForm({...form, mpvPath: e.target.value})}
              className="min-w-0 flex-1 border-border/40 bg-card"
            />
            <Button
              type="button"
              variant="muted"
              onClick={() => DetectMpv().then((p) => setForm({...form, mpvPath: p})).catch((err) => notice(errorMessage(err), true))}
            >
              Detect
            </Button>
            <Button
              type="button"
              variant="muted"
              onClick={() => PickMpvPath().then((p) => p && setForm({...form, mpvPath: p}))}
            >
              Browse
            </Button>
          </div>
        </Field>

        <Field label="Download folder" htmlFor="downloadDir">
          <div className="flex flex-wrap gap-2">
            <Input
              id="downloadDir"
              value={form.downloadDir}
              onChange={(e) => setForm({...form, downloadDir: e.target.value})}
              className="min-w-0 flex-1 border-border/40 bg-card"
            />
            <Button
              type="button"
              variant="muted"
              onClick={() => PickDownloadDir().then((p) => p && setForm({...form, downloadDir: p}))}
            >
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

        <Field label="AniList sync threshold (%)" htmlFor="threshold">
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

        <Card>
          <h3 className="text-sm font-medium">AniList account</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            {status.connected ? `Connected as ${status.username}` : 'Not connected. Open login, then authorize in the browser.'}
          </p>
          <div className="mt-3 flex flex-wrap gap-2">
            <Button
              type="button"
              variant="secondary"
              onClick={() => OpenAnilistLogin().catch((err) => notice(errorMessage(err), true))}
            >
              Open login
            </Button>
            {status.connected && (
              <Button
                type="button"
                variant="ghost"
                className="text-destructive hover:text-destructive"
                onClick={() => LogoutAnilist().then(() => reload()).catch((err) => notice(errorMessage(err), true))}
              >
                Log out
              </Button>
            )}
          </div>
        </Card>

        <Button type="submit" disabled={saving} className="w-fit">
          {saving ? 'Saving…' : 'Save settings'}
        </Button>
      </form>
    </section>
  )
}

function Field({label, htmlFor, children}: {label: string; htmlFor: string; children: ReactNode}) {
  return (
    <div>
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
