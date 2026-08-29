import type {Dispatch, FormEvent, SetStateAction} from 'react'
import type {SettingsView} from '../../lib/types'
import {SettingsCheckboxRow} from './SettingsCheckboxRow'
import {SettingsField} from './SettingsField'
import {Alert, AlertDescription} from '@/components/ui/alert'
import {Button} from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {Input} from '@/components/ui/input'

const mpvPathHint = 'MPV path changes take effect after restart.'

const anime4kHint =
  'Applies Anime4K Mode A shaders when MPV starts. Shaders are cached in your Miru config folder.'

const discordRpcHint =
  'Show the anime you are watching on your Discord profile while MPV is playing. Discord Rich Presence requires the Discord desktop app to be running.'

const discordAppIdHint =
  'Leave empty to use DISCORD_APP_ID from the build .env file.'

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
        <CardContent className="flex flex-col gap-4">
          <SettingsField
            label="MPV path"
            htmlFor="mpvPath"
            hint={mpvPathHint}
            className="mt-0"
          >
            <div className="flex flex-wrap gap-2">
              <Input
                id="mpvPath"
                value={form.mpvPath}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    mpvPath: event.target.value,
                  }))
                }
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
          <SettingsCheckboxRow
            id="anime4kEnabled"
            label="Enable Anime4K upscaling"
            hint={anime4kHint}
            checked={form.anime4kEnabled}
            onChange={(value) =>
              setForm((current) => ({...current, anime4kEnabled: value}))
            }
          />
          {form.anime4kEnabled && !form.anime4kShadersReady && (
            <Alert variant="destructive">
              <AlertDescription>
                Anime4K shaders are not installed yet. Save playback settings to
                download them.
              </AlertDescription>
            </Alert>
          )}
          {form.anime4kEnabled && form.anime4kShadersReady && (
            <p className="text-xs text-muted-foreground">
              Anime4K shaders are installed.
            </p>
          )}
          <SettingsCheckboxRow
            id="discordRpcEnabled"
            label="Discord Rich Presence"
            hint={discordRpcHint}
            checked={form.discordRpcEnabled}
            onChange={(value) =>
              setForm((current) => ({...current, discordRpcEnabled: value}))
            }
          />
          {form.discordRpcEnabled && (
            <SettingsField
              label="Discord application ID"
              htmlFor="discordAppId"
              hint={discordAppIdHint}
            >
              <Input
                id="discordAppId"
                value={form.discordAppId}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    discordAppId: event.target.value,
                  }))
                }
                placeholder="From the Discord Developer Portal"
                className="bg-card"
              />
            </SettingsField>
          )}
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
