# S3-06 Contribution Heatmap Component Implementation Plan

> Checkbox (`- [ ]`) steps. Windows PowerShell. Next.js 14 App Router + Tailwind. Match existing module conventions (`@/modules/*`, `.styles.ts`, manual `useState`/`useEffect` data hooks like `useDailyView`).

**Goal:** Reusable `ContributionHeatmap` (52w×7d grid, 5-level intensity, cell tooltip) + data hook/API + integration on the stats page. Closes #19.

## Global Constraints
- `'use client'` where state/effects used · auth via `localStorage 'ritualx_access_token'` + `credentials:'include'`
- No `any`; escape apostrophes in JSX (`react/no-unescaped-entities`); correct `useEffect` deps
- No commits until instructed · Branch: `feat/s3-06-heatmap-component`

---

### Task 1: Types + API

`modules/stats/types.ts`:
```ts
export type HeatmapData = Record<string, number> // "YYYY-MM-DD" -> count
```

`modules/stats/api.ts`:
```ts
import { apiClient } from '@/shared/api-client'
import { HeatmapData } from './types'

function authHeaders() {
  const token = typeof window !== 'undefined' ? localStorage.getItem('ritualx_access_token') ?? '' : ''
  return { Authorization: `Bearer ${token}` }
}

export const statsApi = {
  getHeatmap(routineId: string, year: number): Promise<HeatmapData> {
    return apiClient.get<HeatmapData>(`/routines/${routineId}/heatmap?year=${year}`, {
      credentials: 'include',
      headers: authHeaders(),
    })
  },
}
```

---

### Task 2: Hook

`modules/stats/hooks/useRoutineHeatmap.ts` (`'use client'`) — manual fetch (mirror `useDailyView`):
returns `{ data: HeatmapData; isLoading: boolean; error: string | null; refresh: () => void }`. `useEffect` deps `[routineId, year]`. On error set a friendly message; default `data` to `{}`.

`modules/stats/hooks/index.ts` — barrel.

---

### Task 3: ContributionHeatmap (presentational)

`modules/stats/components/ContributionHeatmap.tsx` (`'use client'`), props `{ data, year?, color?, className? }`:
- `levelFor(count)`: 0/1/2/3/≥4 → 0..4.
- `buildCells(year, data)`: leading padding = `Jan1.getUTCDay()`; iterate days to Dec 31; each cell `{date, count, level, row=i%7, col=floor(i/7)}`.
- Render: horizontal-scroll wrapper → month-label row (labels at month-start columns) + `flex` [weekday labels col (Mon/Wed/Fri)] [grid `grid-flow-col grid-rows-7`, `gridAutoColumns` = cell px, `gap`].
- Cell = `<button>` with `title`/`aria-label` `"{count} on {date}"`, `onMouseEnter/onFocus`→setActive, `onMouseLeave/onBlur`→clear; bg = level 0 `#27272a` else `color` at opacity `[_,0.28,0.5,0.75,1][level]`.
- Caption above grid: active cell label, else `"{total} completions in {year}"`.
- Legend: Less → 5 swatches → More.

---

### Task 4: Container + list

`RoutineHeatmapCard.tsx` (`'use client'`, props `{ routineId, title, color?, year? }`): `useRoutineHeatmap` → loading skeleton / error+retry / `ContributionHeatmap`. Wrap in `glass-card`.

`RoutineHeatmaps.tsx` (`'use client'`): `useState`/`useEffect` → `routinesApi.list()`; render a `RoutineHeatmapCard` per routine (pass `color`); loading + empty states.

`components/index.ts` — barrel (all three).

---

### Task 5: Module + page wiring

`modules/stats/index.ts`: also `export * from './components'; export * from './api'; export * from './types'; export * from './hooks'`.

`StatsPage.tsx`: render `<RoutineHeatmaps />` where the "Charts — coming soon" card was. Add a `section`/`sectionTitle` style if needed.

---

### Task 6: Verify
- [x] `cd frontend; npx tsc --noEmit` clean
- [x] `npm run lint` clean
- [x] `npm run build` succeeds

---

### Task 7: Commit / Push / PR
Branch `feat/s3-06-heatmap-component`; commit (identity `-c`), push, PR REST API, body ends `Closes #19`.
