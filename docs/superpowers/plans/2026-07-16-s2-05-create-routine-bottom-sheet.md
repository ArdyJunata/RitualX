# S2-05 Create Routine Bottom Sheet — Implementation Plan

> **For agentic workers:** Steps use checkbox (`- [ ]`) syntax for tracking. Execute tasks in order. Each task ends with a commit.

**Goal:** Build a 5-step wizard bottom sheet that lets users create a routine, wired to `POST /api/v1/routines`.

**Architecture:** Generic `BottomSheet` shared UI component. Wizard state + API mutation in `useCreateRoutine` hook inside a new `routines` module. Step components are pure presentational. FAB `+` button added to `BottomNav`.

**Tech Stack:** Next.js 14 App Router, React, TypeScript, Tailwind CSS, TanStack Query (`@tanstack/react-query`), Lucide React, `clsx`.

## Global Constraints

- Mobile-first: max container width `430px`.
- Dark theme only. Background color `bg-zinc-950`. Accent color `emerald` (existing token).
- All new components must be `"use client"` where they use state or browser APIs.
- API base URL from `@/shared/api-client` (`apiClient.post`).
- Access token sent via `Authorization: Bearer` header — read from `localStorage.getItem('ritualx_access_token')`.
- No external animation libraries — CSS transitions only.
- Follow existing file patterns: styles in a `.styles.ts` sibling, barrel `index.ts` in each component folder.

---

## Task 1: Routine domain types + API client

**Files:**
- Create: `frontend/src/modules/routines/types.ts`
- Create: `frontend/src/modules/routines/api.ts`
- Create: `frontend/src/modules/routines/index.ts`

**Interfaces:**
- Produces: `Routine`, `CreateRoutineRequest`, `routinesApi.create()`

---

- [ ] **Step 1: Create `types.ts`**

```typescript
// frontend/src/modules/routines/types.ts

export type PeriodType = 'daily' | 'weekly' | 'monthly'

export interface Routine {
  id: string
  user_id: string
  title: string
  period_type: PeriodType
  target_count: number
  icon: string
  color: string
  sort_order: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CreateRoutineRequest {
  title: string
  period_type: PeriodType
  target_count: number
  icon: string
  color: string
}
```

- [ ] **Step 2: Create `api.ts`**

```typescript
// frontend/src/modules/routines/api.ts

import { apiClient } from '@/shared/api-client'
import { CreateRoutineRequest, Routine } from './types'

export const routinesApi = {
  create(body: CreateRoutineRequest): Promise<Routine> {
    const token = localStorage.getItem('ritualx_access_token') ?? ''
    return apiClient.post<Routine>('/api/v1/routines', body, {
      credentials: 'include',
      headers: { Authorization: `Bearer ${token}` },
    })
  },
}
```

- [ ] **Step 3: Create `index.ts` barrel**

```typescript
// frontend/src/modules/routines/index.ts

export * from './types'
export * from './api'
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/modules/routines/types.ts \
        frontend/src/modules/routines/api.ts \
        frontend/src/modules/routines/index.ts
git commit -m "feat(routines): add domain types and API client"
```

---

## Task 2: `useCreateRoutine` hook

**Files:**
- Create: `frontend/src/modules/routines/hooks/useCreateRoutine.ts`

**Interfaces:**
- Consumes: `routinesApi.create(body: CreateRoutineRequest): Promise<Routine>` from Task 1
- Produces:
  - `useCreateRoutine(): UseCreateRoutineReturn`
  - `WizardState` — the shape of all form data
  - `UseCreateRoutineReturn` — all state + handlers the sheet needs

---

- [ ] **Step 1: Create the hook**

```typescript
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
  step: number               // 1–5
  state: WizardState
  isLoading: boolean
  error: string | null
  isDirty: boolean           // true if any field touched
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
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/modules/routines/hooks/useCreateRoutine.ts
git commit -m "feat(routines): add useCreateRoutine wizard hook"
```

---

## Task 3: Generic `BottomSheet` component

**Files:**
- Create: `frontend/src/shared/components/ui/BottomSheet.tsx`

**Interfaces:**
- Produces:
  ```typescript
  interface BottomSheetProps {
    isOpen: boolean
    onClose: () => void
    children: React.ReactNode
  }
  function BottomSheet(props: BottomSheetProps): JSX.Element | null
  ```

---

- [ ] **Step 1: Create `BottomSheet.tsx`**

