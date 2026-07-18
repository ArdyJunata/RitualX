# S2-06 — Daily View with Routine Cards & Log Tap

**Date:** 2026-07-18
**Sprint:** 2
**Points:** 8

---

## Overview

Replace the `HomePlaceholderPage` placeholder with a fully functional daily view. Users can see all their routines for today, tap to log completions, and tap again to undo. Optimistic UI — every tap updates instantly without waiting for the API.

---

## Goals

- Show today's routines as cards with live completion state
- Tap = log (+1), tap again when done = undo (delete most recent log)
- Multi-target routines show counted progress (e.g. 2/3)
- Glowing "done" state when `todayCount >= target_count`
- Empty state when user has no routines yet
- Top bar: date + placeholder streak/XP (real data in Sprint 3)
- Daily goal ring: "X/Y done today"

## Non-Goals

- Real streak data (Sprint 3 — S3-02)
- Real XP data (Sprint 3 — S3-05)
- Long-press for detail (Sprint 3 — S3-08)
- Ritual chains section (Sprint 4+)
- Quest banner (Sprint 5+)

---

## Approach

**Optimistic UI with local state** — fetch routines + today's logs on mount, store merged view model in local state. Each tap updates state immediately, fires API in background, rolls back on error.

No new libraries. Follows existing codebase pattern.

---

## Design Details

### 1. Data & API

**New API calls** (added to `modules/routines/api.ts`):

```ts
list(): Promise<Routine[]>
  -> GET /routines

logRoutine(id: string): Promise<RoutineLog>
  -> POST /routines/:id/log

deleteLog(routineId: string, logId: string): Promise<void>
  -> DELETE /routines/:id/log/:logId
```

**New types** (added to `modules/routines/types.ts`):

```ts
export interface RoutineLog {
  id: string
  routine_id: string
  user_id: string
  logged_at: string   // ISO timestamp
}

export interface DailyRoutine extends Routine {
  todayLogs: RoutineLog[]
  todayCount: number    // todayLogs.length
  isDone: boolean       // todayCount >= target_count
}
```

### 2. Backend Log Fetch Strategy

`GET /routines` returns routines only — no inline logs.
Strategy: after fetching routines, fetch `GET /routines/:id/log` for each routine to get today's logs, filtered client-side by `logged_at` date = today.

### 3. State & Hook

**File:** `frontend/src/modules/home/hooks/useDailyView.ts`

```ts
export function useDailyView(): {
  routines: DailyRoutine[]
  isLoading: boolean
  error: string | null
  tap: (routine: DailyRoutine) => Promise<void>
  refresh: () => Promise<void>
}
```

**Fetch on mount:**
1. GET /routines -> list
2. For each routine: GET /routines/:id/log -> logs
3. Filter logs by today's date client-side
4. Merge into DailyRoutine[] sorted by sort_order

**tap(routine) logic:**
```
if routine.todayCount < routine.target_count:
  optimistically increment todayCount
  POST /routines/:id/log -> on error: rollback + toast

else (isDone, undo):
  latestLog = routine.todayLogs[last]
  optimistically decrement todayCount
  DELETE /routines/:id/log/:latestLog.id -> on error: rollback + toast
```

### 4. UI Components

All in `frontend/src/modules/home/`.

#### DailyViewPage (replaces HomePlaceholderPage)
- Calls useDailyView()
- Renders: DailyHeader -> DailyGoalRing -> RoutineCardList or EmptyState

#### DailyHeader
- Left: today's date (e.g. "Friday, Jul 18")
- Center: streak fire 0 (placeholder)
- Right: XP bar (static/empty placeholder)

#### DailyGoalRing
- SVG circular ring
- Shows doneCount / totalCount
- Animates on count change

#### RoutineCard
- Props: routine: DailyRoutine, onTap: () => void
- Left: colored emoji icon
- Center: title + progress bar (todayCount / target_count)
- Right: count badge "2/3"
- States:
  - not-started: neutral, dim
  - in-progress: accent color on progress bar
  - done: glowing border, checkmark, full green bar
- Tap anywhere -> onTap()
- active:scale-95 press feedback

#### EmptyState
- "No routines yet" + CTA -> opens CreateRoutineSheet

---

## Error Cases

| Case | Handling |
|------|----------|
| GET /routines fails | Error state + retry button |
| Log tap fails | Rollback + toast |
| Undo fails | Rollback + toast |
| Not authenticated | Redirect to /login |

---

## File Map

| File | Action |
|------|--------|
| modules/routines/types.ts | Add RoutineLog, DailyRoutine |
| modules/routines/api.ts | Add list, logRoutine, deleteLog, getRoutineLogs |
| modules/home/hooks/useDailyView.ts | New hook |
| modules/home/hooks/index.ts | Export hook |
| modules/home/pages/home/DailyViewPage.tsx | New (replaces placeholder) |
| modules/home/pages/home/DailyHeader.tsx | New |
| modules/home/pages/home/DailyGoalRing.tsx | New |
| modules/home/pages/home/RoutineCardList.tsx | New |
| modules/home/pages/home/RoutineCard.tsx | New |
| modules/home/pages/home/EmptyState.tsx | New |
| modules/home/pages/home/DailyViewPage.styles.ts | Tailwind class map |
| modules/home/index.ts | Update export |
| app/(main)/page.tsx | Swap to DailyViewPage |

---

## Dependencies

- CHECK: GET /routines backend (S2-03) - done
- CHECK: POST /routines/:id/log (S2-04) - done
- CHECK: DELETE /routines/:id/log/:logId (S2-04) - done
- CHECK: GET /routines/:id/log endpoint - verify exists
- CHECK: CreateRoutineSheet (S2-05) - done
