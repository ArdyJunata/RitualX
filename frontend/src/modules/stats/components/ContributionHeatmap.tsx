// frontend/src/modules/stats/components/ContributionHeatmap.tsx

'use client'

import { useMemo, useState } from 'react'
import type { CSSProperties } from 'react'

export interface ContributionHeatmapProps {
  data: Record<string, number>
  year?: number
  color?: string
  className?: string
}

interface HeatmapCell {
  date: string | null
  count: number
  level: number
  row: number
  col: number
}

const DEFAULT_COLOR = '#10b981'
const TRACK_COLOR = '#27272a'
const LEVEL_OPACITY = [0, 0.28, 0.5, 0.75, 1]
const CELL = 12
const GAP = 3
const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
const WEEKDAYS = ['', 'Mon', '', 'Wed', '', 'Fri', '']

function levelFor(count: number): number {
  if (count <= 0) return 0
  if (count >= 4) return 4
  return count
}

function pad2(n: number): string {
  return n < 10 ? `0${n}` : `${n}`
}

function dateKey(d: Date): string {
  return `${d.getUTCFullYear()}-${pad2(d.getUTCMonth() + 1)}-${pad2(d.getUTCDate())}`
}

interface BuiltGrid {
  cells: HeatmapCell[]
  weeks: number
  total: number
  monthLabels: { label: string; col: number }[]
}

function buildGrid(year: number, data: Record<string, number>): BuiltGrid {
  const first = new Date(Date.UTC(year, 0, 1))
  const last = new Date(Date.UTC(year, 11, 31))
  const startWeekday = first.getUTCDay() // 0 = Sunday

  const cells: HeatmapCell[] = []
  const monthLabels: { label: string; col: number }[] = []
  let index = 0
  let total = 0
  let lastMonth = -1

  // Leading padding so Jan 1 lands on its weekday row.
  for (let i = 0; i < startWeekday; i++) {
    cells.push({ date: null, count: 0, level: 0, row: index % 7, col: Math.floor(index / 7) })
    index++
  }

  const cursor = new Date(first)
  while (cursor.getTime() <= last.getTime()) {
    const col = Math.floor(index / 7)
    const month = cursor.getUTCMonth()
    if (month !== lastMonth) {
      monthLabels.push({ label: MONTHS[month], col })
      lastMonth = month
    }
    const key = dateKey(cursor)
    const count = data[key] ?? 0
    total += count
    cells.push({ date: key, count, level: levelFor(count), row: index % 7, col })
    index++
    cursor.setUTCDate(cursor.getUTCDate() + 1)
  }

  return { cells, weeks: Math.ceil(index / 7), total, monthLabels }
}

export function ContributionHeatmap({ data, year, color, className }: ContributionHeatmapProps) {
  const yr = year ?? new Date().getUTCFullYear()
  const base = color || DEFAULT_COLOR
  const { cells, weeks, total, monthLabels } = useMemo(() => buildGrid(yr, data), [yr, data])
  const [active, setActive] = useState<HeatmapCell | null>(null)

  const gridWidth = weeks * (CELL + GAP)
  const rowsTemplate = `repeat(7, ${CELL}px)`

  function cellStyle(level: number): CSSProperties {
    if (level === 0) return { width: CELL, height: CELL, backgroundColor: TRACK_COLOR }
    return { width: CELL, height: CELL, backgroundColor: base, opacity: LEVEL_OPACITY[level] }
  }

  const captionText =
    active && active.date
      ? `${active.count} ${active.count === 1 ? 'completion' : 'completions'} on ${active.date}`
      : `${total} ${total === 1 ? 'completion' : 'completions'} in ${yr}`

  return (
    <div className={className}>
      <div className="mb-2 h-4 text-xs text-zinc-400">{captionText}</div>

      <div className="overflow-x-auto pb-1">
        <div style={{ width: gridWidth + 28 }}>
          {/* month labels */}
          <div className="relative ml-7 h-4" style={{ width: gridWidth }}>
            {monthLabels.map((m) => (
              <span
                key={`${m.label}-${m.col}`}
                className="absolute text-[10px] text-zinc-500"
                style={{ left: m.col * (CELL + GAP) }}
              >
                {m.label}
              </span>
            ))}
          </div>

          <div className="flex gap-1">
            {/* weekday labels */}
            <div className="grid" style={{ gridTemplateRows: rowsTemplate, gap: GAP, width: 24 }}>
              {WEEKDAYS.map((w, i) => (
                <span key={i} className="flex items-center text-[9px] leading-none text-zinc-600">
                  {w}
                </span>
              ))}
            </div>

            {/* day cells */}
            <div
              className="grid grid-flow-col"
              style={{ gridTemplateRows: rowsTemplate, gridAutoColumns: `${CELL}px`, gap: GAP }}
            >
              {cells.map((cell, i) =>
                cell.date === null ? (
                  <div key={`pad-${i}`} style={{ width: CELL, height: CELL }} />
                ) : (
                  <button
                    key={cell.date}
                    type="button"
                    title={`${cell.count} on ${cell.date}`}
                    aria-label={`${cell.count} completions on ${cell.date}`}
                    onMouseEnter={() => setActive(cell)}
                    onMouseLeave={() => setActive(null)}
                    onFocus={() => setActive(cell)}
                    onBlur={() => setActive(null)}
                    className="rounded-sm border border-white/5 focus:outline focus:outline-1 focus:outline-emerald-400"
                    style={cellStyle(cell.level)}
                  />
                )
              )}
            </div>
          </div>
        </div>
      </div>

      {/* legend */}
      <div className="mt-2 flex items-center gap-1 text-[10px] text-zinc-500">
        <span>Less</span>
        {[0, 1, 2, 3, 4].map((l) => (
          <span key={l} className="rounded-sm border border-white/5" style={cellStyle(l)} />
        ))}
        <span>More</span>
      </div>
    </div>
  )
}
