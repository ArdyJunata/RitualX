// frontend/src/modules/home/pages/home/DailyHeader.tsx

'use client'

function formatDate(date: Date): string {
  return date.toLocaleDateString('en-US', { weekday: 'long', month: 'short', day: 'numeric' })
}

export function DailyHeader() {
  const today = new Date()

  return (
    <div className="flex items-center justify-between px-1 py-2">
      <div>
        <p className="text-xs text-zinc-500 font-medium uppercase tracking-wider">Today</p>
        <h1 className="text-base font-bold text-white">{formatDate(today)}</h1>
      </div>
      <div className="flex items-center gap-3">
        {/* Streak — placeholder until Sprint 3 */}
        <div className="flex items-center gap-1 bg-zinc-800 rounded-full px-3 py-1.5">
          <span className="text-base">🔥</span>
          <span className="text-sm font-bold text-white">0</span>
        </div>
      </div>
    </div>
  )
}
