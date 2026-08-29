import type {Dispatch, FormEvent, SetStateAction} from 'react'
import type {SettingsView} from '../../lib/types'
import {Alert, AlertDescription} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'
import {Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle} from '@/components/ui/card'
import {Input} from '@/components/ui/input'
import {Label} from '@/components/ui/label'
import {SettingsField} from './SettingsField'

type Props = {
  form: SettingsView
  setForm: Dispatch<SetStateAction<SettingsView>>
  saving: boolean
  onSubmit: (event: FormEvent) => void
  onDetectMpv: () => void
  onPickMpvPath: () => void
}

export function SettingsPlaybackPanel({
  form,
  setForm,
  saving,
  onSubmit,
  onDetectMpv,
  onPickMpvPath,
}: Props) {
  return (
    <form onSubmit={onSubmit} className="w-full">
      <Card className="w-full border border-border">
        <CardHeader>
          <CardTitle>Playback</CardTitle>
          <CardDescription>MPV binary used to play episodes.</CardDescription>
        </CardHeader>
        <CardContent>
          <SettingsField label="MPV path" htmlFor="mpvPath">
            <div className="flex flex-wrap gap-2">
              <Input
                id="mpvPath"
                value={form.mpvPath}
                onChange={(event) => setForm((current) => ({...current, mpvPath: event.target.value}))}
                className="min-w-0 flex-1 bg-card"
              />
              <Button type="button" variant="muted" onClick={onDetectMpv}>
                Detect
              </Button>
              <Button type="button" variant="muted" onClick={onPickMpvPath}>
                Browse
              </Button>
            </div>
          </SettingsField>
          <div className="mt-4">
            <label className="flex min-h-11 cursor-pointer items-center gap-3">
              <input
                type="checkbox"
                checked={form.anime4kEnabled}
                onChange={(event) =>
                  setForm((current) => ({...current, anime4kEnabled: event.target.checked}))
                }
                className="size-4 accent-primary"
              />
              <span className="text-sm">Enable Anime4K upscaling</span>
            </label>
            <p className="mt-1 text-xs text-muted-foreground">
              Applies Anime4K Mode A shaders when MPV starts. Shaders are cached in your Miru config
              folder.
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
              onChange={(event) =>
                setForm((current) => ({...current, discordRpcEnabled: event.target.checked}))
              }
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
            <SettingsField label="Discord application ID" htmlFor="discordAppId">
              <Input
                id="discordAppId"
                value={form.discordAppId}
                onChange={(event) =>
                  setForm((current) => ({...current, discordAppId: event.target.value}))
                }
                placeholder="From the Discord Developer Portal"
                className="bg-card"
              />
              <p className="mt-1 text-xs text-muted-foreground">
                Leave empty to use DISCORD_APP_ID from the build .env file.
              </p>
            </SettingsField>
          )}
          <p className="mt-4 text-xs text-muted-foreground">
            Discord Rich Presence requires the Discord desktop app to be running. MPV path changes
            take effect after restart.
          </p>
        </CardContent>
        <CardFooter>
          <Button type="submit" variant="secondary" disabled={saving}>
            {saving ? 'Saving…' : 'Save'}
          </Button>
        </CardFooter>
      </Card>
    </form>
  )
}
