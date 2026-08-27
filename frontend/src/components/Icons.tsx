import {
  Calendar,
  Download,
  Eye,
  Folder,
  LayoutGrid,
  Play,
  Search,
  Settings,
  type LucideProps,
} from 'lucide-react'

type IconProps = Pick<LucideProps, 'className'>

const stroke = {
  'aria-hidden': true,
  strokeWidth: 1.75,
} as const

export function IconLibrary({className}: IconProps) {
  return <LayoutGrid className={className} {...stroke} />
}

export function IconWatching({className}: IconProps) {
  return <Eye className={className} {...stroke} />
}

export function IconSearch({className}: IconProps) {
  return <Search className={className} {...stroke} />
}

export function IconDownload({className}: IconProps) {
  return <Download className={className} {...stroke} />
}

export function IconCalendar({className}: IconProps) {
  return <Calendar className={className} {...stroke} />
}

export function IconSettings({className}: IconProps) {
  return <Settings className={className} {...stroke} />
}

export function IconPlay({className}: IconProps) {
  return <Play className={className} aria-hidden="true" fill="currentColor" strokeWidth={0} />
}

export function IconFolder({className}: IconProps) {
  return <Folder className={className} {...stroke} />
}
