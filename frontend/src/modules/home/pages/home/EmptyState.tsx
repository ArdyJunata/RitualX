// frontend/src/modules/home/pages/home/EmptyState.tsx

'use client'

interface EmptyStateProps {
  onCreateClick: () => void
}

export function EmptyState({ onCreateClick }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 gap-4 text-center px-6">
      <div className="text-6xl">🌱</div>
      <div>
        <p className="text-white font-semibold text-lg">No routines yet</p>
        <p className="text-zinc-400 text-sm mt-1">Build your first ritual to start the journey</p>
      </div>
      <button
        type="button"
        onClick={onCreateClick}
        className="mt-2 px-6 py-3 rounded-2xl bg-emerald-500 text-white font-semibold text-sm active:scale-95 transition-transform"
      >
        Create your first routine
      </button>
    </div>
  )
}