```typescript
// frontend/src/shared/components/ui/BottomSheet.tsx

'use client'

import { useEffect } from 'react'

interface BottomSheetProps {
  isOpen: boolean
  onClose: () => void
  children: React.ReactNode
}

export function BottomSheet({ isOpen, onClose, children }: BottomSheetProps) {
  // Lock body scroll when open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
    return () => { document.body.style.overflow = '' }
  }, [isOpen])

  return (
    <>
      {/* Backdrop */}
      <div
        aria-hidden="true"
        onClick={onClose}
        className={[
          'fixed inset-0 z-40 bg-black/60 backdrop-blur-sm',
          'transition-opacity duration-300',
          isOpen ? 'opacity-100 pointer-events-auto' : 'opacity-0 pointer-events-none',
        ].join(' ')}
      />

      {/* Sheet */}
      <div
        role="dialog"
        aria-modal="true"
        className={[
          'fixed bottom-0 left-0 right-0 z-50',
          'max-w-[430px] mx-auto',
          'bg-zinc-900 rounded-t-2xl',
          'flex flex-col',
          'transition-transform duration-300 ease-out',
          isOpen ? 'translate-y-0' : 'translate-y-full',
        ].join(' ')}
        style={{ height: '85dvh' }}
      >
        {/* Drag handle */}
        <div className="flex justify-center pt-3 pb-1 flex-shrink-0">
          <div className="w-10 h-1 rounded-full bg-zinc-700" />
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto px-5 pb-8">
          {children}
        </div>
      </div>
    </>
  )
}
```

- [ ] **Step 2: Verify it renders** — temporarily import into any page, wrap content with it, confirm it opens/closes without errors. Remove the temporary usage after.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/shared/components/ui/BottomSheet.tsx
git commit -m "feat(ui): add generic BottomSheet component"
```

---

## Task 4: Step components (pure UI)

**Files:**
- Create: `frontend/src/modules/routines/components/steps/StepPeriod.tsx`
- Create: `frontend/src/modules/routines/components/steps/StepTitleIcon.tsx`
- Create: `frontend/src/modules/routines/components/steps/StepTarget.tsx`
- Create: `frontend/src/modules/routines/components/steps/StepColor.tsx`

**Interfaces:**
- Consumes: `WizardState`, `PeriodType` from Tasks 1–2
- Produces: 4 step components, each accepting props + `onNext` / `onBack` callbacks

---

### Step components shared UI pattern

Each step renders:
1. A heading at the top.
2. The input (options / text / stepper / swatches).
3. A footer with Back (if step > 1) and Next buttons.

Use this footer pattern in every step:

```tsx
<div className="flex gap-3 mt-8">
  {showBack && (
    <button
      type="button"
      onClick={onBack}
      className="flex-1 py-3 rounded-xl border border-zinc-700 text-zinc-300 font-semibold text-sm"
    >
      ← Back
    </button>
  )}
  <button
    type="button"
    onClick={onNext}
    disabled={!canNext}
    className="flex-1 py-3 rounded-xl bg-emerald-500 text-white font-semibold text-sm disabled:opacity-40 disabled:cursor-not-allowed"
  >
    Next →
  </button>
</div>
```

---

- [ ] **Step 1: Create `StepPeriod.tsx`**

```typescript
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
          className="w-full py-3 rounded-xl bg-emerald-500 text-white font-semibold text-sm disabled:opacity-40 disabled:cursor-not-allowed"
        >
          Next →
        </button>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Create `StepTitleIcon.tsx`**

```typescript
// frontend/src/modules/routines/components/steps/StepTitleIcon.tsx

'use client'

import { useEffect, useRef } from 'react'

const EMOJIS = [
  '🏃','🏋️','📚','💧','🧘','🎯',
  '🎮','🍎','✍️','🎵','🧹','😴',
  '🚴','🧠','💊','🌿','🔥','⚡',
  '🎨','📷','🛁','🍳','💪','🌙',
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
  title, icon, onTitleChange, onIconChange, onNext, onBack,
}: StepTitleIconProps) {
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    inputRef.current?.focus()
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
        className="w-full bg-zinc-800 border border-zinc-700 rounded-xl px-4 py-3 text-white placeholder-zinc-500 focus:outline-none focus:border-emerald-500 mb-6"
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
          className="flex-1 py-3 rounded-xl bg-emerald-500 text-white font-semibold text-sm disabled:opacity-40 disabled:cursor-not-allowed"
        >
          Next →
        </button>
      </div>
    </div>
  )
}
```

- [ ] **Step 3: Create `StepTarget.tsx`**

