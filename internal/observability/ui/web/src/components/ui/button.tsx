import { Slot } from '@radix-ui/react-slot'
import { cva, type VariantProps } from 'class-variance-authority'
import type { ButtonHTMLAttributes } from 'react'
import { cn } from '@/lib/utils'

const buttonVariants = cva(
  'inline-flex items-center justify-center gap-1.5 whitespace-nowrap rounded-(--radius-sm) text-[13px] font-medium transition-[background-color,border-color,color,transform] duration-150 ease-out active:scale-[0.97] disabled:pointer-events-none disabled:opacity-40 focus-visible:outline-2 focus-visible:outline-(--color-focus-ring) focus-visible:outline-offset-2',
  {
    variants: {
      variant: {
        primary:
          'bg-(--color-signal) text-(--color-canvas) hover:brightness-110 font-semibold',
        outline:
          'border border-(--color-hairline-strong) text-(--color-ink-secondary) hover:border-(--color-hairline-strong) hover:text-(--color-ink) hover:bg-(--color-canvas-raised)',
        ghost: 'text-(--color-ink-secondary) hover:bg-(--color-canvas-raised) hover:text-(--color-ink)',
      },
      size: {
        sm: 'h-7 px-2.5',
        md: 'h-8 px-3',
        icon: 'h-7 w-7',
      },
    },
    defaultVariants: {
      variant: 'outline',
      size: 'md',
    },
  },
)

export interface ButtonProps
  extends ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
}

export function Button({ className, variant, size, asChild = false, ...props }: ButtonProps) {
  const Comp = asChild ? Slot : 'button'
  return <Comp className={cn(buttonVariants({ variant, size, className }))} {...props} />
}
