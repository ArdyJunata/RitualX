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
  const unitLabel =
    periodType === 'daily' ? 'day' : periodType === 'weekly' ? 'week' : 'month'

  return (
    <div className="bg-zinc-800 border border-zinc-700 rounded-2xl p-4 flex items-center gap-4">
      {/* Icon */}
      <span className="text-3xl flex-shrink-0">{icon}</span>

      {/* Info */}
      <div className="flex-1 min-w-0">
        <p className="font-bold text-white truncate">{title}</p>
        <p className="text-xs text-zinc-400 mt-0.5">
          {periodLabel} · {targetCount}× per {unitLabel}
        </p>

        {/* Empty progress bar */}
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
