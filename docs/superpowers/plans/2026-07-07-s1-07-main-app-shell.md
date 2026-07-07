# S1-07 — Main App Shell Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the main app shell — sticky top header with avatar/XP bar/streak, wire `(main)/layout.tsx`, and add styled placeholder pages for all 5 tabs.

**Architecture:** `AppHeader` lives in `shared/components/ui/AppHeader/`; it is a Server Component (no interactivity needed — static props). `(main)/layout.tsx` imports AppHeader + BottomNav and wraps all protected routes. Each module gets a real placeholder page component in its `pages/` directory, exported via barrel, and wired via the thin App Router `page.tsx`.

**Tech Stack:** Next.js 14 App Router, Tailwind CSS v3, lucide-react, TypeScript 5.

## Global Constraints

- Mobile-first, max content width 430px (`max-w-[430px] mx-auto`)
- Dark theme only: bg `#09090b` (zinc-950), text `#fafafa` (zinc-50)
- Tailwind color tokens only — no inline hex values
- No `fetch()` / API calls in this task
- No commits until explicitly instructed
- All new components: co-located `.styles.ts` file, no inline styles in JSX
- Interface names prefixed with `I` (e.g. `IAppHeaderProps`)
- Page directories: `{name}.tsx` + `{name}.styles.ts` + `index.ts`

---

## File Map

| Action | File |
|--------|------|
| Create | `src/shared/components/ui/AppHeader/AppHeader.tsx` |
| Create | `src/shared/components/ui/AppHeader/AppHeader.styles.ts` |
| Create | `src/shared/components/ui/AppHeader/index.ts` |
| Modify | `src/app/(main)/layout.tsx` |
| Modify | `src/shared/components/ui/BottomNav/BottomNav.styles.ts` |
| Modify | `src/modules/home/pages/home/HomePlaceholderPage.tsx` |
| Create | `src/modules/home/pages/home/HomePlaceholderPage.styles.ts` |
| Create | `src/modules/calendar/pages/calendar/CalendarPlaceholderPage.tsx` |
| Create | `src/modules/calendar/pages/calendar/CalendarPlaceholderPage.styles.ts` |
| Create | `src/modules/calendar/pages/calendar/index.ts` |
| Modify | `src/modules/calendar/index.ts` |
| Modify | `src/app/(main)/calendar/page.tsx` |
| Create | `src/modules/stats/pages/stats/StatsPlaceholderPage.tsx` |
| Create | `src/modules/stats/pages/stats/StatsPlaceholderPage.styles.ts` |
| Create | `src/modules/stats/pages/stats/index.ts` |
| Modify | `src/modules/stats/index.ts` |
| Create | `src/app/(main)/stats/page.tsx` |
| Create | `src/modules/quest/pages/quest/QuestPlaceholderPage.tsx` |
| Create | `src/modules/quest/pages/quest/QuestPlaceholderPage.styles.ts` |
| Create | `src/modules/quest/pages/quest/index.ts` |
| Modify | `src/modules/quest/index.ts` |
| Create | `src/app/(main)/quest/page.tsx` |
| Create | `src/modules/profile/pages/profile/ProfilePlaceholderPage.tsx` |
| Create | `src/modules/profile/pages/profile/ProfilePlaceholderPage.styles.ts` |
| Create | `src/modules/profile/pages/profile/index.ts` |
| Modify | `src/modules/profile/index.ts` |
| Create | `src/app/(main)/profile/page.tsx` |

---

## Task 1: AppHeader Component

**Files:**
- Create: `src/shared/components/ui/AppHeader/AppHeader.styles.ts`
- Create: `src/shared/components/ui/AppHeader/AppHeader.tsx`
- Create: `src/shared/components/ui/AppHeader/index.ts`

**Interfaces:**
- Produces: `AppHeader` component, `IAppHeaderProps` interface

