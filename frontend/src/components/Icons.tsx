import {
  Calendar,
  Check,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  CircleHelp,
  Download,
  Folder,
  LayoutGrid,
  Play,
  Search,
  Settings,
  User,
  X,
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

export function IconSearch({className}: IconProps) {
  return <Search className={className} {...stroke} />
}

export function IconProfile({className}: IconProps) {
  return <User className={className} {...stroke} />
}

export function IconClose({className}: IconProps) {
  return <X className={className} {...stroke} />
}

export function IconDownload({className}: IconProps) {
  return <Download className={className} {...stroke} />
}

export function IconCalendar({className}: IconProps) {
  return <Calendar className={className} {...stroke} />
}

export function IconCheck({className}: IconProps) {
  return <Check className={className} {...stroke} />
}

export function IconChevronDown({className}: IconProps) {
  return <ChevronDown className={className} {...stroke} />
}

export function IconSettings({className}: IconProps) {
  return <Settings className={className} {...stroke} />
}

export function IconBack({className}: IconProps) {
  return <ChevronLeft className={className} {...stroke} />
}

export function IconChevronRight({className}: IconProps) {
  return <ChevronRight className={className} {...stroke} />
}

export function IconPlay({className}: IconProps) {
  return (
    <Play
      className={className}
      aria-hidden="true"
      fill="currentColor"
      strokeWidth={0}
    />
  )
}

export function IconFolder({className}: IconProps) {
  return <Folder className={className} {...stroke} />
}

export function IconHelp({className}: IconProps) {
  return <CircleHelp className={className} {...stroke} />
}

export function IconMiru({className}: IconProps) {
  return (
    <svg
      className={className}
      viewBox="0 0 100 100"
      fill="currentColor"
      aria-hidden="true"
    >
      <path d="M4.8 0 L50.1 39.6 L95.2 0 L95.2 100 L75.3 84.6 L75.2 43.9 L50.2 66.2 L25.1 43.9 L24.7 84.7 L4.8 100 Z" />
    </svg>
  )
}
