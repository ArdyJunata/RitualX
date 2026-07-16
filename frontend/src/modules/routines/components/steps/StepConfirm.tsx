// frontend/src/modules/routines/components/steps/StepConfirm.tsx

'use client'

import { Loader2 } from 'lucide-react'
import { WizardState } from '../../hooks/useCreateRoutine'
import { RoutinePreviewCard } from '../RoutinePreviewCard'
import { PeriodType } from '../../types'

interface StepConfirmProps {
  state: WizardState
  isLoading: boolean
  error: string | null
  onSubmit: () => void
  onBack: () => void
}

export function StepConfirm({ state, isLoading, error, onSubmit, onBack }: StepConfirmProps) {
  return (
    <div className="pt-4">
      <h2 className="text-xl font-bold text-white mb-1">Ready to go! 🎉</h2>
      <p className="text-zinc-400 text-sm mb-6">Here&apos;s your new routine.</p>

      <RoutinePreviewCard
        title={state.title}
        icon={state.icon}
        color={state.color}
        periodType={state.periodType as PeriodType}
        targetCount={state.targetCount}
      />

      {error && (
        <p className="mt-4 text-sm text-red-400 text-center">{error}</p>
      )}

      <div className="flex gap-3 mt-8">
        <button
          type="button"
          onClick={onBack}
          disabled={isLoading}
          className="flex-1 py-3 rounded-xl border border-zinc-700 text-zinc-300 font-semibold text-sm disabled:opacity-40 transition-opacity duration-150"
        >
          ← Back
        </button>
        <button
          type="button"
          onClick={onSubmit}
          disabled={isLoading}
          className="flex-1 py-3 rounded-xl bg-emerald-500 text-white font-semibold text-sm disabled:opacity-40 flex items-center justify-center gap-2 transition-opacity duration-150"
        >
          {isLoading && <Loader2 size={16} className="animate-spin" />}
          Create Routine
        </button>
      </div>
    </div>
  )
}