- [ ] **Step 1: Create the styles file**

  Create `src/shared/components/ui/AppHeader/AppHeader.styles.ts`:

  ```ts
  export const appHeaderStyles = {
    root: "sticky top-0 z-40 w-full bg-zinc-950/95 backdrop-blur-md border-b border-white/5",
    inner: "max-w-[430px] mx-auto px-4 py-2 flex flex-col gap-1.5",
    topRow: "flex items-center justify-between",
    // Left side: avatar + app name
    leftGroup: "flex items-center gap-2.5",
    avatar:
      "w-9 h-9 rounded-full bg-emerald-600 flex items-center justify-center text-white text-sm font-bold shrink-0",
    appName: "font-heading text-base font-semibold text-zinc-100 tracking-wide",
    // Right side: streak
    streakGroup: "flex items-center gap-1.5",
    streakIcon: "text-base leading-none select-none",
    streakCount: "text-sm font-semibold text-amber",
    // XP row
    xpRow: "flex items-center gap-2",
    levelBadge:
      "text-xs font-semibold text-purple px-1.5 py-0.5 rounded-full border border-purple/30 shrink-0",
    xpTrack: "flex-1 h-1 rounded-full bg-zinc-800 overflow-hidden",
    xpFill: "h-full rounded-full bg-emerald transition-[width] duration-500 ease-out",
    xpText: "text-xs text-zinc-500 shrink-0 tabular-nums",
  } as const;
  ```

- [ ] **Step 2: Create the component**

  Create `src/shared/components/ui/AppHeader/AppHeader.tsx`:

  ```tsx
  import { appHeaderStyles as s } from "./AppHeader.styles";

  export interface IAppHeaderProps {
    displayName: string;
    level: number;
    xp: number;
    xpToNextLevel: number;
    streakCount: number;
  }

  export function AppHeader({
    displayName,
    level,
    xp,
    xpToNextLevel,
    streakCount,
  }: IAppHeaderProps) {
    const initials = displayName.slice(0, 2).toUpperCase();
    const xpPercent = Math.min(100, Math.round((xp / xpToNextLevel) * 100));

    return (
      <header className={s.root}>
        <div className={s.inner}>
          {/* Top row: avatar + name | streak */}
          <div className={s.topRow}>
            <div className={s.leftGroup}>
              <div className={s.avatar} aria-hidden="true">
                {initials}
              </div>
              <span className={s.appName}>RitualX</span>
            </div>
            <div className={s.streakGroup} aria-label={`${streakCount} day streak`}>
              <span className={s.streakIcon}>🔥</span>
              <span className={s.streakCount}>{streakCount}</span>
            </div>
          </div>

          {/* XP row: level | bar | xp text */}
          <div className={s.xpRow}>
            <span className={s.levelBadge} aria-label={`Level ${level}`}>
              Lv.{level}
            </span>
            <div
              className={s.xpTrack}
              role="progressbar"
              aria-valuenow={xp}
              aria-valuemin={0}
              aria-valuemax={xpToNextLevel}
              aria-label="XP progress"
            >
              <div
                className={s.xpFill}
                style={{ width: `${xpPercent}%` }}
              />
            </div>
            <span className={s.xpText}>{xp}/{xpToNextLevel} XP</span>
          </div>
        </div>
      </header>
    );
  }
  ```

  > **Note on `style={{ width }}`:** The XP bar fill width is dynamic (a computed percentage). This is the one permitted inline style — a co-located styles file cannot express a runtime value. All other styling is in `AppHeader.styles.ts`.

- [ ] **Step 3: Create the barrel**

  Create `src/shared/components/ui/AppHeader/index.ts`:

  ```ts
  export { AppHeader } from "./AppHeader";
  export type { IAppHeaderProps } from "./AppHeader";
  ```

- [ ] **Step 4: Verify no TypeScript errors**

  Run from `frontend/`:
  ```bash
  npx tsc --noEmit
  ```
  Expected: no errors related to AppHeader.

---

## Task 2: Wire (main)/layout.tsx

