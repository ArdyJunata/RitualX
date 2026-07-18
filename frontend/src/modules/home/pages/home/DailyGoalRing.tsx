// frontend/src/modules/home/pages/home/DailyGoalRing.tsx

'use client'

interface DailyGoalRingProps {
  done: number
  total: number
}

export function DailyGoalRing({ done, total }: DailyGoalRingProps) {
  const size = 96
  const stroke = 7
  const r = (size - stroke) / 2
  const circ = 2 * Math.PI * r
  const progress = total === 0 ? 0 : Math.min(done / total, 1)
  const offset = circ - progress * circ

  return (
    <div className="flex flex-col items-center gap-1">
      <div className="relative flex items-center justify-center" style={{ width: size, height: size }}>
        <svg width={size} height={size} className="-rotate-90">
          {/* Track */}
          <circle
            cx={size / 2} cy={size / 2} r={r}
            fill="none" stroke="#27272a" strokeWidth={stroke}
          />
          {/* Progress */}
          <circle
            cx={size / 2} cy={size / 2} r={r}
            fill="none"
            stroke={progress >= 1 ? '#10b981' : '#6366f1'}
            strokeWidth={stroke}
            strokeDasharray={circ}
            strokeDashoffset={offset}
            strokeLinecap="round"
            style={{ transition: 'stroke-dashoffset 0.4s ease' }}
          />
        </svg>
        <div className="absolute flex flex-col items-center">
          <span className="text-xl font-bold text-white">{done}</span>
          <span className="text-xs text-zinc-500">of {total}</span>
        </div>
      </div>
      <p className="text-xs text-zinc-400 font-medium">Today&apos;s Goal</p>
    </div>
  )
}
