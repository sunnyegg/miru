import {IconLibrary} from './Icons'

export function Splash() {
  return (
    <div
      className="flex h-full w-full flex-col items-center justify-center bg-background"
      role="status"
      aria-label="Loading Miru"
    >
      <div className="flex items-center gap-3 text-foreground">
        <IconLibrary className="h-6 w-6" />
        <span className="text-lg font-semibold tracking-tight">Miru</span>
      </div>
      <div className="mt-6 h-px w-48 overflow-hidden bg-border" aria-hidden="true">
        <div className="h-full w-1/3 bg-accent animate-pulse" />
      </div>
    </div>
  )
}