**Files:**
- Modify: `src/app/(main)/layout.tsx`

**Interfaces:**
- Consumes: `AppHeader` from `@/shared/components/ui/AppHeader`
- Consumes: `BottomNav` from `@/shared/components/ui/BottomNav`

- [ ] **Step 1: Replace layout.tsx content**

  Current content of `src/app/(main)/layout.tsx`:
  ```tsx
  import type { ReactNode } from "react";
  import { BottomNav } from "@/shared/components/ui/BottomNav";

  export default function MainLayout({ children }: { children: ReactNode }) {
    return (
      <div className="min-h-screen pb-28">
        {children}
        <BottomNav />
      </div>
    );
  }
  ```

  Replace with:
  ```tsx
  import type { ReactNode } from "react";
  import { BottomNav } from "@/shared/components/ui/BottomNav";
  import { AppHeader } from "@/shared/components/ui/AppHeader";

  export default function MainLayout({ children }: { children: ReactNode }) {
    return (
      <div className="min-h-screen bg-background">
        <AppHeader
          displayName="You"
          level={1}
          xp={0}
          xpToNextLevel={100}
          streakCount={0}
        />
        <main className="max-w-[430px] mx-auto px-4 pt-4 pb-32">
          {children}
        </main>
        <BottomNav />
      </div>
    );
  }
  ```

- [ ] **Step 2: Verify build**

  Run from `frontend/`:
  ```bash
  npx tsc --noEmit
  ```
  Expected: no errors.

---

## Task 3: Improved Home Placeholder Page

**Files:**
- Modify: `src/modules/home/pages/home/HomePlaceholderPage.tsx`
- Create: `src/modules/home/pages/home/HomePlaceholderPage.styles.ts`

**Interfaces:**
- Produces: `HomePlaceholderPage` component (no props)

- [ ] **Step 1: Create styles file**

  Create `src/modules/home/pages/home/HomePlaceholderPage.styles.ts`:

  ```ts
  export const homePlaceholderStyles = {
    root: "flex flex-col gap-6 py-4",
    greeting: "flex flex-col gap-1",
    title: "font-heading text-2xl font-bold text-zinc-100",
    subtitle: "text-sm text-zinc-500",
    card: "glass-card p-5 flex flex-col gap-3",
    cardTitle: "text-base font-semibold text-zinc-200",
    cardBody: "text-sm text-zinc-500",
    pillRow: "flex gap-2 flex-wrap",
    pill: "text-xs px-2.5 py-1 rounded-full bg-zinc-800 text-zinc-400 border border-white/5",
  } as const;
  ```

- [ ] **Step 2: Replace HomePlaceholderPage.tsx**

  Replace `src/modules/home/pages/home/HomePlaceholderPage.tsx`:

  ```tsx
  import { homePlaceholderStyles as s } from "./HomePlaceholderPage.styles";

  export function HomePlaceholderPage() {
    return (
      <div className={s.root}>
        <div className={s.greeting}>
          <h1 className={s.title}>Good morning 👋</h1>
          <p className={s.subtitle}>Your daily rituals await</p>
        </div>

        <div className={s.card}>
          <p className={s.cardTitle}>Today's Goal</p>
          <p className={s.cardBody}>Track your routines to build your streak</p>
          <div className={s.pillRow}>
            <span className={s.pill}>🏃 Morning Run</span>
            <span className={s.pill}>📚 Reading</span>
            <span className={s.pill}>💧 Hydration</span>
          </div>
        </div>

        <div className={s.card}>
          <p className={s.cardTitle}>Heatmap</p>
          <p className={s.cardBody}>GitHub-style contribution calendar — coming soon</p>
        </div>
      </div>
    );
  }
  ```

---

## Task 4: Calendar Placeholder Page

