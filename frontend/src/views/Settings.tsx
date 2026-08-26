import {useEffect, useState, type ReactNode} from 'react'
import {
  AnilistStatus,
  DetectMpv,
  GetSettings,
  LogoutAnilist,
  OpenAnilistLogin,
  PickDownloadDir,
  PickMpvPath,
  SaveAnilistToken,
  SaveSettings,
} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import type {AnilistStatus as Status, SettingsView} from '../lib/types'

const empty: SettingsView = {
  mpvPath: '',
  downloadDir: '',
  syncThreshold: 85,
  anilistClientId: '',
  downloadRateLimit: 0,
  uploadRateLimit: 0,
}

type Props = {
  notice: (msg: string, isError?: boolean) => void
  refreshKey: number
}

export function SettingsView({notice, refreshKey}: Props) {
  const [form, setForm] = useState<SettingsView>(empty)
  const [token, setToken] = useState('')
  const [status, setStatus] = useState<Status>({connected: false, username: ''})
  const [saving, setSaving] = useState(false)

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
        <Field label="MPV path" htmlFor="mpvPath">
          <div className="flex flex-wrap gap-2">
            <input
              id="mpvPath"
              value={form.mpvPath}
              onChange={(e) => setForm({...form, mpvPath: e.target.value})}
              className="min-h-11 min-w-0 flex-1 rounded-lg border border-border/40 bg-card px-3 text-sm"
            />
            <button type="button" className="min-h-11 cursor-pointer rounded-lg bg-muted px-3 text-sm" onClick={() => DetectMpv().then((p) => setForm({...form, mpvPath: p})).catch((err) => notice(errorMessage(err), true))}>
              Detect
            </button>
            <button type="button" className="min-h-11 cursor-pointer rounded-lg bg-muted px-3 text-sm" onClick={() => PickMpvPath().then((p) => p && setForm({...form, mpvPath: p}))}>
              Browse
            </button>
          </div>
        </Field>

        <Field label="Download folder" htmlFor="downloadDir">
          <div className="flex flex-wrap gap-2">
            <input
              id="downloadDir"
              value={form.downloadDir}
              onChange={(e) => setForm({...form, downloadDir: e.target.value})}
              className="min-h-11 min-w-0 flex-1 rounded-lg border border-border/40 bg-card px-3 text-sm"
            />
            <button type="button" className="min-h-11 cursor-pointer rounded-lg bg-muted px-3 text-sm" onClick={() => PickDownloadDir().then((p) => p && setForm({...form, downloadDir: p}))}>
              Browse
            </button>
          </div>
        </Field>

        <Field label="Download speed limit (KB/s)" htmlFor="downloadRateLimit">
          <input
            id="downloadRateLimit"
            type="number"
            min={0}
            step={1}
            value={form.downloadRateLimit}
            onChange={(e) => setForm({...form, downloadRateLimit: Number(e.target.value)})}
            className="min-h-11 w-32 rounded-lg border border-border/40 bg-card px-3 text-sm"
          />
          <p className="mt-1 text-xs text-muted-foreground">0 = unlimited</p>
        </Field>

        <Field label="Upload speed limit (KB/s)" htmlFor="uploadRateLimit">
          <input
            id="uploadRateLimit"
            type="number"
            min={0}
            step={1}
            value={form.uploadRateLimit}
            onChange={(e) => setForm({...form, uploadRateLimit: Number(e.target.value)})}
            className="min-h-11 w-32 rounded-lg border border-border/40 bg-card px-3 text-sm"
          />
          <p className="mt-1 text-xs text-muted-foreground">0 = unlimited</p>
        </Field>

        <Field label="AniList sync threshold (%)" htmlFor="threshold">
          <input
            id="threshold"
            type="number"
            min={1}
            max={100}
            value={form.syncThreshold}
            onChange={(e) => setForm({...form, syncThreshold: Number(e.target.value)})}
            className="min-h-11 w-32 rounded-lg border border-border/40 bg-card px-3 text-sm"
          />
        </Field>

        <Field label="AniList client ID" htmlFor="clientId">
          <input
            id="clientId"
            value={form.anilistClientId}
            onChange={(e) => setForm({...form, anilistClientId: e.target.value})}
            className="min-h-11 w-full rounded-lg border border-border/40 bg-card px-3 text-sm"
          />
          <p className="mt-1 text-xs text-muted-foreground">
            Redirect URL on the AniList app must be http://127.0.0.1:58496/callback
          </p>
        </Field>

        <div className="rounded-xl bg-card p-4">
          <h3 className="text-sm font-medium">AniList account</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            {status.connected ? `Connected as ${status.username}` : 'Not connected. Open login, then authorize in the browser.'}
          </p>
          <div className="mt-3 flex flex-wrap gap-2">
            <button
              type="button"
              className="min-h-11 cursor-pointer rounded-lg bg-secondary px-4 text-sm text-on-secondary"
              onClick={() => OpenAnilistLogin().catch((err) => notice(errorMessage(err), true))}
            >
              Open login
            </button>
            {status.connected && (
              <button
                type="button"
                className="min-h-11 cursor-pointer rounded-lg px-4 text-sm text-destructive"
                onClick={() => LogoutAnilist().then(() => reload()).catch((err) => notice(errorMessage(err), true))}
              >
                Log out
              </button>
            )}
          </div>
          <label htmlFor="token" className="mt-4 block text-sm font-medium">Paste access token</label>
          <textarea
            id="token"
            value={token}
            onChange={(e) => setToken(e.target.value)}
            rows={2}
            className="mt-2 w-full rounded-lg border border-border/40 bg-muted px-3 py-2 text-sm"
          />
          <button
            type="button"
            disabled={!token.trim()}
            className="mt-2 min-h-11 cursor-pointer rounded-lg bg-accent px-4 text-sm font-medium text-on-accent disabled:opacity-50"
            onClick={() => SaveAnilistToken(token.trim()).then(() => { setToken(''); return reload() }).then(() => notice('AniList connected')).catch((err) => notice(errorMessage(err), true))}
          >
            Save token
          </button>
        </div>

        <button
          type="submit"
          disabled={saving}
          className="min-h-11 w-fit cursor-pointer rounded-lg bg-accent px-5 text-sm font-medium text-on-accent disabled:opacity-50"
        >
          {saving ? 'Saving…' : 'Save settings'}
        </button>
      </form>
    </section>
  )
}

function Field({label, htmlFor, children}: {label: string; htmlFor: string; children: ReactNode}) {
  return (
    <div>
      <label htmlFor={htmlFor} className="mb-2 block text-sm font-medium">{label}</label>
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
