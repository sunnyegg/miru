import { Button as ButtonPrimitive } from "@base-ui/react/button"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex shrink-0 cursor-pointer items-center justify-center gap-2 border border-transparent text-sm whitespace-nowrap transition-colors duration-200 outline-none select-none disabled:cursor-not-allowed disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
  {
    variants: {
      variant: {
        default: "bg-primary px-5 font-medium text-primary-foreground",
        secondary: "bg-secondary px-4 text-secondary-foreground hover:bg-muted",
        muted: "bg-muted px-3 text-foreground hover:bg-secondary",
        ghost: "px-3 text-muted-foreground hover:text-foreground",
        destructive: "bg-destructive px-4 text-destructive-foreground",
        outline: "border-border bg-background px-4 hover:bg-muted hover:text-foreground",
        link: "text-primary underline-offset-4 hover:underline",
      },
      size: {
        default: "h-11 min-h-11",
        sm: "h-11 min-h-11",
        lg: "h-11 min-h-11",
        icon: "size-11",
        "icon-sm": "size-11",
        "icon-lg": "size-11",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

function Button({
  className,
  variant = "default",
  size = "default",
  ...props
}: ButtonPrimitive.Props & VariantProps<typeof buttonVariants>) {
  return (
    <ButtonPrimitive
      data-slot="button"
      className={cn(buttonVariants({ variant, size, className }))}
      {...props}
    />
  )
}

export { Button, buttonVariants }
