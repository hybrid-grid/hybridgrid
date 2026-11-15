import * as TooltipPrimitive from '@radix-ui/react-tooltip'
import { cn } from '@/lib/utils'

export const TooltipProvider = TooltipPrimitive.Provider
export const Tooltip = TooltipPrimitive.Root
export const TooltipTrigger = TooltipPrimitive.Trigger

export function TooltipContent({
  className,
  sideOffset = 6,
  ...props
}: React.ComponentProps<typeof TooltipPrimitive.Content>) {
  return (
    <TooltipPrimitive.Portal>
      <TooltipPrimitive.Content
        sideOffset={sideOffset}
        className={cn(
          'z-50 rounded-(--radius-xs) border border-(--color-hairline-strong) bg-(--color-canvas-overlay) px-2 py-1 text-[11px] font-medium text-(--color-ink-secondary) shadow-[0_0_0_1px_rgba(0,0,0,0.4)]',
          className,
        )}
        {...props}
      />
    </TooltipPrimitive.Portal>
  )
}
