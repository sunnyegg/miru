import {IconCalendar, IconDownload, IconLibrary, IconSearch, IconSettings, IconWatching} from './Icons'
import type {TabId} from '../lib/types'

const destinations: {id: TabId; label: string; icon: typeof IconLibrary}[] = [
  {id: 'library', label: 'Library', icon: IconLibrary},
  {id: 'watching', label: 'Watching', icon: IconWatching},
  {id: 'search', label: 'Search', icon: IconSearch},
  {id: 'downloads', label: 'Downloads', icon: IconDownload},
  {id: 'calendar', label: 'Airing', icon: IconCalendar},
]

type Props = {
  current: TabId
  onChange: (id: TabId) => void
}

function NavButton({
  id,
  label,
  icon: Icon,
  current,
  onChange,
}: {
  id: TabId
  label: string
  icon: typeof IconLibrary
  current: TabId
  onChange: (id: TabId) => void
}) {
  const active = current === id
  return (
    <button
      type="button"
      onClick={() => onChange(id)}
      aria-current={active ? 'page' : undefined}
      className={`flex min-h-11 w-full cursor-pointer items-center justify-center gap-3 border-l px-0 text-left text-sm transition-colors duration-200 sm:justify-start sm:px-3 ${
        active
          ? 'border-accent bg-muted text-foreground'
          : 'border-transparent text-muted-foreground hover:bg-muted hover:text-foreground'
      }`}
    >
      <Icon className="h-4 w-4 shrink-0" />
      <span className="sr-only sm:not-sr-only">{label}</span>
    </button>
  )
}

export function Sidebar({current, onChange}: Props) {
  return (
    <nav
      className="flex w-12 shrink-0 flex-col border-r border-border bg-bezel py-4 sm:w-44"
      aria-label="Main"
    >
      <div className="px-2 pb-5 sm:px-5">
        <h1 className="text-center text-[11px] font-semibold tracking-tight text-foreground sm:text-left sm:text-lg">Miru</h1>
      </div>
      <ul className="flex flex-col">
        {destinations.map((item) => (
          <li key={item.id}>
            <NavButton {...item} current={current} onChange={onChange} />
          </li>
        ))}
      </ul>
      <div className="mt-auto">
        <NavButton id="settings" label="Settings" icon={IconSettings} current={current} onChange={onChange} />
      </div>
    </nav>
  )
}
