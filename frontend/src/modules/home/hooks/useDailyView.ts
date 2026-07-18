// frontend/src/modules/home/hooks/useDailyView.ts

'use client'

import { useState, useEffect, useCallback } from 'react'
import { routinesApi } from '@/modules/routines'
import { DailyRoutine, Routine, RoutineLog } from '@/modules/routines'

import { useRouter } from 'next/navigation'

function toDailyRoutine(routine: Routine, todayLog: RoutineLog | null): DailyRoutine {
  const todayCount = todayLog?.count ?? 0
  return {
    ...routine,
    todayLog,
    todayCount,
    isDone: todayCount >= routine.target_count,
  }
}

export interface UseDailyViewReturn {
  routines: DailyRoutine[]
  isLoading: boolean
  error: string | null
  tap: (routine: DailyRoutine) => Promise<void>
  refresh: () => Promise<void>
}

export function useDailyView(): UseDailyViewReturn {
  const [routines, setRoutines] = useState<DailyRoutine[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const router = useRouter()

  const load = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const list = await routinesApi.list()
      const withLogs = await Promise.all(
        list.map(async (r) => {
          const todayLog = await routinesApi.getTodayLog(r.id)
          return toDailyRoutine(r, todayLog)
        })
      )
      withLogs.sort((a, b) => a.sort_order - b.sort_order)
      setRoutines(withLogs)
    } catch (err: unknown) {
      if (err && typeof err === 'object' && 'status' in err && (err as { status: number }).status === 401) {
        router.push('/login')
        return
      }
      setError('Failed to load routines. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }, [router])


  useEffect(() => {
    load()
    
    const onRoutineCreated = () => load()
    window.addEventListener('routineCreated', onRoutineCreated)
    return () => window.removeEventListener('routineCreated', onRoutineCreated)
  }, [load])

  const tap = useCallback(async (routine: DailyRoutine) => {
    const snapshot = [...routines]

    if (!routine.isDone) {
      // Optimistic: increment count
      setRoutines(prev =>
        prev.map(r =>
          r.id === routine.id
            ? { ...r, todayCount: r.todayCount + 1, isDone: r.todayCount + 1 >= r.target_count }
            : r
        )
      )
      try {
        const log = await routinesApi.logRoutine(routine.id)
        setRoutines(prev =>
          prev.map(r =>
            r.id === routine.id
              ? { ...r, todayLog: log, todayCount: log.count, isDone: log.count >= r.target_count }
              : r
          )
        )
      } catch {
        setRoutines(snapshot)
      }
    } else {
      // Optimistic: reset to 0 (undo = delete entire day's log)
      if (!routine.todayLog) return
      const logId = routine.todayLog.id
      setRoutines(prev =>
        prev.map(r =>
          r.id === routine.id
            ? { ...r, todayLog: null, todayCount: 0, isDone: false }
            : r
        )
      )
      try {
        await routinesApi.deleteLog(routine.id, logId)
      } catch {
        setRoutines(snapshot)
      }
    }
  }, [routines])

  return { routines, isLoading, error, tap, refresh: load }
}
