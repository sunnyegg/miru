import {Button} from '@/components/ui/button'
import {Dialog} from '@/components/ui/dialog'

type Props = {
  open: boolean
  torrentName: string
  busy: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => void
}

export function DeleteDownloadDialog({open, torrentName, busy, onOpenChange, onConfirm}: Props) {
  return (
    <Dialog.Root
      open={open}
      onOpenChange={(nextOpen) => {
        if (!busy) {
          onOpenChange(nextOpen)
        }
      }}
    >
      <Dialog.Portal>
        <Dialog.Backdrop />
        <Dialog.Viewport>
          <Dialog.Panel aria-labelledby="delete-download-title">
            <Dialog.Title id="delete-download-title">Delete downloaded files?</Dialog.Title>
            <Dialog.Description>
              This removes <span className="font-medium text-foreground">{torrentName}</span> from
              Miru and permanently deletes its files from disk. This cannot be undone.
            </Dialog.Description>
            <div className="mt-4 flex flex-wrap gap-2">
              <Button type="button" variant="destructive" disabled={busy} onClick={onConfirm}>
                {busy ? 'Deleting…' : 'Delete files'}
              </Button>
              <Button type="button" variant="ghost" disabled={busy} onClick={() => onOpenChange(false)}>
                Cancel
              </Button>
            </div>
          </Dialog.Panel>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
