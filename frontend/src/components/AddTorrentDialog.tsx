import {useEffect, useState, type FormEvent} from 'react'
import {Button} from '@/components/ui/button'
import {Dialog} from '@/components/ui/dialog'
import {Label} from '@/components/ui/label'
import {Textarea} from '@/components/ui/textarea'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  busy: boolean
  onStartMagnet: (source: string) => Promise<void>
  onStartFile: () => Promise<void>
}

export function AddTorrentDialog({open, onOpenChange, busy, onStartMagnet, onStartFile}: Props) {
  const [magnet, setMagnet] = useState('')

  useEffect(() => {
    if (!open) {
      setMagnet('')
    }
  }, [open])

  async function handleMagnetSubmit(event: FormEvent) {
    event.preventDefault()
    const source = magnet.trim()
    if (!source) {
      return
    }
    await onStartMagnet(source)
  }

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
          <Dialog.Panel aria-labelledby="add-torrent-title">
            <Dialog.Title id="add-torrent-title">Add torrent</Dialog.Title>
            <Dialog.Description>
              Choose which files to keep before the download starts. Extra torrents wait in queue.
            </Dialog.Description>

            <form className="mt-4 flex flex-col gap-3" onSubmit={(event) => void handleMagnetSubmit(event)}>
              <Label htmlFor="add-torrent-magnet">Magnet link</Label>
              <Textarea
                id="add-torrent-magnet"
                value={magnet}
                onChange={(event) => setMagnet(event.target.value)}
                rows={3}
                placeholder="magnet:?xt=urn:btih:..."
                disabled={busy}
              />
              <div className="flex flex-wrap gap-2">
                <Button type="submit" disabled={busy || !magnet.trim()}>
                  {busy ? 'Loading…' : 'Add magnet'}
                </Button>
                <Button
                  type="button"
                  variant="secondary"
                  disabled={busy}
                  onClick={() => void onStartFile()}
                >
                  Open .torrent file
                </Button>
                <Button type="button" variant="ghost" disabled={busy} onClick={() => onOpenChange(false)}>
                  Cancel
                </Button>
              </div>
            </form>
          </Dialog.Panel>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