**Files:**
- Create: `src/modules/calendar/pages/calendar/CalendarPlaceholderPage.tsx`
- Create: `src/modules/calendar/pages/calendar/CalendarPlaceholderPage.styles.ts`
- Create: `src/modules/calendar/pages/calendar/index.ts`
- Modify: `src/modules/calendar/index.ts`
- Modify: `src/app/(main)/calendar/page.tsx`

**Interfaces:**
- Produces: `CalendarPlaceholderPage` component (no props)

- [ ] **Step 1: Check existing calendar module barrel**

  Read `src/modules/calendar/index.ts`. It currently has a placeholder export. Note the existing export name.

- [ ] **Step 2: Create styles file**

  Create `src/modules/calendar/pages/calendar/CalendarPlaceholderPage.styles.ts`:

  ```ts
  export const calendarPlaceholderStyles = {
    root: "flex flex-col gap-6 py-4",
    header: "flex flex-col gap-1",
    title: "font-heading text-2xl font-bold text-zinc-100",
    subtitle: "text-sm text-zinc-500",
    card: "glass-card p-5 flex flex-col gap-3",
    cardTitle: "text-base font-semibold text-zinc-200",
    cardBody: "text-sm text-zinc-500",
    grid: "grid grid-cols-7 gap-1",
    cell: "aspect-square rounded-sm bg-zinc-800/60",
    cellActive: "aspect-square rounded-sm bg-emerald/60",
  } as const;
  ```

- [ ] **Step 3: Create page component**

  Create `src/modules/calendar/pages/calendar/CalendarPlaceholderPage.tsx`:

  ```tsx
  import { calendarPlaceholderStyles as s } from "./CalendarPlaceholderPage.styles";

  // 35 cells = 5 weeks × 7 days — mock heatmap preview
  const MOCK_CELLS = Array.from({ length: 35 }, (_, i) => ({
    id: i,
    active: Math.random() > 0.6,
  }));

  export function CalendarPlaceholderPage() {
    return (
      <div className={s.root}>
        <div className={s.header}>
          <h1 className={s.title}>Calendar</h1>
          <p className={s.subtitle}>Contribution heatmap — coming soon</p>
        </div>

        <div className={s.card}>
          <p className={s.cardTitle}>Activity Heatmap Preview</p>
          <div className={s.grid}>
            {MOCK_CELLS.map((cell) => (
              <div
                key={cell.id}
                className={cell.active ? s.cellActive : s.cell}
                aria-hidden="true"
              />
            ))}
          </div>
        </div>
      </div>
    );
  }
  ```

- [ ] **Step 4: Create page barrel**

  Create `src/modules/calendar/pages/calendar/index.ts`:

  ```ts
  export { CalendarPlaceholderPage } from "./CalendarPlaceholderPage";
  ```

- [ ] **Step 5: Update module barrel**

  Read the current `src/modules/calendar/index.ts` content first, then replace it entirely with:

  ```ts
  export { CalendarPlaceholderPage } from "./pages/calendar";
  ```

- [ ] **Step 6: Wire route**

  Read the current `src/app/(main)/calendar/page.tsx` content, then replace entirely with:

  ```tsx
  import { CalendarPlaceholderPage } from "@/modules/calendar";
  export default function Page() { return <CalendarPlaceholderPage />; }
  ```

---

## Task 5: Stats Placeholder Page

**Files:**
- Create: `src/modules/stats/pages/stats/StatsPlaceholderPage.tsx`
- Create: `src/modules/stats/pages/stats/StatsPlaceholderPage.styles.ts`
- Create: `src/modules/stats/pages/stats/index.ts`
- Modify: `src/modules/stats/index.ts`
- Create: `src/app/(main)/stats/page.tsx`

**Interfaces:**
- Produces: `StatsPlaceholderPage` component (no props)

