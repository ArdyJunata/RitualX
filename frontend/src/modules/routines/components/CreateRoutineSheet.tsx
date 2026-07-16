// frontend/src/modules/routines/components/CreateRoutineSheet.tsx

'use client'

import { useCallback } from 'react'
import { BottomSheet } from '@/shared/components/ui/BottomSheet'
import { useCreateRoutine } from '../hooks/useCreateRoutine'
import { StepPeriod } from './steps/StepPeriod'
import { StepTitleIcon } from './steps/StepTitleIcon'
import { StepTarget } from './steps/StepTarget'
import { StepColor } from './steps/StepColor'
import { StepConfirm } from './steps/StepConfirm'
import { PeriodType } from '../types'

interface CreateRoutineSheetProps {
  isOpen: boolean
  onClose: () => void
}

export function CreateRoutineSheet({ isOpen, onClose }: CreateRoutineSheetProps) {
  const handleSuccess = useCallback(() => {
    onClose()
  }, [onClose])

  const {
    step,
    state,
    isLoading,
    error,
    isDirty,
    nextStep,
    prevStep,
    setField,
    submit,
    reset,
  } = useCreateRoutine(handleSuccess)

  function handleClose() {
    if (isDirty) {
      const confirmed = window.confirm('Discard this routine?')
      if (!confirmed) return
    }
    reset()
    onClose()
  }

  return (
    <BottomSheet isOpen={isOpen} onClose={handleClose}>
      {/* Progress dots */}
      <div className="flex justify-center gap-2 py-4 flex-shrink-0">
        {[1, 2, 3, 4, 5].map(n => (
          <div
            key={n}
            className={[
              'w-2 h-2 rounded-full transition-colors duration-200',
              n <= step ? 'bg-emerald-500' : 'bg-zinc-700',
            ].join(' ')}
          />
        ))}
      </div>

      {/* Steps */}
      {step === 1 && (
        <StepPeriod
          selected={state.periodType}
          onSelect={v => setField('periodType', v as PeriodType)}
          onNext={nextStep}
        />
      )}
      {step === 2 && (
        <StepTitleIcon
          title={state.title}
          icon={state.icon}
          onTitleChange={v => setField('title', v)}
          onIconChange={v => setField('icon', v)}
          onNext={nextStep}
          onBack={prevStep}
        />
      )}
      {step === 3 && (
        <StepTarget
          targetCount={state.targetCount}
          periodType={state.periodType!}
          onChange={v => setField('targetCount', v)}
          onNext={nextStep}
          onBack={prevStep}
        />
      )}
      {step === 4 && (
        <StepColor
          color={state.color}
          onSelect={v => setField('color', v)}
          onNext={nextStep}
          onBack={prevStep}
        />
      )}
      {step === 5 && (
        <StepConfirm
          state={state}
          isLoading={isLoading}
          error={error}
          onSubmit={submit}
          onBack={prevStep}
        />
      )}
    </BottomSheet>
  )
}
