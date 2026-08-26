import {IconCalendar, IconDownload, IconLibrary, IconSettings} from './Icons'
import type {TabId} from '../lib/types'

const items: {id: TabId; label: string; icon: typeof IconLibrary}[] = [
  {id: 'library', label: 'Library', icon: IconLibrary},
  {id: 'downloads', label: 'Downloads', icon: IconDownload},
  {id: 'calendar', label: 'Airing', icon: IconCalendar},
  {id: 'settings', label: 'Settings', icon: IconSettings},
]

type Props = {
  current: TabId
  onChange: (id: TabId) => void
}

export function Sidebar({current, onChange}: Props) {
  return (
    <nav className="flex w-56 shrink-0 flex-col border-r border-border/40 bg-primary px-3 py-5" aria-label="Main">
      <div className="px-3 pb-6">
        <p className="text-xs font-medium uppercase tracking-[0.2em] text-muted-foreground">見る</p>
        <h1 className="mt-1 text-xl font-semibold text-foreground">Miru</h1>
      </div>
      <ul className="flex flex-col gap-2">
        {items.map((item) => {
          const active = current === item.id
          const Icon = item.icon
          return (
            <li key={item.id}>
              <button
                type="button"
                onClick={() => onChange(item.id)}
                aria-current={active ? 'page' : undefined}
                className={`flex min-h-11 w-full cursor-pointer items-center gap-3 rounded-lg px-3 text-left text-sm transition-colors duration-200 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring ${
                  active
                    ? 'bg-secondary text-on-secondary'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                }`}
              >
                <Icon className="h-5 w-5 shrink-0" />
                <span>{item.label}</span>
              </button>
            </li>
          )
        })}
      </ul>
    </nav>
  )
}
