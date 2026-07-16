// frontend/src/modules/routines/components/steps/StepTarget.tsx

'use client'

import { PeriodType } from '../../types'

interface StepTargetProps {
  targetCount: number
  periodType: PeriodType
  onChange: (v: number) => void
  onNext: () => void
  onBack: () => void
}

export function StepTarget({ targetCount, periodType, onChange, onNext, onBack }: StepTargetProps) {
  const periodLabel =
    periodType === 'daily' ? 'day' : periodType === 'weekly' ? 'week' : 'month'

  return (
    <div className="pt-4">
      <h2 className="text-xl font-bold text-white mb-1">Set your target</h2>
      <p className="text-zinc-400 text-sm mb-8">
        How many times per {periodLabel}?
      </p>

      <div className="flex items-center justify-center gap-8">
        <button
          type="button"
          onClick={() => onChange(Math.max(1, targetCount - 1))}
          className="w-14 h-14 rounded-full bg-zinc-800 border border-zinc-700 text-white text-2xl font-bold flex items-center justify-center active:scale-95 transition-transform duration-100"
        >
          −
        </button>
        <span className="text-6xl font-bold text-white w-20 text-center tabular-nums">
          {targetCount}
        </span>
        <button
          type="button"
          onClick={() => onChange(Math.min(99, targetCount + 1))}
          className="w-14 h-14 rounded-full bg-zinc-800 border border-zinc-700 text-white text-2xl font-bold flex items-center justify-center active:scale-95 transition-transform duration-100"
        >
          +
        </button>
      </div>

      <p className="text-center text-zinc-400 text-sm mt-4">
        {targetCount}× per {periodLabel}
      </p>

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
          className="flex-1 py-3 rounded-xl bg-emerald-500 text-white font-semibold text-sm"
        >
          Next →
        </button>
      </div>
    </div>
  )
}
