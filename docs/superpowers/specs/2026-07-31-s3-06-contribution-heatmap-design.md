# S3-06 — GitHub-style Contribution Heatmap Component Design

> **Date:** 2026-07-31 · **Issue:** #19 · **Points:** 8 · **Frontend**

## Overview

A reusable Next.js + Tailwind contribution heatmap for a routine: a 53-column × 7-row (weeks × weekdays) grid for a calendar year, cells colored by 5-level intensity, with a tooltip per cell. Consumes the S3-04 endpoint `GET /routines/:id/heatmap?year=YYYY` (returns `{ "YYYY-MM-DD": count }`).

## Acceptance Criteria (issue #19)

- 52-week × 7-day grid ✔ (render up to 53 columns to cover a full year)
- 5-level color intensity ✔ (levels 0–4)
- Tooltip on cell ✔ (native `title` + live caption; keyboard/tap accessible)

## Non-Goals

- Full statistics screen (S3-09) · routine detail screen (S3-08)
- Weekly/monthly chart views (S3-07)

## Components (module `@/modules/stats`)

- **`ContributionHeatmap`** (presentational, `'use client'`) — props `{ data: Record<string,number>; year?: number; color?: string; className?: string }`. Pure: builds the grid from the date→count map. No fetching → trivially reusable/testable.
- **`RoutineHeatmapCard`** (container) — props `{ routineId; title; color?; year? }`; uses `useRoutineHeatmap`, renders loading / error / empty / `ContributionHeatmap`.
- **`RoutineHeatmaps`** (client) — fetches the user's routines (`routinesApi.list()`) and renders a `RoutineHeatmapCard` per routine. Mounted by `StatsPage` (verification surface; S3-09 will build the full screen).

## Data

- `modules/stats/types.ts`: `export type HeatmapData = Record<string, number>`.
- `modules/stats/api.ts`: `statsApi.getHeatmap(routineId, year): Promise<HeatmapData>` → `GET /routines/:id/heatmap?year=` with auth headers + `credentials:'include'` (mirrors `routinesApi`).
- `modules/stats/hooks/useRoutineHeatmap.ts`: manual `useState`/`useEffect` (matches `useDailyView`) → `{ data, isLoading, error, refresh }`.

## Grid Algorithm

- Days Jan 1 → Dec 31 of `year`. Leading padding cells = weekday of Jan 1 (Sun=0). Cell index `i`: `col = floor(i/7)`, `row = i % 7`. CSS grid `grid-flow-col grid-rows-7` lays cells column-by-column (top→bottom then next week).
- Month labels positioned by the column of each month's 1st. Optional weekday labels (Mon/Wed/Fri) in a left column.
- Horizontal scroll wrapper (year width ≈ 53×15px > mobile viewport).

## Intensity (5 levels)

Habit counts are small, so fixed thresholds: `0→0`, `1→1`, `2→2`, `3→3`, `≥4→4`. Level 0 = track color `#27272a` (zinc-800). Levels 1–4 = base `color` (routine color, default emerald `#10b981`) at opacity `[0.28, 0.5, 0.75, 1]`. A "Less→More" legend renders the 5 swatches.

## Tooltip

Each cell is a `<button>` with `title` + `aria-label` = `"{count} on {YYYY-MM-DD}"` (native tooltip, works on desktop hover; a11y). A live caption above the grid updates on `mouseenter`/`focus`/tap to the hovered cell's label (mobile-friendly; no clipping vs. an absolutely-positioned tooltip inside the scroll container).

## Verification Surface

`StatsPage` renders `<RoutineHeatmaps />` (replaces the "Charts — coming soon" card). Keeps existing header + placeholder stat cards (real overview wiring is S3-09).

## File Map

| File | Action |
|------|--------|
| `modules/stats/types.ts` | New — `HeatmapData` |
| `modules/stats/api.ts` | New — `statsApi.getHeatmap` |
| `modules/stats/hooks/useRoutineHeatmap.ts` + `hooks/index.ts` | New — data hook |
| `modules/stats/components/ContributionHeatmap.tsx` | New — grid component |
| `modules/stats/components/RoutineHeatmapCard.tsx` | New — container |
| `modules/stats/components/RoutineHeatmaps.tsx` | New — routines list → cards |
| `modules/stats/components/index.ts` | New — barrel |
| `modules/stats/index.ts` | Update — export components/api/hooks/types |
| `modules/stats/pages/stats/StatsPage.tsx` + `.styles.ts` | Update — mount `RoutineHeatmaps` |

## Dependencies

- S3-04 `/routines/:id/heatmap` ✅ · `routinesApi.list()` (S2) ✅ · React 18, Tailwind, lucide-react ✅

## Open Questions

- Fixed intensity thresholds vs. max-relative quartiles — chose fixed (habit counts are small); reviewer may tune.
