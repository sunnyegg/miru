import type {Dispatch, FormEvent, SetStateAction} from 'react'
import type {SettingsView} from '../../lib/types'
import {SettingsCheckboxRow} from './SettingsCheckboxRow'
import {SettingsField} from './SettingsField'
import {Button} from '@/components/ui/button'
import {Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle} from '@/components/ui/card'
import {Input} from '@/components/ui/input'

type Props = {
  form: SettingsView
  setForm: Dispatch<SetStateAction<SettingsView>>
  saving: boolean
  onSubmit: (event: FormEvent) => void
  onPickDownloadDir: () => void
}

export function SettingsDownloadsPanel({form, setForm, saving, onSubmit, onPickDownloadDir}: Props) {
  return (
    <form onSubmit={onSubmit} className="w-full">
      <Card className="w-full border border-border">
        <CardHeader>
          <CardTitle>Downloads</CardTitle>
          <CardDescription>
            Folder, speed limits, queue, seeding, and RSS auto-download.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <SettingsField label="Download folder" htmlFor="downloadDir" className="mt-0">
            <div className="flex flex-wrap gap-2">
              <Input
                id="downloadDir"
                value={form.downloadDir}
                onChange={(event) =>
                  setForm((current) => ({...current, downloadDir: event.target.value}))
                }
                className="min-w-0 flex-1 bg-card"
              />
              <Button type="button" variant="muted" onClick={onPickDownloadDir}>
                Browse
              </Button>
            </div>
          </SettingsField>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <SettingsField
              label="Download speed limit (KB/s)"
              htmlFor="downloadRateLimit"
              hint="0 = unlimited"
              className="mt-0"
            >
              <Input
                id="downloadRateLimit"
                type="number"
                min={0}
                step={1}
                value={form.downloadRateLimit}
                onChange={(event) =>
                  setForm((current) => ({...current, downloadRateLimit: Number(event.target.value)}))
                }
                className="w-full bg-card"
              />
            </SettingsField>
            <SettingsField
              label="Upload speed limit (KB/s)"
              htmlFor="uploadRateLimit"
              hint="0 = unlimited"
              className="mt-0"
            >
              <Input
                id="uploadRateLimit"
                type="number"
                min={0}
                step={1}
                value={form.uploadRateLimit}
                onChange={(event) =>
                  setForm((current) => ({...current, uploadRateLimit: Number(event.target.value)}))
                }
                className="w-full bg-card"
              />
            </SettingsField>
            <SettingsField
              label="Max concurrent downloads"
              htmlFor="maxConcurrentDownloads"
              hint="Queued torrents start when a slot is free."
              className="mt-0"
            >
              <Input
                id="maxConcurrentDownloads"
                type="number"
                min={1}
                max={8}
                step={1}
                value={form.maxConcurrentDownloads}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    maxConcurrentDownloads: Number(event.target.value),
                  }))
                }
                className="w-full bg-card"
              />
            </SettingsField>
            <SettingsField
              label="Seed ratio"
              htmlFor="seedRatio"
              hint="Upload ratio before auto-finish (0.5 = half the download size). 0 stops seeding right away."
              className="mt-0"
            >
              <Input
                id="seedRatio"
                type="number"
                min={0}
                max={10}
                step={0.1}
                value={form.seedRatio}
                onChange={(event) =>
                  setForm((current) => ({...current, seedRatio: Number(event.target.value)}))
                }
                className="w-full bg-card"
              />
            </SettingsField>
            <SettingsField
              label="RSS poll interval (minutes)"
              htmlFor="rssPollIntervalMinutes"
              hint="How often subscribed RSS feeds are checked in the background (5–1440)."
              className="mt-0"
            >
              <Input
                id="rssPollIntervalMinutes"
                type="number"
                min={5}
                max={1440}
                step={5}
                value={form.rssPollIntervalMinutes}
                onChange={(event) =>
                  setForm((current) => ({
                    ...current,
                    rssPollIntervalMinutes: Number(event.target.value),
                  }))
                }
                className="w-full bg-card"
              />
            </SettingsField>
          </div>

          <SettingsCheckboxRow
            id="rssAutoDownload"
            label="Auto-download new RSS items"
            hint="Queue torrent downloads when subscribed feeds publish new items with magnet links."
            checked={form.rssAutoDownload}
            onChange={(value) => setForm((current) => ({...current, rssAutoDownload: value}))}
          />
          {form.rssAutoDownload && (
            <SettingsCheckboxRow
              id="rssAutoDownloadLibraryOnly"
              label="Only your library"
              hint="Only auto-download when the item title matches an anime in your local library."
              checked={form.rssAutoDownloadLibraryOnly}
              onChange={(value) =>
                setForm((current) => ({...current, rssAutoDownloadLibraryOnly: value}))
              }
            />
          )}
          <SettingsCheckboxRow
            id="downloadNotifications"
            label="Desktop notifications"
            hint="Show an OS notification when a download you started finishes in the background."
            checked={form.downloadNotifications}
            onChange={(value) => setForm((current) => ({...current, downloadNotifications: value}))}
          />
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
