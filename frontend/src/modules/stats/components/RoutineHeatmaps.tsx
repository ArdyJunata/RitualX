// frontend/src/modules/stats/components/RoutineHeatmaps.tsx

'use client'

import { useState, useEffect } from 'react'
import { routinesApi, Routine } from '@/modules/routines'
import { RoutineHeatmapCard } from './RoutineHeatmapCard'

export function RoutineHeatmaps() {
  const [routines, setRoutines] = useState<Routine[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    async function load() {
      setIsLoading(true)
      setError(null)
      try {
        const list = await routinesApi.list()
        if (!cancelled) setRoutines(list)
      } catch {
        if (!cancelled) setError('Failed to load routines.')
      } finally {
        if (!cancelled) setIsLoading(false)
      }
    }
    load()
    return () => {
      cancelled = true
    }
  }, [])

  if (isLoading) {
    return <div className="h-40 w-full animate-pulse rounded-2xl bg-zinc-800" />
  }

  if (error) {
    return <p className="text-sm text-zinc-400">{error}</p>
  }

  if (routines.length === 0) {
    return (
      <p className="text-sm text-zinc-500">Create a routine to see its contribution heatmap.</p>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      {routines.map((r) => (
        <RoutineHeatmapCard key={r.id} routineId={r.id} title={r.title} color={r.color} />
      ))}
    </div>
  )
}
