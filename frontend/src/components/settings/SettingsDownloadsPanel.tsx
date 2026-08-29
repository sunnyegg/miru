import type {Dispatch, FormEvent, SetStateAction} from 'react'
import type {SettingsView} from '../../lib/types'
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
            <SettingsField label="Download speed limit (KB/s)" htmlFor="downloadRateLimit" className="mt-0">
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
              <p className="mt-1 text-xs text-muted-foreground">0 = unlimited</p>
            </SettingsField>
            <SettingsField label="Upload speed limit (KB/s)" htmlFor="uploadRateLimit" className="mt-0">
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
              <p className="mt-1 text-xs text-muted-foreground">0 = unlimited</p>
            </SettingsField>
            <SettingsField label="Max concurrent downloads" htmlFor="maxConcurrentDownloads" className="mt-0">
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
              <p className="mt-1 text-xs text-muted-foreground">
                Queued torrents start when a slot is free.
              </p>
            </SettingsField>
            <SettingsField label="Seed ratio" htmlFor="seedRatio" className="mt-0">
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
              <p className="mt-1 text-xs text-muted-foreground">
                Upload ratio before auto-finish (0.5 = half the download size). 0 stops seeding right
                away.
              </p>
            </SettingsField>
            <SettingsField label="RSS poll interval (minutes)" htmlFor="rssPollIntervalMinutes" className="mt-0">
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
              <p className="mt-1 text-xs text-muted-foreground">
                How often subscribed RSS feeds are checked in the background (5–1440).
              </p>
            </SettingsField>
          </div>

          <div className="flex items-start gap-3">
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
            <div className="flex items-start gap-3">
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
          <div className="flex items-start gap-3">
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
