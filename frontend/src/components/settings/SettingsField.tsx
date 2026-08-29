import type {ReactNode} from 'react'
import {LabelWithHint} from '../LabelWithHint'

type Props = {
  label: string
  htmlFor: string
  hint?: ReactNode
  className?: string
  children: ReactNode
}

export function SettingsField({label, htmlFor, hint, className, children}: Props) {
  return (
    <div className={className ?? 'mt-4'}>
      <LabelWithHint htmlFor={htmlFor} label={label} hint={hint} />
      {children}
    </div>
  )
}
