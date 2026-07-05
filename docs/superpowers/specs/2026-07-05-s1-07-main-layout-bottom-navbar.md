# S1-07: Main App Layout + Bottom Navbar — Design Spec

> **Date:** 2026-07-05
> **Status:** Approved
> **Sprint:** S1
> **Depends on:** S1-06 ?

---

## Overview

Add the main app shell layout with a floating glassmorphism bottom navigation bar. Wire 5 routes under the `(main)` route group. Create stub pages for Calendar, Stats, Quest, and Profile.

---

## Goals

1. Implement floating pill bottom navbar with glassmorphism style
2. 5 tabs: Home, Calendar, Stats, Quest, Profile
3. Active tab: emerald icon color + dot indicator below icon
4. Inactive tab: zinc-500 icon
5. Stub pages for Calendar, Stats, Quest, Profile (Coming Soon)
6. Follow frontend-architecture skill — module-based structure

---

## Non-Goals

- Actual content for Calendar, Stats, Quest, Profile pages
- Authentication guard on routes (future sprint)
- Animations beyond transition-colors

---

## Architecture

### Modules Created

| Module | Path | Purpose |
|--------|------|---------|
| home | src/modules/home/ | Home page barrel |
| calendar | src/modules/calendar/ | Calendar page stub |
| stats | src/modules/stats/ | Stats page stub |
| quest | src/modules/quest/ | Quest page stub |
| profile | src/modules/profile/ | Profile page stub |

### Component Location

BottomNav used by (main) layout — consumed by all main-route pages -> lives in shared/components/ui/BottomNav/.

### App Router Routes

src/app/(main)/
- layout.tsx       <- mounts BottomNav, pb-28 for nav clearance
- page.tsx         <- Home (daily view placeholder)
- calendar/page.tsx
- stats/page.tsx
- quest/page.tsx
- profile/page.tsx

---

## Design Details

### BottomNav Visual Spec

- Container: fixed, bottom-6, centered horizontally (left-1/2 -translate-x-1/2)
- Shape: pill (rounded-full)
- Glass: bg-zinc-900/70 backdrop-blur-xl border border-white/10
- Shadow: shadow-lg shadow-black/40
- Padding: px-6 py-3
- Gap between tabs: gap-6

### Tab Item Spec

- Icon library: Lucide React
- Icon size: w-6 h-6
- Active icon color: text-emerald-DEFAULT (#10B981)
- Inactive icon color: text-zinc-500
- Active indicator: w-1.5 h-1.5 rounded-full bg-emerald-DEFAULT below icon
- Inactive indicator: same size, bg-transparent
- Transition: transition-colors duration-200

### Tab -> Route -> Icon Map

| Tab | Route | Icon (Lucide) |
|-----|-------|---------------|
| Home | / | Home |
| Calendar | /calendar | CalendarDays |
| Stats | /stats | BarChart2 |
| Quest | /quest | Sword |
| Profile | /profile | User |

### Active Detection

Use usePathname() from next/navigation.
- Home active: pathname === "/"
- Others active: pathname.startsWith("/calendar") etc.

---

## Styling Constraints

- No inline style={{}} — all styles in *.styles.ts
- Colors from Tailwind config tokens only
- "use client" directive on BottomNav (uses usePathname)
- Stub pages are Server Components (no "use client")

---

## File List (Complete)

Create:
- src/shared/components/ui/BottomNav/BottomNav.tsx
- src/shared/components/ui/BottomNav/BottomNav.styles.ts
- src/shared/components/ui/BottomNav/index.ts
- src/modules/home/index.ts
- src/modules/home/pages/home/HomePlaceholderPage.tsx
- src/modules/home/pages/home/index.ts
- src/modules/calendar/index.ts
- src/modules/calendar/pages/calendar/CalendarPage.tsx
- src/modules/calendar/pages/calendar/index.ts
- src/modules/stats/index.ts
- src/modules/stats/pages/stats/StatsPage.tsx
- src/modules/stats/pages/stats/index.ts
- src/modules/quest/index.ts
- src/modules/quest/pages/quest/QuestPage.tsx
- src/modules/quest/pages/quest/index.ts
- src/modules/profile/index.ts
- src/modules/profile/pages/profile/ProfilePage.tsx
- src/modules/profile/pages/profile/index.ts
- src/app/(main)/calendar/page.tsx
- src/app/(main)/stats/page.tsx
- src/app/(main)/quest/page.tsx
- src/app/(main)/profile/page.tsx

Modify:
- src/app/(main)/layout.tsx — add BottomNav import + pb-28
- src/app/(main)/page.tsx — thin route, import from modules/home
