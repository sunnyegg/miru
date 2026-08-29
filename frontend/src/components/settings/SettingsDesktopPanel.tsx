import type {FormEvent} from 'react'
import {SettingsCheckboxRow} from './SettingsCheckboxRow'
import {Button} from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

const closeToTrayHint =
  'Hide Miru instead of quitting so downloads and RSS polling keep running. Use the tray icon to show Miru again or quit fully. Requires a system tray host (KDE, GNOME with SNI, Waybar, and similar). If no tray is available, Miru stays open when you close the window.'

type Props = {
  closeToTray: boolean
  saving: boolean
  onCloseToTrayChange: (value: boolean) => void
  onSubmit: (event: FormEvent) => void
}

export function SettingsDesktopPanel({
  closeToTray,
  saving,
  onCloseToTrayChange,
  onSubmit,
}: Props) {
  return (
    <form onSubmit={onSubmit} className="w-full">
      <Card className="w-full border border-border">
        <CardHeader>
          <CardTitle>Desktop</CardTitle>
          <CardDescription>
            Control what happens when you close the Miru window.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <SettingsCheckboxRow
            id="closeToTray"
            label="Close to system tray"
            hint={closeToTrayHint}
            checked={closeToTray}
            onChange={onCloseToTrayChange}
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
