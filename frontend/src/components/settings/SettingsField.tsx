import type {ReactNode} from 'react'
import {Label} from '@/components/ui/label'

type Props = {
  label: string
  htmlFor: string
  className?: string
  children: ReactNode
}

export function SettingsField({label, htmlFor, className, children}: Props) {
  return (
    <div className={className ?? 'mt-4'}>
      <Label htmlFor={htmlFor} className="mb-2">
        {label}
      </Label>
      {children}
    </div>
  )
}
