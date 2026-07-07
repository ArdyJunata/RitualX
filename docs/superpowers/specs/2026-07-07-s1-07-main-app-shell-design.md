# S1-07 — Main App Shell Design

> **Date:** 2026-07-07
> **Status:** Approved
> **Sprint:** S1-07

---

## Overview

Build the full main app shell for RitualX:
- Fixed top header (user avatar, XP bar, streak counter)
- `(main)/layout.tsx` wraps all protected routes
- BottomNav already exists; polish it
- Placeholder pages for each tab wired to the layout

---

## Goals

1. Top header with: avatar initials, level badge, XP progress bar, streak counter
2. `(main)/layout.tsx` = header + `{children}` + BottomNav
3. Each tab route (`/`, `/calendar`, `/stats`, `/quest`, `/profile`) renders a real placeholder page (not blank)
4. Shell uses static/hardcoded values (no API calls yet)
5. Fully responsive mobile-first layout (max-width ~430px, centered)

## Non-Goals

- Live data fetching (no API calls in the shell)
- Auth gate / redirect (added later)
- Animations beyond simple transitions

---

## Design

### Layout Structure

```
┌──────────────────────────────┐  ← sticky top, z-40
│  [Avatar]  RitualX  🔥 7     │  ← header row
│  Lv.5  ███████░░░  420 XP   │  ← XP bar row
└──────────────────────────────┘
│                              │
│         {children}           │  ← scrollable content area
│                              │
└──────────────────────────────┘
           BottomNav           ← fixed floating pill (already built)
```

### Color / Style

- Header bg: `bg-zinc-950/95 backdrop-blur-md border-b border-white/5`
- Avatar: 36×36 circle, bg `emerald-600`, initials in white
- Level badge: `text-purple` pill
- XP bar: thin (4px), `bg-zinc-800` track, `bg-emerald` fill, smooth width transition
- Streak: 🔥 icon + count, `text-amber`
- Dark theme only (zinc-950 bg, zinc-50 text) — matches existing globals.css

### Module / File Locations

Follows `frontend-architecture` skill conventions:

| What | Where |
|------|-------|
| Header component | `src/shared/components/ui/AppHeader/AppHeader.tsx` |
| Header styles | `src/shared/components/ui/AppHeader/AppHeader.styles.ts` |
| Header barrel | `src/shared/components/ui/AppHeader/index.ts` |
| Main layout | `src/app/(main)/layout.tsx` (modify) |
| BottomNav polish | `src/shared/components/ui/BottomNav/BottomNav.styles.ts` (minor tweak) |
| Home placeholder | `src/modules/home/pages/home/HomePlaceholderPage.tsx` (improve) |
| Calendar placeholder | `src/modules/calendar/pages/calendar/CalendarPlaceholderPage.tsx` |
| Stats placeholder | `src/modules/stats/pages/stats/StatsPlaceholderPage.tsx` |
| Quest placeholder | `src/modules/quest/pages/quest/QuestPlaceholderPage.tsx` |
| Profile placeholder | `src/modules/profile/pages/profile/ProfilePlaceholderPage.tsx` |
| Route pages | `src/app/(main)/calendar/page.tsx`, `stats/page.tsx`, `quest/page.tsx`, `profile/page.tsx` (wire) |

### AppHeader Props (static for now)

```tsx
interface IAppHeaderProps {
  displayName: string;   // e.g. "Alex"
  level: number;         // e.g. 5
  xp: number;            // e.g. 420
  xpToNextLevel: number; // e.g. 1000
  streakCount: number;   // e.g. 7
}
```

Header is hardcoded with `displayName="You"`, `level=1`, `xp=0`, `xpToNextLevel=100`, `streakCount=0` until API is ready.

### Placeholder Pages

Each tab gets a page that shows:
- Tab title (h1)
- Relevant icon
- "Coming soon" tagline
- Styled consistently with the dark theme

---

## Open Questions

None — all resolved.

---

## Spec Self-Review

- [x] No TBD/TODO placeholders
- [x] No internal contradictions
- [x] Requirements specific and testable (visual inspection)
- [x] Scope bounded to shell + placeholder pages only
- [x] No API calls → no error cases needed
- [x] Depends on: S1-06 (auth UI done, BottomNav exists)