```typescript
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
  const periodLabel = periodType === 'daily' ? 'day' : periodType === 'weekly' ? 'week' : 'month'

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
          className="w-14 h-14 rounded-full bg-zinc-800 border border-zinc-700 text-white text-2xl font-bold flex items-center justify-center"
        >
          −
        </button>
        <span className="text-5xl font-bold text-white w-16 text-center">{targetCount}</span>
        <button
          type="button"
          onClick={() => onChange(Math.min(99, targetCount + 1))}
          className="w-14 h-14 rounded-full bg-zinc-800 border border-zinc-700 text-white text-2xl font-bold flex items-center justify-center"
        >
          +
        </button>
      </div>

      <p className="text-center text-zinc-400 text-sm mt-4">
        {targetCount}x per {periodLabel}
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
```

- [ ] **Step 4: Create `StepColor.tsx`**

```typescript
// frontend/src/modules/routines/components/steps/StepColor.tsx

'use client'

const COLORS = [
  '#ef4444', '#f97316', '#eab308', '#22c55e', '#10b981', '#06b6d4',
  '#3b82f6', '#8b5cf6', '#ec4899', '#f43f5e', '#a3e635', '#e4e4e7',
]

interface StepColorProps {
  color: string
  onSelect: (c: string) => void
  onNext: () => void
  onBack: () => void
}

export function StepColor({ color, onSelect, onNext, onBack }: StepColorProps) {
  return (
    <div className="pt-4">
      <h2 className="text-xl font-bold text-white mb-1">Pick a color</h2>
      <p className="text-zinc-400 text-sm mb-6">
        Used for your heatmap and routine card.
      </p>

      <div className="grid grid-cols-6 gap-3">
        {COLORS.map(hex => (
          <button
            key={hex}
            type="button"
            onClick={() => onSelect(hex)}
            style={{ backgroundColor: hex }}
            className={[
              'w-full aspect-square rounded-full transition-all duration-100',
              color === hex
                ? 'ring-2 ring-white ring-offset-2 ring-offset-zinc-900 scale-110'
                : '',
            ].join(' ')}
            aria-label={hex}
          />
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
          disabled={!color}
          className="flex-1 py-3 rounded-xl bg-emerald-500 text-white font-semibold text-sm disabled:opacity-40 disabled:cursor-not-allowed"
        >
          Next →
        </button>
      </div>
    </div>
  )
}
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/modules/routines/components/steps/
git commit -m "feat(routines): add wizard step components (Period, TitleIcon, Target, Color)"
```

---

## Task 5: `RoutinePreviewCard` component

**Files:**
- Create: `frontend/src/modules/routines/components/RoutinePreviewCard.tsx`

**Interfaces:**
- Consumes: `WizardState` from Task 2
- Produces:
  ```typescript
  interface RoutinePreviewCardProps {
    title: string
    icon: string
    color: string
    periodType: PeriodType
    targetCount: number
  }
  function RoutinePreviewCard(props: RoutinePreviewCardProps): JSX.Element
  ```

---

- [ ] **Step 1: Create `RoutinePreviewCard.tsx`**

```typescript
// frontend/src/modules/routines/components/RoutinePreviewCard.tsx

import { PeriodType } from '../types'

interface RoutinePreviewCardProps {
  title: string
  icon: string
  color: string
  periodType: PeriodType
  targetCount: number
}

export function RoutinePreviewCard({
  title,
  icon,
  color,
  periodType,
  targetCount,
}: RoutinePreviewCardProps) {
  const periodLabel =
    periodType === 'daily' ? 'Daily' : periodType === 'weekly' ? 'Weekly' : 'Monthly'

  return (
    <div className="bg-zinc-800 border border-zinc-700 rounded-2xl p-4 flex items-center gap-4">
      {/* Icon */}
      <span className="text-3xl">{icon}</span>

      {/* Info */}
      <div className="flex-1 min-w-0">
        <p className="font-bold text-white truncate">{title}</p>
        <p className="text-xs text-zinc-400 mt-0.5">
          {periodLabel} · {targetCount}x per{' '}
          {periodType === 'daily' ? 'day' : periodType === 'weekly' ? 'week' : 'month'}
        </p>
        {/* Progress bar (empty — 0 / targetCount) */}
        <div className="mt-2 h-1.5 bg-zinc-700 rounded-full overflow-hidden">
          <div className="h-full w-0 rounded-full" style={{ backgroundColor: color }} />
        </div>
        <p className="text-xs text-zinc-500 mt-1">0 / {targetCount} today</p>
      </div>

      {/* Color dot */}
      <div
        className="w-3 h-3 rounded-full flex-shrink-0"
        style={{ backgroundColor: color }}
      />
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/modules/routines/components/RoutinePreviewCard.tsx
git commit -m "feat(routines): add RoutinePreviewCard component"
```

