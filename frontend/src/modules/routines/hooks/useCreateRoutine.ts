// frontend/src/modules/routines/hooks/useCreateRoutine.ts

'use client'

import { useState, useCallback } from 'react'
import { routinesApi } from '../api'
import { PeriodType, CreateRoutineRequest } from '../types'

export interface WizardState {
  periodType: PeriodType | null
  title: string
  icon: string
  targetCount: number
  color: string
}

const INITIAL_STATE: WizardState = {
  periodType: null,
  title: '',
  icon: '',
  targetCount: 1,
  color: '',
}

export interface UseCreateRoutineReturn {
  step: number
  state: WizardState
  isLoading: boolean
  error: string | null
  isDirty: boolean
  setStep: (s: number) => void
  nextStep: () => void
  prevStep: () => void
  setField: <K extends keyof WizardState>(key: K, value: WizardState[K]) => void
  submit: () => Promise<void>
  reset: () => void
}

export function useCreateRoutine(onSuccess: () => void): UseCreateRoutineReturn {
  const [step, setStep] = useState(1)
  const [state, setState] = useState<WizardState>(INITIAL_STATE)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const isDirty =
    state.periodType !== null ||
    state.title.trim().length > 0 ||
    state.icon !== '' ||
    state.color !== ''

  const setField = useCallback(<K extends keyof WizardState>(key: K, value: WizardState[K]) => {
    setState(prev => ({ ...prev, [key]: value }))
  }, [])

  const nextStep = useCallback(() => setStep(s => Math.min(s + 1, 5)), [])
  const prevStep = useCallback(() => setStep(s => Math.max(s - 1, 1)), [])

  const reset = useCallback(() => {
    setStep(1)
    setState(INITIAL_STATE)
    setError(null)
    setIsLoading(false)
  }, [])

  const submit = useCallback(async () => {
    if (
      !state.periodType ||
      !state.title.trim() ||
      !state.icon ||
      !state.color
    ) return

    setIsLoading(true)
    setError(null)
    try {
      const body: CreateRoutineRequest = {
        title: state.title.trim(),
        period_type: state.periodType,
        target_count: state.targetCount,
        icon: state.icon,
        color: state.color,
      }
      await routinesApi.create(body)
      reset()
      onSuccess()
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to create routine.'
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }, [state, reset, onSuccess])

  return { step, state, isLoading, error, isDirty, setStep, nextStep, prevStep, setField, submit, reset }
}
