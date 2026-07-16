// frontend/src/modules/routines/components/steps/StepTitleIcon.tsx

'use client'

import { useEffect, useRef } from 'react'

const EMOJIS = [
  '🏃', '🏋️', '📚', '💧', '🧘', '🎯',
  '🎮', '🍎', '✍️', '🎵', '🧹', '😴',
  '🚴', '🧠', '💊', '🌿', '🔥', '⚡',
  '🎨', '📷', '🛁', '🍳', '💪', '🌙',
]

interface StepTitleIconProps {
  title: string
  icon: string
  onTitleChange: (v: string) => void
  onIconChange: (v: string) => void
  onNext: () => void
  onBack: () => void
}

export function StepTitleIcon({
  title,
  icon,
  onTitleChange,
  onIconChange,
  onNext,
  onBack,
}: StepTitleIconProps) {
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    // Delay to let sheet animation finish before focusing
    const t = setTimeout(() => inputRef.current?.focus(), 350)
    return () => clearTimeout(t)
  }, [])

  const canNext = title.trim().length > 0 && icon !== ''

  return (
    <div className="pt-4">
      <h2 className="text-xl font-bold text-white mb-1">Name your routine</h2>
      <p className="text-zinc-400 text-sm mb-6">Give it a name and pick an icon.</p>

      <input
        ref={inputRef}
        type="text"
        value={title}
        onChange={e => onTitleChange(e.target.value)}
        placeholder="e.g. Morning Run"
        maxLength={50}
        className="w-full bg-zinc-800 border border-zinc-700 rounded-xl px-4 py-3 text-white placeholder-zinc-500 focus:outline-none focus:border-emerald-500 transition-colors duration-150 mb-6"
      />

      <p className="text-sm text-zinc-400 mb-3">Pick an icon</p>
      <div className="grid grid-cols-6 gap-2">
        {EMOJIS.map(emoji => (
          <button
            key={emoji}
            type="button"
            onClick={() => onIconChange(emoji)}
            className={[
              'text-2xl p-2 rounded-xl border-2 transition-all duration-100',
              icon === emoji
                ? 'border-emerald-500 bg-emerald-500/10'
                : 'border-transparent bg-zinc-800',
            ].join(' ')}
          >
            {emoji}
          </button>
        ))}
      </div>

      <div className="flex gap-3 mt-8">
        <button
          type="button"
          onClick={onBack}
          className="flex-1 py-3 rounded-xl border border-zinc-700 text-zinc-300 font-semibold text-sm"
        >
          ← Back
        </button>
        <button
          type="button"
          onClick={onNext}
          disabled={!canNext}
          className="flex-1 py-3 rounded-xl bg-emerald-500 text-white font-semibold text-sm disabled:opacity-40 disabled:cursor-not-allowed transition-opacity duration-150"
        >
          Next →
        </button>
      </div>
    </div>
  )
}