---

## Task 6: `StepConfirm` component

**Files:**
- Create: `frontend/src/modules/routines/components/steps/StepConfirm.tsx`

**Interfaces:**
- Consumes: `RoutinePreviewCard` from Task 5, `WizardState` from Task 2
- Produces: `StepConfirm` component

---

- [ ] **Step 1: Create `StepConfirm.tsx`**

```typescript
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
          className="flex-1 py-3 rounded-xl border border-zinc-700 text-zinc-300 font-semibold text-sm disabled:opacity-40"
        >
          ← Back
        </button>
        <button
          type="button"
          onClick={onSubmit}
          disabled={isLoading}
          className="flex-1 py-3 rounded-xl bg-emerald-500 text-white font-semibold text-sm disabled:opacity-40 flex items-center justify-center gap-2"
        >
          {isLoading && <Loader2 size={16} className="animate-spin" />}
          Create Routine
        </button>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/modules/routines/components/steps/StepConfirm.tsx
git commit -m "feat(routines): add StepConfirm component"
```

---

## Task 7: `CreateRoutineSheet` — assembles everything

**Files:**
- Create: `frontend/src/modules/routines/components/CreateRoutineSheet.tsx`

**Interfaces:**
- Consumes: `BottomSheet` (Task 3), all Step components (Tasks 4, 6), `useCreateRoutine` (Task 2)
- Produces:
  ```typescript
  interface CreateRoutineSheetProps {
    isOpen: boolean
    onClose: () => void
  }
  function CreateRoutineSheet(props: CreateRoutineSheetProps): JSX.Element
  ```

---

- [ ] **Step 1: Create `CreateRoutineSheet.tsx`**

```typescript
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

  const { step, state, isLoading, error, isDirty, nextStep, prevStep, setField, submit, reset } =
    useCreateRoutine(handleSuccess)

  function handleClose() {
    if (isDirty) {
      const confirmed = window.confirm('Discard this routine?')
      if (!confirmed) return
    }
    reset()
    onClose()
  }

  // Stepper dots
  const dots = Array.from({ length: 5 }, (_, i) => i + 1)

  return (
    <BottomSheet isOpen={isOpen} onClose={handleClose}>
      {/* Progress dots */}
      <div className="flex justify-center gap-2 py-4">
        {dots.map(n => (
          <div
            key={n}
            className={[
              'w-2 h-2 rounded-full transition-colors duration-150',
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
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/modules/routines/components/CreateRoutineSheet.tsx
git commit -m "feat(routines): add CreateRoutineSheet wizard container"
```

---

## Task 8: Wire FAB `+` button into `BottomNav` + update `routines` module barrel

**Files:**
- Modify: `frontend/src/shared/components/ui/BottomNav/BottomNav.tsx`
- Create: `frontend/src/modules/routines/components/index.ts`
- Modify: `frontend/src/modules/routines/index.ts`

**Interfaces:**
- Consumes: `CreateRoutineSheet` (Task 7)

---

- [ ] **Step 1: Create component barrel `frontend/src/modules/routines/components/index.ts`**

```typescript
// frontend/src/modules/routines/components/index.ts

export { CreateRoutineSheet } from './CreateRoutineSheet'
export { RoutinePreviewCard } from './RoutinePreviewCard'
```

- [ ] **Step 2: Update module barrel `frontend/src/modules/routines/index.ts`**

```typescript
// frontend/src/modules/routines/index.ts

export * from './types'
export * from './api'
export * from './components'
```

- [ ] **Step 3: Update `BottomNav.tsx`** to add FAB + open state + render `CreateRoutineSheet`

Replace the entire file content:

