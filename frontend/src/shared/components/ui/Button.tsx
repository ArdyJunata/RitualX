// frontend/src/shared/components/ui/Button.tsx

import { ButtonHTMLAttributes } from 'react'
import { clsx } from 'clsx'
import { Loader2 } from 'lucide-react'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  isLoading?: boolean
  children: React.ReactNode
}

export function Button({ isLoading, disabled, children, className, ...props }: ButtonProps) {
  return (
    <button
      disabled={disabled || isLoading}
      className={clsx(
        'w-full flex items-center justify-center gap-2 rounded-lg px-4 py-2.5',
        'bg-emerald text-white font-semibold text-sm',
        'hover:bg-emerald-600 transition-colors duration-150',
        'focus:outline-none focus:ring-2 focus:ring-emerald focus:ring-offset-2 focus:ring-offset-zinc-950',
        'disabled:opacity-50 disabled:cursor-not-allowed',
        className,
      )}
      {...props}
    >
      {isLoading && <Loader2 size={16} className="animate-spin" />}
      {children}
    </button>
  )
}
