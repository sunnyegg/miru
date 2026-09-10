import {
  IconCalendar,
  IconDownload,
  IconLibrary,
  IconSearch,
  IconSettings,
  IconWatching,
} from './Icons'
import {Button} from '@/components/ui/button'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/components/ui/tooltip'
import {cn} from '@/lib/utils'
import type {TabId} from '../lib/types'
import {useNavigationStore} from '../stores/navigationStore'

const destinations: {
  id: TabId
  label: string
  icon: typeof IconLibrary
  disabled?: boolean
  hint?: string
}[] = [
  {id: 'library', label: 'Library', icon: IconLibrary},
  {id: 'watching', label: 'Watching', icon: IconWatching},
  {id: 'search', label: 'Search', icon: IconSearch},
  {id: 'downloads', label: 'Downloads', icon: IconDownload},
  {
    id: 'calendar',
    label: 'Airing',
    icon: IconCalendar,
    disabled: true,
    hint: 'Temporarily unavailable. AniList is returning 403 for its API.',
  },
]

function NavButton({
  id,
  label,
  icon: Icon,
  current,
  onChange,
  disabled = false,
  hint,
}: {
  id: TabId
  label: string
  icon: typeof IconLibrary
  current: TabId
  onChange: (id: TabId) => void
  disabled?: boolean
  hint?: string
}) {
  const active = current === id
  const button = (
    <Button
      type="button"
      variant="ghost"
      onClick={() => {
        if (!disabled) {
          onChange(id)
        }
      }}
      aria-current={active ? 'page' : undefined}
      aria-disabled={disabled || undefined}
      className={cn(
        'w-full justify-center gap-3 border-l px-0 motion-reduce:transition-none sm:justify-start sm:px-3',
        disabled
          ? 'cursor-not-allowed border-transparent text-muted-foreground opacity-50 hover:bg-transparent hover:text-muted-foreground'
          : active
            ? 'border-accent bg-muted text-foreground hover:text-foreground'
            : 'border-transparent hover:bg-muted hover:text-foreground',
      )}
    >
      <Icon className="h-4 w-4 shrink-0" />
      <span className="sr-only sm:not-sr-only">{label}</span>
    </Button>
  )

  if (!hint) {
    return button
  }

  return (
    <Tooltip>
      <TooltipTrigger render={button} />
      <TooltipContent side="right">{hint}</TooltipContent>
    </Tooltip>
  )
}

export function Sidebar() {
  const current = useNavigationStore((state) => state.tab)
  const onChange = useNavigationStore((state) => state.setTab)

  return (
    <nav
      className="flex w-12 shrink-0 flex-col border-r border-border bg-bezel py-4 sm:w-44"
      aria-label="Main"
    >
      <div className="px-2 pb-5 sm:px-5">
        <h1 className="text-center text-[11px] font-semibold tracking-tight text-foreground sm:text-left sm:text-lg">
          Miru
        </h1>
      </div>
      <ul className="flex flex-col">
        {destinations.map((item) => (
          <li key={item.id}>
            <NavButton {...item} current={current} onChange={onChange} />
          </li>
        ))}
      </ul>
      <div className="mt-auto">
        <NavButton
          id="settings"
          label="Settings"
          icon={IconSettings}
          current={current}
          onChange={onChange}
        />
      </div>
    </nav>
  )
}
