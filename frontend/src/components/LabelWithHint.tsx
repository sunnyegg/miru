import type {ReactNode} from 'react'
import {HintTooltip} from './HintTooltip'
import {Label} from '@/components/ui/label'

type Props = {
  htmlFor: string
  label: string
  hint?: ReactNode
  className?: string
}

export function LabelWithHint({htmlFor, label, hint, className}: Props) {
  return (
    <div className={className ?? 'mb-2 flex items-center gap-1'}>
      <Label htmlFor={htmlFor}>{label}</Label>
      {hint ? <HintTooltip content={hint} /> : null}
    </div>
  )
}
