import {Dialog as DialogPrimitive} from '@base-ui/react/dialog'
import type * as React from 'react'
import {cn} from '@/lib/utils'

function Root({...props}: DialogPrimitive.Root.Props) {
  return <DialogPrimitive.Root data-slot="dialog" {...props} />
}

function Trigger({...props}: DialogPrimitive.Trigger.Props) {
  return <DialogPrimitive.Trigger data-slot="dialog-trigger" {...props} />
}

function Portal({...props}: DialogPrimitive.Portal.Props) {
  return <DialogPrimitive.Portal data-slot="dialog-portal" {...props} />
}

function Backdrop({
  className,
  ...props
}: DialogPrimitive.Backdrop.Props) {
  return (
    <DialogPrimitive.Backdrop
      data-slot="dialog-backdrop"
      className={cn('fixed inset-0 z-50 bg-bezel/80', className)}
      {...props}
    />
  )
}

function Viewport({
  className,
  ...props
}: DialogPrimitive.Viewport.Props) {
  return (
    <DialogPrimitive.Viewport
      data-slot="dialog-viewport"
      className={cn(
        'fixed inset-0 z-50 flex items-center justify-center p-6',
        className,
      )}
      {...props}
    />
  )
}

function Panel({className, ...props}: DialogPrimitive.Popup.Props) {
  return (
    <DialogPrimitive.Popup
      data-slot="dialog-panel"
      className={cn(
        'max-h-[calc(100vh-3rem)] w-full max-w-lg overflow-y-auto border border-border bg-card p-4 text-foreground',
        className,
      )}
      {...props}
    />
  )
}

function Title({className, ...props}: DialogPrimitive.Title.Props) {
  return (
    <DialogPrimitive.Title
      data-slot="dialog-title"
      className={cn('text-base font-medium', className)}
      {...props}
    />
  )
}

function Description({
  className,
  ...props
}: DialogPrimitive.Description.Props) {
  return (
    <DialogPrimitive.Description
      data-slot="dialog-description"
      className={cn('mt-1 text-sm text-muted-foreground', className)}
      {...props}
    />
  )
}

function Close({className, ...props}: DialogPrimitive.Close.Props) {
  return (
    <DialogPrimitive.Close
      data-slot="dialog-close"
      className={cn('cursor-pointer', className)}
      {...props}
    />
  )
}

export const Dialog = {
  Root,
  Trigger,
  Portal,
  Backdrop,
  Viewport,
  Panel,
  Title,
  Description,
  Close,
}

export type DialogProps = {
  open?: boolean
  onOpenChange?: (open: boolean) => void
  children?: React.ReactNode
}
