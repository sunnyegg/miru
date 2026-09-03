import {useRef, useState} from 'react'
import {formatBytes} from '../../lib/format'
import type {
  DataSizeView,
  SettingsView,
  UpdateInfo,
  UpdateProgress,
} from '../../lib/types'
import {LabelWithHint} from '../LabelWithHint'
import {UpdateProgressBar} from '../UpdateProgressBar'
import {Button} from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {Dialog} from '@/components/ui/dialog'
import {NativeSelect, NativeSelectOption} from '@/components/ui/native-select'

type Props = {
  appVersion: string
  form: SettingsView
  update: UpdateInfo | null
  updateProgress: UpdateProgress | null
  saving: boolean
  checkingUpdate: boolean
  applyingUpdate: boolean
  onSaveUpdateChannel: (channel: string) => void
  onCheckUpdate: () => void
  onApplyUpdate: () => void
  onOpenRelease: () => void
  onOpenChangelog: () => void
  dataSize: DataSizeView | null
  dataSizeError: string
  loadingDataSize: boolean
  playbackActive: boolean
  resettingData: boolean
  resetDataError: string
  onReloadDataSize: () => void
  onDeleteAllData: () => void
}

export function SettingsAboutPanel({
  appVersion,
  form,
  update,
  updateProgress,
  saving,
  checkingUpdate,
  applyingUpdate,
  onSaveUpdateChannel,
  onCheckUpdate,
  onApplyUpdate,
  onOpenRelease,
  onOpenChangelog,
  dataSize,
  dataSizeError,
  loadingDataSize,
  playbackActive,
  resettingData,
  resetDataError,
  onReloadDataSize,
  onDeleteAllData,
}: Props) {
  const [resetDialogOpen, setResetDialogOpen] = useState(false)
  const cancelResetRef = useRef<HTMLButtonElement | null>(null)

  return (
    <>
      <Card className="w-full border border-border">
        <CardHeader>
          <CardTitle>About</CardTitle>
          <CardDescription>Version {appVersion || 'dev'}</CardDescription>
        </CardHeader>
        <CardContent>
          <LabelWithHint
            htmlFor="updateChannel"
            label="Update channel"
            hint="Prerelease includes alpha and beta builds. Stable ignores them."
          />
          <NativeSelect
            id="updateChannel"
            value={form.updateChannel}
            disabled={saving || checkingUpdate}
            onChange={(event) => onSaveUpdateChannel(event.target.value)}
          >
            <NativeSelectOption value="stable">Stable</NativeSelectOption>
            <NativeSelectOption value="prerelease">
              Prerelease
            </NativeSelectOption>
          </NativeSelect>
          {update?.available ? (
            <>
              <p className="mt-3 text-sm">Miru {update.latest} is available.</p>
              <div className="mt-3 flex flex-wrap gap-2">
                <Button
                  type="button"
                  disabled={applyingUpdate}
                  onClick={onApplyUpdate}
                >
                  {applyingUpdate ? 'Updating…' : 'Update now'}
                </Button>
                <Button type="button" variant="muted" onClick={onOpenChangelog}>
                  View changelog
                </Button>
                <Button type="button" variant="muted" onClick={onOpenRelease}>
                  Open download page
                </Button>
              </div>
              {applyingUpdate && updateProgress && (
                <div className="mt-3">
                  <UpdateProgressBar progress={updateProgress} />
                </div>
              )}
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

          <section
            className="mt-6 border-t border-border pt-6"
            aria-labelledby="delete-all-data-title"
          >
            <h3 id="delete-all-data-title" className="text-sm font-medium">
              Delete all Miru data
            </h3>
            <p className="mt-1 text-sm text-muted-foreground">
              Miru app data uses{' '}
              {loadingDataSize
                ? 'calculating…'
                : dataSize
                  ? formatBytes(dataSize.bytes)
                  : 'an unknown amount of space'}
              . Downloaded files are not included.
            </p>
            {dataSizeError && (
              <div className="mt-3" role="alert">
                <p className="text-sm text-destructive">{dataSizeError}</p>
                <Button
                  type="button"
                  variant="ghost"
                  className="mt-1"
                  disabled={loadingDataSize}
                  onClick={onReloadDataSize}
                >
                  Try size again
                </Button>
              </div>
            )}
            {(dataSize?.cleanupPending || dataSize?.resetError) && (
              <p className="mt-3 text-sm text-destructive" role="alert">
                {dataSize.resetError ||
                  'Old Miru data could not be fully removed. Quit and open Miru again to retry.'}
              </p>
            )}
            <Button
              type="button"
              variant="destructive"
              className="mt-4"
              disabled={playbackActive || resettingData}
              onClick={() => setResetDialogOpen(true)}
            >
              Delete all data
            </Button>
            {playbackActive && (
              <p className="mt-2 text-sm text-muted-foreground">
                Stop playback before deleting all Miru data.
              </p>
            )}
          </section>
        </CardContent>
      </Card>

      <Dialog.Root
        open={resetDialogOpen}
        onOpenChange={(open) => {
          if (!resettingData) {
            setResetDialogOpen(open)
          }
        }}
      >
        <Dialog.Portal>
          <Dialog.Backdrop />
          <Dialog.Viewport>
            <Dialog.Panel
              aria-labelledby="delete-all-data-dialog-title"
              initialFocus={cancelResetRef}
            >
              <Dialog.Title id="delete-all-data-dialog-title">
                Delete all Miru data?
              </Dialog.Title>
              <Dialog.Description>
                Settings, library history, watched progress, AniList login, RSS
                feeds, torrent history, cache, Anime4K shaders, logs, and saved
                search state will be permanently deleted. Downloaded files stay
                on disk.
              </Dialog.Description>
              <p className="mt-3 text-sm text-muted-foreground">
                Active downloads will stop and Miru will quit. Open Miru again
                to finish the reset.
              </p>
              {resetDataError && (
                <p className="mt-3 text-sm text-destructive" role="alert">
                  {resetDataError}
                </p>
              )}
              <div className="mt-4 flex flex-wrap gap-2">
                <Button
                  type="button"
                  variant="destructive"
                  disabled={resettingData}
                  onClick={onDeleteAllData}
                >
                  {resettingData ? 'Preparing reset…' : 'Delete data and quit'}
                </Button>
                <Button
                  ref={cancelResetRef}
                  type="button"
                  variant="ghost"
                  disabled={resettingData}
                  onClick={() => setResetDialogOpen(false)}
                >
                  Cancel
                </Button>
              </div>
            </Dialog.Panel>
          </Dialog.Viewport>
        </Dialog.Portal>
      </Dialog.Root>
    </>
  )
}
