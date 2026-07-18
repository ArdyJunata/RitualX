// frontend/src/modules/home/pages/home/RoutineCard.tsx

'use client'

import { CheckCircle } from 'lucide-react'
import { DailyRoutine } from '@/modules/routines'

interface RoutineCardProps {
  routine: DailyRoutine
  onTap: () => void
}

export function RoutineCard({ routine, onTap }: RoutineCardProps) {
  const { title, icon, color, todayCount, target_count, isDone } = routine
  const progress = Math.min(todayCount / target_count, 1)

  return (
    <button
      type="button"
      onClick={onTap}
      aria-label={`Log ${title}`}
      className={[
        'w-full flex items-center gap-4 p-4 rounded-2xl bg-zinc-900 border text-left',
        'transition-all duration-200 active:scale-95',
        isDone
          ? 'border-emerald-500 shadow-[0_0_16px_rgba(16,185,129,0.3)]'
          : 'border-zinc-800',
      ].join(' ')}
    >
      {/* Icon */}
      <div
        className="flex items-center justify-center w-12 h-12 rounded-xl text-2xl flex-shrink-0"
        style={{ backgroundColor: `${color}22`, border: `1.5px solid ${color}55` }}
      >
        {icon}
      </div>

      {/* Body */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between mb-2">
          <p className="text-sm font-semibold text-white truncate">{title}</p>
          {isDone ? (
            <CheckCircle size={18} className="text-emerald-500 flex-shrink-0 ml-2" />
          ) : (
            <span className="text-xs text-zinc-400 flex-shrink-0 ml-2">
              {todayCount}/{target_count}
            </span>
          )}
        </div>

        {/* Progress bar */}
        <div className="h-1.5 bg-zinc-800 rounded-full overflow-hidden">
          <div
            className="h-full rounded-full transition-all duration-300"
            style={{
              width: `${progress * 100}%`,
              backgroundColor: isDone ? '#10b981' : color,
            }}
          />
        </div>
      </div>
    </button>
  )
}
