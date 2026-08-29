import {useState} from 'react'
import {ConfirmWindowClose} from '../../wailsjs/go/main/App'
import {errorMessage} from '../lib/format'
import {Button} from '@/components/ui/button'
import {Dialog} from '@/components/ui/dialog'
import {Label} from '@/components/ui/label'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  notice: (message: string, isError?: boolean) => void
}

export function CloseToTrayDialog({open, onOpenChange, notice}: Props) {
  const [rememberChoice, setRememberChoice] = useState(false)
  const [busyAction, setBusyAction] = useState<'hide' | 'quit' | null>(null)

  async function handleClose(action: 'hide' | 'quit') {
    setBusyAction(action)
    try {
      await ConfirmWindowClose(action, rememberChoice)
      if (action === 'hide') {
        onOpenChange(false)
      }
    } catch (err) {
      notice(errorMessage(err), true)
      onOpenChange(false)
    } finally {
      setBusyAction(null)
    }
  }

  return (
    <Dialog.Root
      open={open}
      onOpenChange={(nextOpen) => {
        if (!busyAction) {
          onOpenChange(nextOpen)
        }
      }}
    >
      <Dialog.Portal>
        <Dialog.Backdrop />
        <Dialog.Viewport>
          <Dialog.Panel aria-labelledby="close-to-tray-title">
            <Dialog.Title id="close-to-tray-title">Keep Miru running?</Dialog.Title>
            <Dialog.Description>
              Hide Miru to the system tray so downloads and RSS polling keep running in the
              background. You can change this behavior in Settings.
            </Dialog.Description>
            <label className="mt-4 flex min-h-11 cursor-pointer items-center gap-3">
              <input
                id="rememberCloseChoice"
                type="checkbox"
                checked={rememberChoice}
                onChange={(event) => setRememberChoice(event.target.checked)}
                className="size-4 accent-primary"
              />
              <Label htmlFor="rememberCloseChoice">Remember this choice</Label>
            </label>
            <div className="mt-4 flex flex-wrap gap-2">
              <Button
                type="button"
                disabled={busyAction !== null}
                onClick={() => void handleClose('hide')}
              >
                {busyAction === 'hide' ? 'Hiding…' : 'Hide to tray'}
              </Button>
              <Button
                type="button"
                variant="secondary"
                disabled={busyAction !== null}
                onClick={() => void handleClose('quit')}
              >
                {busyAction === 'quit' ? 'Quitting…' : 'Quit'}
              </Button>
            </div>
          </Dialog.Panel>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