```typescript
// frontend/src/shared/components/ui/BottomNav/BottomNav.tsx

'use client'

import { useState } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Home, CalendarDays, BarChart2, Sword, User, Plus } from 'lucide-react'
import { bottomNavStyles as s } from './BottomNav.styles'
import { CreateRoutineSheet } from '@/modules/routines'

const NAV_TABS = [
  { label: 'Home',     href: '/',         Icon: Home         },
  { label: 'Calendar', href: '/calendar', Icon: CalendarDays },
  { label: 'Stats',    href: '/stats',    Icon: BarChart2    },
  { label: 'Quest',    href: '/quest',    Icon: Sword        },
  { label: 'Profile',  href: '/profile',  Icon: User         },
] as const

function isActive(pathname: string, href: string): boolean {
  if (href === '/') return pathname === '/'
  return pathname.startsWith(href)
}

export function BottomNav() {
  const pathname = usePathname()
  const [sheetOpen, setSheetOpen] = useState(false)

  // Split nav tabs into left 2 and right 2 to place FAB in center
  const leftTabs = NAV_TABS.slice(0, 2)
  const rightTabs = NAV_TABS.slice(3, 5)

  return (
    <>
      <nav aria-label="Main navigation" className={s.container}>
        {/* Left tabs */}
        {leftTabs.map(({ label, href, Icon }) => {
          const active = isActive(pathname, href)
          return (
            <Link
              key={href}
              href={href}
              aria-label={label}
              title={label}
              aria-current={active ? 'page' : undefined}
              className={s.tab}
            >
              <Icon className={`${s.icon} ${active ? s.iconActive : s.iconInactive}`} />
              <span className={`${s.dot} ${active ? s.dotActive : s.dotInactive}`} />
            </Link>
          )
        })}

        {/* FAB center button */}
        <button
          type="button"
          aria-label="Create routine"
          onClick={() => setSheetOpen(true)}
          className="flex items-center justify-center w-12 h-12 rounded-full bg-emerald-500 shadow-lg shadow-emerald-500/40 -mt-5 transition-transform duration-150 active:scale-95"
        >
          <Plus size={22} className="text-white" />
        </button>

        {/* Right tabs */}
        {rightTabs.map(({ label, href, Icon }) => {
          const active = isActive(pathname, href)
          return (
            <Link
              key={href}
              href={href}
              aria-label={label}
              title={label}
              aria-current={active ? 'page' : undefined}
              className={s.tab}
            >
              <Icon className={`${s.icon} ${active ? s.iconActive : s.iconInactive}`} />
              <span className={`${s.dot} ${active ? s.dotActive : s.dotInactive}`} />
            </Link>
          )
        })}
      </nav>

      {/* Sheet renders as portal via fixed positioning */}
      <CreateRoutineSheet
        isOpen={sheetOpen}
        onClose={() => setSheetOpen(false)}
      />
    </>
  )
}
```

> **Note:** The center tab (Stats at index 2) is now replaced by the FAB. `leftTabs` = Home + Calendar, `rightTabs` = Quest + Profile. Stats tab is removed from nav — it remains accessible via route but not in the bottom bar for now. If Stats tab must stay, adjust the split and add a 6th slot or add Stats to one side.

- [ ] **Step 4: Start dev server and visually verify**

```bash
cd frontend && npm run dev
```

Open `http://localhost:3000` in browser (mobile viewport, e.g. 390×844 in DevTools).

Check:
1. BottomNav renders with FAB `+` center button (elevated emerald circle).
2. Tap `+` → sheet slides up from bottom.
3. Step 1 shows three period options.
4. Select period → tap Next → Step 2 appears.
5. Fill title + icon → Next → Step 3 stepper.
6. Adjust count → Next → Step 4 color swatches.
7. Select color → Next → Step 5 preview card.
8. Tap "Create Routine" → (may fail if backend not running — that's OK, error message shows).
9. Tap backdrop or drag → sheet closes (confirm dialog if data entered).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/modules/routines/components/index.ts \
        frontend/src/modules/routines/index.ts \
        frontend/src/shared/components/ui/BottomNav/BottomNav.tsx
git commit -m "feat(routines): wire CreateRoutineSheet into BottomNav FAB"
```

---

## Task 9: Add hooks barrel (cleanup)

**Files:**
- Create: `frontend/src/modules/routines/hooks/index.ts`

---

- [ ] **Step 1: Create hooks barrel**

```typescript
// frontend/src/modules/routines/hooks/index.ts

export { useCreateRoutine } from './useCreateRoutine'
export type { WizardState, UseCreateRoutineReturn } from './useCreateRoutine'
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/modules/routines/hooks/index.ts
git commit -m "chore(routines): add hooks barrel export"
```

---

## Done ✅

At the end of all tasks, the following works:

- FAB `+` in BottomNav opens the 5-step Create Routine sheet.
- All 5 steps render correctly with proper validation (Next disabled until required fields filled).
- Submitting calls `POST /api/v1/routines` with the access token.
- Success closes the sheet.
- Error shows below the Create button without closing.
- Discard confirmation fires if user tries to close with data entered.
- `RoutinePreviewCard` is ready to reuse in S2-06.
