import type {FormEvent} from 'react'
import {Button} from '@/components/ui/button'
import {Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle} from '@/components/ui/card'
import {Label} from '@/components/ui/label'

type Props = {
  closeToTray: boolean
  saving: boolean
  onCloseToTrayChange: (value: boolean) => void
  onSubmit: (event: FormEvent) => void
}

export function SettingsDesktopPanel({closeToTray, saving, onCloseToTrayChange, onSubmit}: Props) {
  return (
    <form onSubmit={onSubmit} className="w-full">
      <Card className="w-full border border-border">
        <CardHeader>
          <CardTitle>Desktop</CardTitle>
          <CardDescription>Control what happens when you close the Miru window.</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-start gap-3">
            <input
              id="closeToTray"
              type="checkbox"
              checked={closeToTray}
              onChange={(event) => onCloseToTrayChange(event.target.checked)}
              className="mt-1 size-4 shrink-0 accent-primary"
            />
            <div>
              <Label htmlFor="closeToTray">Close to system tray</Label>
              <p className="mt-1 text-sm text-muted-foreground">
                Hide Miru instead of quitting so downloads and RSS polling keep running. Use the
                tray icon to show Miru again or quit fully.
              </p>
              <p className="mt-1 text-xs text-muted-foreground">
                Requires a system tray host (KDE, GNOME with SNI, Waybar, and similar). If no tray
                is available, Miru stays open when you close the window.
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