- [ ] **Step 1: Create styles file**

  Create `src/modules/stats/pages/stats/StatsPlaceholderPage.styles.ts`:

  ```ts
  export const statsPlaceholderStyles = {
    root: "flex flex-col gap-6 py-4",
    header: "flex flex-col gap-1",
    title: "font-heading text-2xl font-bold text-zinc-100",
    subtitle: "text-sm text-zinc-500",
    grid: "grid grid-cols-2 gap-3",
    statCard: "glass-card p-4 flex flex-col gap-1",
    statValue: "font-heading text-3xl font-bold text-emerald",
    statLabel: "text-xs text-zinc-500",
    card: "glass-card p-5 flex flex-col gap-2",
    cardTitle: "text-base font-semibold text-zinc-200",
    cardBody: "text-sm text-zinc-500",
  } as const;
  ```

- [ ] **Step 2: Create page component**

  Create `src/modules/stats/pages/stats/StatsPlaceholderPage.tsx`:

  ```tsx
  import { statsPlaceholderStyles as s } from "./StatsPlaceholderPage.styles";

  const MOCK_STATS = [
    { value: "0",  label: "Total Logs" },
    { value: "0",  label: "Best Streak" },
    { value: "0",  label: "XP Earned" },
    { value: "0",  label: "Quests Done" },
  ] as const;

  export function StatsPlaceholderPage() {
    return (
      <div className={s.root}>
        <div className={s.header}>
          <h1 className={s.title}>Statistics</h1>
          <p className={s.subtitle}>Your progress at a glance</p>
        </div>

        <div className={s.grid}>
          {MOCK_STATS.map((stat) => (
            <div key={stat.label} className={s.statCard}>
              <span className={s.statValue}>{stat.value}</span>
              <span className={s.statLabel}>{stat.label}</span>
            </div>
          ))}
        </div>

        <div className={s.card}>
          <p className={s.cardTitle}>Charts</p>
          <p className={s.cardBody}>Weekly and monthly breakdown — coming soon</p>
        </div>
      </div>
    );
  }
  ```

- [ ] **Step 3: Create page barrel**

  Create `src/modules/stats/pages/stats/index.ts`:

  ```ts
  export { StatsPlaceholderPage } from "./StatsPlaceholderPage";
  ```

- [ ] **Step 4: Update module barrel**

  Replace `src/modules/stats/index.ts` entirely with:

  ```ts
  export { StatsPlaceholderPage } from "./pages/stats";
  ```

- [ ] **Step 5: Check if stats route page.tsx already exists**

  Check `src/app/(main)/stats/` — if `page.tsx` exists, read and replace; if not, create.

  Content of `src/app/(main)/stats/page.tsx`:

  ```tsx
  import { StatsPlaceholderPage } from "@/modules/stats";
  export default function Page() { return <StatsPlaceholderPage />; }
  ```

---

## Task 6: Quest Placeholder Page

**Files:**
- Create: `src/modules/quest/pages/quest/QuestPlaceholderPage.tsx`
- Create: `src/modules/quest/pages/quest/QuestPlaceholderPage.styles.ts`
- Create: `src/modules/quest/pages/quest/index.ts`
- Modify: `src/modules/quest/index.ts`
- Create: `src/app/(main)/quest/page.tsx`

**Interfaces:**
- Produces: `QuestPlaceholderPage` component (no props)

- [ ] **Step 1: Create styles file**

  Create `src/modules/quest/pages/quest/QuestPlaceholderPage.styles.ts`:

  ```ts
  export const questPlaceholderStyles = {
    root: "flex flex-col gap-6 py-4",
    header: "flex flex-col gap-1",
    title: "font-heading text-2xl font-bold text-zinc-100",
    subtitle: "text-sm text-zinc-500",
    card: "glass-card p-5 flex flex-col gap-3",
    cardTitle: "text-base font-semibold text-zinc-200",
    cardBody: "text-sm text-zinc-500",
    questItem: "flex items-center gap-3 py-2 border-b border-white/5 last:border-0",
    questIcon: "text-xl shrink-0",
    questInfo: "flex flex-col gap-0.5 min-w-0",
    questName: "text-sm font-medium text-zinc-200 truncate",
    questReward: "text-xs text-purple",
    questProgress: "ml-auto shrink-0 text-xs text-zinc-500 tabular-nums",
  } as const;
  ```

