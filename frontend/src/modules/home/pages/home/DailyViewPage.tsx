// frontend/src/modules/home/pages/home/DailyViewPage.tsx

'use client'

import { useState } from 'react'
import { useDailyView } from '../../hooks/useDailyView'
import { DailyHeader } from './DailyHeader'
import { DailyGoalRing } from './DailyGoalRing'
import { RoutineCard } from './RoutineCard'
import { EmptyState } from './EmptyState'
import { CreateRoutineSheet } from '@/modules/routines'

export function DailyViewPage() {
  const { routines, isLoading, error, tap, refresh } = useDailyView()
  const [sheetOpen, setSheetOpen] = useState(false)

  const doneCount = routines.filter(r => r.isDone).length
  const totalCount = routines.length

  function handleCreateSuccess() {
    setSheetOpen(false)
    refresh()
  }

  if (isLoading) {
    return (
      <div className="flex flex-col gap-4 p-4">
        <DailyHeader />
        <div className="flex justify-center py-8">
          <div className="w-24 h-24 rounded-full bg-zinc-800 animate-pulse" />
        </div>
        {[1, 2, 3].map(i => (
          <div key={i} className="h-20 rounded-2xl bg-zinc-800 animate-pulse" />
        ))}
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-4 px-6 text-center">
        <p className="text-zinc-400 text-sm">{error}</p>
        <button
          type="button"
          onClick={refresh}
          className="px-5 py-2 rounded-xl bg-zinc-800 text-white text-sm font-medium active:scale-95 transition-transform"
        >
          Retry
        </button>
      </div>
    )
  }

  return (
    <>
      <div className="flex flex-col gap-5 p-4 pb-28">
        <DailyHeader />

        {totalCount > 0 && (
          <div className="flex justify-center">
            <DailyGoalRing done={doneCount} total={totalCount} />
          </div>
        )}

        {totalCount === 0 ? (
          <EmptyState onCreateClick={() => setSheetOpen(true)} />
        ) : (
          <div className="flex flex-col gap-3">
            {routines.map(routine => (
              <RoutineCard
                key={routine.id}
                routine={routine}
                onTap={() => tap(routine)}
              />
            ))}
          </div>
        )}
      </div>

      <CreateRoutineSheet
        isOpen={sheetOpen}
        onClose={handleCreateSuccess}
      />
    </>
  )
}
