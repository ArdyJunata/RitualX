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

      <div className="grid grid-cols-6 gap-4 px-2">
        {COLORS.map(hex => (
          <button
            key={hex}
            type="button"
            onClick={() => onSelect(hex)}
            style={{ backgroundColor: hex }}
            className={[
              'w-full aspect-square rounded-full transition-all duration-150',
              color === hex
                ? 'ring-2 ring-white ring-offset-2 ring-offset-zinc-900 scale-110'
                : 'scale-100 hover:scale-105',
            ].join(' ')}
            aria-label={`Color ${hex}`}
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
          className="flex-1 py-3 rounded-xl bg-emerald-500 text-white font-semibold text-sm disabled:opacity-40 disabled:cursor-not-allowed transition-opacity duration-150"
        >
          Next →
        </button>
      </div>
    </div>
  )
}