- [ ] **Step 2: Create page component**

  Create `src/modules/quest/pages/quest/QuestPlaceholderPage.tsx`:

  ```tsx
  import { questPlaceholderStyles as s } from "./QuestPlaceholderPage.styles";

  const MOCK_QUESTS = [
    { icon: "⚡", name: "Log 3 routines today",  reward: "+50 XP", progress: "0/3" },
    { icon: "🔥", name: "Reach a 7-day streak",   reward: "+100 XP", progress: "0/7" },
    { icon: "🏆", name: "Complete 10 total logs",  reward: "+200 XP", progress: "0/10" },
  ] as const;

  export function QuestPlaceholderPage() {
    return (
      <div className={s.root}>
        <div className={s.header}>
          <h1 className={s.title}>Quests</h1>
          <p className={s.subtitle}>Complete challenges to earn XP and rewards</p>
        </div>

        <div className={s.card}>
          <p className={s.cardTitle}>Active Quests</p>
          {MOCK_QUESTS.map((quest) => (
            <div key={quest.name} className={s.questItem}>
              <span className={s.questIcon}>{quest.icon}</span>
              <div className={s.questInfo}>
                <span className={s.questName}>{quest.name}</span>
                <span className={s.questReward}>{quest.reward}</span>
              </div>
              <span className={s.questProgress}>{quest.progress}</span>
            </div>
          ))}
        </div>
      </div>
    );
  }
  ```

- [ ] **Step 3: Create page barrel**

  Create `src/modules/quest/pages/quest/index.ts`:

  ```ts
  export { QuestPlaceholderPage } from "./QuestPlaceholderPage";
  ```

- [ ] **Step 4: Update module barrel**

  Replace `src/modules/quest/index.ts` entirely with:

  ```ts
  export { QuestPlaceholderPage } from "./pages/quest";
  ```

- [ ] **Step 5: Wire route**

  Create `src/app/(main)/quest/page.tsx`:

  ```tsx
  import { QuestPlaceholderPage } from "@/modules/quest";
  export default function Page() { return <QuestPlaceholderPage />; }
  ```

---

## Task 7: Profile Placeholder Page

**Files:**
- Create: `src/modules/profile/pages/profile/ProfilePlaceholderPage.tsx`
- Create: `src/modules/profile/pages/profile/ProfilePlaceholderPage.styles.ts`
- Create: `src/modules/profile/pages/profile/index.ts`
- Modify: `src/modules/profile/index.ts`
- Create: `src/app/(main)/profile/page.tsx`

**Interfaces:**
- Produces: `ProfilePlaceholderPage` component (no props)

- [ ] **Step 1: Create styles file**

  Create `src/modules/profile/pages/profile/ProfilePlaceholderPage.styles.ts`:

  ```ts
  export const profilePlaceholderStyles = {
    root: "flex flex-col gap-6 py-4",
    avatarSection: "flex flex-col items-center gap-3 pt-4",
    avatar:
      "w-20 h-20 rounded-full bg-emerald-600 flex items-center justify-center text-white text-3xl font-bold",
    displayName: "font-heading text-xl font-bold text-zinc-100",
    title: "text-sm text-zinc-500",
    statsRow: "flex gap-4 justify-center",
    statItem: "flex flex-col items-center gap-0.5",
    statValue: "font-heading text-lg font-bold text-zinc-100",
    statLabel: "text-xs text-zinc-500",
    card: "glass-card p-5 flex flex-col gap-3",
    cardTitle: "text-base font-semibold text-zinc-200",
    menuItem:
      "flex items-center justify-between py-2 border-b border-white/5 last:border-0 text-sm text-zinc-300",
    menuArrow: "text-zinc-600",
  } as const;
  ```

