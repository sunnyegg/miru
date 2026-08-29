import type {ReactNode} from 'react'
import {IconHelp} from './Icons'
import {Button} from '@/components/ui/button'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/components/ui/tooltip'

type Props = {
  content: ReactNode
  side?: 'top' | 'right' | 'bottom' | 'left'
}

export function HintTooltip({content, side = 'top'}: Props) {
  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label="More information"
            className="size-11 shrink-0 text-muted-foreground hover:text-foreground"
          />
        }
      >
        <IconHelp className="size-4" />
      </TooltipTrigger>
      <TooltipContent side={side}>{content}</TooltipContent>
    </Tooltip>
  )
}
