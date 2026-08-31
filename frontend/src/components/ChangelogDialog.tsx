import {useState} from 'react'
import {BrowserOpenURL} from '../../wailsjs/runtime/runtime'
import {errorMessage} from '../lib/format'
import {Button} from '@/components/ui/button'
import {Dialog} from '@/components/ui/dialog'
import ReactMarkdown from 'react-markdown'
import rehypeSanitize from 'rehype-sanitize'
import remarkGfm from 'remark-gfm'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  version: string
  notes: string
  releaseUrl: string
  notice: (message: string, isError?: boolean) => void
  onDismiss: (version: string) => Promise<void> | void
}

export function ChangelogDialog({
  open,
  onOpenChange,
  version,
  notes,
  releaseUrl,
  notice,
  onDismiss,
}: Props) {
  const [marking, setMarking] = useState(false)

  async function handleDismiss() {
    if (marking || !version) {
      onOpenChange(false)
      return
    }
    setMarking(true)
    try {
      await onDismiss(version)
    } catch (err) {
      notice(errorMessage(err), true)
    } finally {
      setMarking(false)
      onOpenChange(false)
    }
  }

  function openRelease() {
    if (releaseUrl) {
      BrowserOpenURL(releaseUrl)
    }
  }

  const trimmedNotes = notes.trim()

  return (
    <Dialog.Root
      open={open}
      onOpenChange={(nextOpen) => {
        if (nextOpen) {
          onOpenChange(true)
          return
        }
        void handleDismiss()
      }}
    >
      <Dialog.Portal>
        <Dialog.Backdrop />
        <Dialog.Viewport>
          <Dialog.Panel
            className="max-w-2xl p-6"
            aria-labelledby="changelog-title"
          >
            <Dialog.Title id="changelog-title">
              {version
                ? `What\u2019s new in Miru ${version}`
                : 'What\u2019s new'}
            </Dialog.Title>
            <Dialog.Description>
              Highlights from the latest release.
            </Dialog.Description>
            <div
              className="mt-4 max-h-[60vh] overflow-y-auto text-sm
                [&_a]:text-accent [&_a]:underline
                [&_code]:bg-muted [&_code]:px-1 [&_code]:rounded
                [&_h1]:mt-3 [&_h1]:text-base [&_h1]:font-medium
                [&_h2]:mt-3 [&_h2]:text-base [&_h2]:font-medium
                [&_h3]:mt-2 [&_h3]:text-sm [&_h3]:font-medium
                [&_p]:mt-2 [&_ul]:mt-2 [&_ol]:mt-2 [&_li]:mt-1
                [&_ul]:list-disc [&_ul]:pl-5
                [&_ol]:list-decimal [&_ol]:pl-5
                [&_blockquote]:mt-2 [&_blockquote]:border-l-2 [&_blockquote]:border-border [&_blockquote]:pl-3 [&_blockquote]:text-muted-foreground
                [&_table]:mt-2 [&_th]:border [&_th]:border-border [&_th]:px-2 [&_th]:py-1
                [&_td]:border [&_td]:border-border [&_td]:px-2 [&_td]:py-1"
            >
              {trimmedNotes ? (
                <ReactMarkdown
                  remarkPlugins={[remarkGfm]}
                  rehypePlugins={[rehypeSanitize]}
                  components={{
                    a: ({href, children}) => (
                      <a href={href} target="_blank" rel="noreferrer">
                        {children}
                      </a>
                    ),
                  }}
                >
                  {notes}
                </ReactMarkdown>
              ) : (
                <p className="text-muted-foreground">
                  No changelog provided for this release.
                </p>
              )}
            </div>
            <div className="mt-4 flex flex-wrap justify-end gap-2">
              {releaseUrl && (
                <Button
                  type="button"
                  variant="muted"
                  disabled={marking}
                  onClick={openRelease}
                >
                  Open on GitHub
                </Button>
              )}
              <Button
                type="button"
                disabled={marking}
                onClick={() => void handleDismiss()}
              >
                {marking ? 'Saving…' : 'Close'}
              </Button>
            </div>
          </Dialog.Panel>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