- [ ] **Step 2: Create page component**

  Create `src/modules/profile/pages/profile/ProfilePlaceholderPage.tsx`:

  ```tsx
  import { profilePlaceholderStyles as s } from "./ProfilePlaceholderPage.styles";

  const MENU_ITEMS = [
    "Edit Profile",
    "Notification Settings",
    "Privacy",
    "Linked Partners",
    "Sign Out",
  ] as const;

  export function ProfilePlaceholderPage() {
    return (
      <div className={s.root}>
        <div className={s.avatarSection}>
          <div className={s.avatar}>YO</div>
          <p className={s.displayName}>You</p>
          <p className={s.title}>Novice · Level 1</p>
          <div className={s.statsRow}>
            <div className={s.statItem}>
              <span className={s.statValue}>0</span>
              <span className={s.statLabel}>Routines</span>
            </div>
            <div className={s.statItem}>
              <span className={s.statValue}>0</span>
              <span className={s.statLabel}>Streak</span>
            </div>
            <div className={s.statItem}>
              <span className={s.statValue}>0</span>
              <span className={s.statLabel}>XP</span>
            </div>
          </div>
        </div>

        <div className={s.card}>
          <p className={s.cardTitle}>Settings</p>
          {MENU_ITEMS.map((item) => (
            <div key={item} className={s.menuItem}>
              <span>{item}</span>
              <span className={s.menuArrow}>›</span>
            </div>
          ))}
        </div>
      </div>
    );
  }
  ```

- [ ] **Step 3: Create page barrel**

  Create `src/modules/profile/pages/profile/index.ts`:

  ```ts
  export { ProfilePlaceholderPage } from "./ProfilePlaceholderPage";
  ```

- [ ] **Step 4: Update module barrel**

  Replace `src/modules/profile/index.ts` entirely with:

  ```ts
  export { ProfilePlaceholderPage } from "./pages/profile";
  ```

- [ ] **Step 5: Wire route**

  Create `src/app/(main)/profile/page.tsx`:

  ```tsx
  import { ProfilePlaceholderPage } from "@/modules/profile";
  export default function Page() { return <ProfilePlaceholderPage />; }
  ```

---

## Task 8: Verify Full Build

**Files:** none (verification only)

- [ ] **Step 1: TypeScript check**

  Run from `frontend/`:
  ```bash
  npx tsc --noEmit
  ```
  Expected: exit code 0, no errors.

- [ ] **Step 2: Start dev server and manual smoke test**

  Run from `frontend/`:
  ```bash
  npm run dev
  ```
  Open `http://localhost:3000` in browser.

  Verify each tab:
  - `/` → Home page with greeting card and heatmap placeholder visible
  - `/calendar` → Calendar page with mock heatmap grid visible
  - `/stats` → Stats page with 4 stat cards visible
  - `/quest` → Quest page with 3 quest items visible
  - `/profile` → Profile page with avatar, stats row, settings card visible
  - All pages: AppHeader visible at top (avatar, XP bar, streak)
  - All pages: BottomNav floating pill visible at bottom
  - Active tab in BottomNav highlighted with emerald dot

- [ ] **Step 3: Update checkpoint**

  Update `.gemini/CHECKPOINT.md`:
  - Mark S1-07 as `✅ Done`
  - Update "What To Do Next" → next = S1-09 (Docker Compose)

---

## Self-Review Against Spec

| Spec requirement | Covered by |
|------------------|------------|
| Top header: avatar initials, level badge, XP bar, streak | Task 1 |
| `(main)/layout.tsx` = header + children + BottomNav | Task 2 |
| Static hardcoded values (no API) | Task 2 (hardcoded props) |
| Mobile-first, max 430px | Global constraint + Task 2 |
| Placeholder pages for all 5 tabs | Tasks 3–7 |
| Co-located `.styles.ts` for all components | Tasks 1, 3–7 |
| Barrel exports via `index.ts` | Tasks 1, 3–7 |
| Thin route `page.tsx` files | Tasks 4–7 |
| TypeScript checks | Task 8 |
