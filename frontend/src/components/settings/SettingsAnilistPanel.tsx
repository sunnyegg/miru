import type {Dispatch, FormEvent, SetStateAction} from 'react'
import type {AnilistStatus, SettingsView} from '../../lib/types'
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
import {SettingsField} from './SettingsField'

type Props = {
  form: SettingsView
  setForm: Dispatch<SetStateAction<SettingsView>>
  status: AnilistStatus
  saving: boolean
  onSubmit: (event: FormEvent) => void
  onOpenLogin: () => void
  onLogout: () => void
}

export function SettingsAnilistPanel({
  form,
  setForm,
  status,
  saving,
  onSubmit,
  onOpenLogin,
  onLogout,
}: Props) {
  return (
    <form onSubmit={onSubmit} className="w-full">
      <Card className="w-full border border-border">
        <CardHeader>
          <CardTitle>AniList</CardTitle>
          <CardDescription>
            {status.connected
              ? `Connected as ${status.username}`
              : 'Not connected. Open login, then authorize in the browser.'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="secondary" onClick={onOpenLogin}>
              Open login
            </Button>
            {status.connected && (
              <Button
                type="button"
                variant="ghost"
                className="text-destructive hover:text-destructive"
                onClick={onLogout}
              >
                Log out
              </Button>
            )}
          </div>
          <SettingsField label="Sync threshold (%)" htmlFor="threshold">
            <Input
              id="threshold"
              type="number"
              min={1}
              max={100}
              value={form.syncThreshold}
              onChange={(event) =>
                setForm((current) => ({
                  ...current,
                  syncThreshold: Number(event.target.value),
                }))
              }
              className="w-32 bg-card"
            />
          </SettingsField>
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
