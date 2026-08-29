import type {ReactNode} from 'react'
import {HintTooltip} from '../HintTooltip'
import {Label} from '@/components/ui/label'
import {cn} from '@/lib/utils'

type Props = {
  id: string
  label: string
  hint?: ReactNode
  checked: boolean
  onChange: (checked: boolean) => void
  className?: string
}

export function SettingsCheckboxRow({id, label, hint, checked, onChange, className}: Props) {
  return (
    <div className={cn('flex min-h-11 items-center gap-3', className)}>
      <input
        id={id}
        type="checkbox"
        checked={checked}
        onChange={(event) => onChange(event.target.checked)}
        className="size-4 shrink-0 accent-primary"
      />
      <Label htmlFor={id} className="mb-0">
        {label}
      </Label>
      {hint ? <HintTooltip content={hint} /> : null}
    </div>
  )
}
