// frontend/src/modules/routines/components/steps/StepPeriod.tsx

'use client'

import { PeriodType } from '../../types'

interface StepPeriodProps {
  selected: PeriodType | null
  onSelect: (p: PeriodType) => void
  onNext: () => void
}

const OPTIONS: { value: PeriodType; label: string; emoji: string; desc: string }[] = [
  { value: 'daily',   label: 'Daily',   emoji: '📅', desc: 'Every day' },
  { value: 'weekly',  label: 'Weekly',  emoji: '📆', desc: 'Every week' },
  { value: 'monthly', label: 'Monthly', emoji: '🗓️', desc: 'Every month' },
]

export function StepPeriod({ selected, onSelect, onNext }: StepPeriodProps) {
  return (
    <div className="pt-4">
      <h2 className="text-xl font-bold text-white mb-1">What kind of routine?</h2>
      <p className="text-zinc-400 text-sm mb-6">Pick how often you want to track it.</p>

      <div className="flex flex-col gap-3">
        {OPTIONS.map(opt => {
          const isSelected = selected === opt.value
          return (
            <button
              key={opt.value}
              type="button"
              onClick={() => onSelect(opt.value)}
              className={[
                'flex items-center gap-4 p-4 rounded-xl border-2 text-left transition-all duration-150',
                isSelected
                  ? 'border-emerald-500 bg-emerald-500/10'
                  : 'border-zinc-700 bg-zinc-800/50',
              ].join(' ')}
            >
              <span className="text-2xl">{opt.emoji}</span>
              <div>
                <p className="font-semibold text-white">{opt.label}</p>
                <p className="text-xs text-zinc-400">{opt.desc}</p>
              </div>
              {isSelected && <span className="ml-auto text-emerald-400 text-lg">✓</span>}
            </button>
          )
        })}
      </div>

      <div className="mt-8">
        <button
          type="button"
          onClick={onNext}
          disabled={!selected}
          className="w-full py-3 rounded-xl bg-emerald-500 text-white font-semibold text-sm disabled:opacity-40 disabled:cursor-not-allowed transition-opacity duration-150"
        >
          Next →
        </button>
      </div>
    </div>
  )
}
