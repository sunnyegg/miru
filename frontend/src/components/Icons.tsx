type IconProps = {
  className?: string
}

export function IconLibrary({className}: IconProps) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" d="M4 5h7v14H4zM13 5h7v14h-7z" />
    </svg>
  )
}

export function IconDownload({className}: IconProps) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v12m0 0 4-4m-4 4-4-4M5 20h14" />
    </svg>
  )
}

export function IconSettings({className}: IconProps) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 15.5a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7Z" />
      <path strokeLinecap="round" strokeLinejoin="round" d="M19.4 13.1c.04-.36.06-.73.06-1.1s-.02-.74-.06-1.1l1.7-1.33-1.6-2.77-2.02.82a7.1 7.1 0 0 0-1.9-1.1L15.2 4h-3.2l-.38 2.52c-.68.26-1.32.63-1.9 1.1l-2.02-.82-1.6 2.77 1.7 1.33c-.04.36-.06.73-.06 1.1s.02.74.06 1.1L4.1 14.43l1.6 2.77 2.02-.82c.58.47 1.22.84 1.9 1.1L12 20h3.2l.38-2.52c.68-.26 1.32-.63 1.9-1.1l2.02.82 1.6-2.77-1.7-1.33Z" />
    </svg>
  )
}

export function IconPlay({className}: IconProps) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
      <path d="M8 5.14v13.72L19 12 8 5.14Z" />
    </svg>
  )
}

export function IconFolder({className}: IconProps) {
  return (
    <svg className={className} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" aria-hidden="true">
      <path strokeLinecap="round" strokeLinejoin="round" d="M3 7h6l2 2h10v10H3V7Z" />
    </svg>
  )
}
