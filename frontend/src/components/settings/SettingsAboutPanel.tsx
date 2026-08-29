import type {SettingsView, UpdateInfo} from '../../lib/types'
import {LabelWithHint} from '../LabelWithHint'
import {Button} from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {NativeSelect, NativeSelectOption} from '@/components/ui/native-select'

type Props = {
  appVersion: string
  form: SettingsView
  update: UpdateInfo | null
  saving: boolean
  checkingUpdate: boolean
  applyingUpdate: boolean
  onSaveUpdateChannel: (channel: string) => void
  onCheckUpdate: () => void
  onApplyUpdate: () => void
  onOpenRelease: () => void
}

export function SettingsAboutPanel({
  appVersion,
  form,
  update,
  saving,
  checkingUpdate,
  applyingUpdate,
  onSaveUpdateChannel,
  onCheckUpdate,
  onApplyUpdate,
  onOpenRelease,
}: Props) {
  return (
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
          <NativeSelectOption value="prerelease">Prerelease</NativeSelectOption>
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
      </CardContent>
    </Card>
  )
}
