// frontend/src/modules/stats/components/RoutineHeatmapCard.tsx

'use client'

import { useRoutineHeatmap } from '../hooks'
import { ContributionHeatmap } from './ContributionHeatmap'

export interface RoutineHeatmapCardProps {
  routineId: string
  title: string
  color?: string
  year?: number
}

export function RoutineHeatmapCard({ routineId, title, color, year }: RoutineHeatmapCardProps) {
  const yr = year ?? new Date().getUTCFullYear()
  const { data, isLoading, error, refresh } = useRoutineHeatmap(routineId, yr)

  return (
    <div className="glass-card flex flex-col gap-3 p-4">
      <div className="flex items-center justify-between">
        <p className="text-sm font-semibold text-zinc-100">{title}</p>
        <span className="text-xs text-zinc-500">{yr}</span>
      </div>

      {isLoading ? (
        <div className="h-24 w-full animate-pulse rounded-lg bg-zinc-800" />
      ) : error ? (
        <div className="flex items-center gap-3">
          <p className="text-xs text-zinc-400">{error}</p>
          <button
            type="button"
            onClick={refresh}
            className="rounded-lg bg-zinc-800 px-3 py-1 text-xs font-medium text-white active:scale-95"
          >
            Retry
          </button>
        </div>
      ) : (
        <ContributionHeatmap data={data} year={yr} color={color} />
      )}
    </div>
  )
}
