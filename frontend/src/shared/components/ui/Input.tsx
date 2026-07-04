// frontend/src/shared/components/ui/Input.tsx

import { InputHTMLAttributes } from 'react'
import { clsx } from 'clsx'

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string
  error?: string
}

export function Input({ label, error, id, className, ...props }: InputProps) {
  const errorId = error && id ? `${id}-error` : undefined

  return (
    <div className="flex flex-col gap-1">
      <label
        htmlFor={id}
        className="text-sm font-medium text-zinc-300"
      >
        {label}
      </label>
      <input
        id={id}
        aria-describedby={errorId}
        aria-invalid={!!error}
        className={clsx(
          'w-full rounded-lg bg-zinc-800 border px-4 py-2.5 text-sm text-white placeholder-zinc-500',
          'focus:outline-none focus:ring-2 focus:ring-emerald focus:border-transparent',
          'transition-colors duration-150',
          error ? 'border-red' : 'border-zinc-700',
          className,
        )}
        {...props}
      />
      {error && (
        <p id={errorId} role="alert" className="text-red text-xs mt-0.5">{error}</p>
      )}
    </div>
  )
}
