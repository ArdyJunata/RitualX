# S1-07: Main App Layout + Bottom Navbar — Implementation Plan

> **For agentic workers:** Use the executing-plans approach. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add floating glassmorphism bottom navbar with 5 tabs and wire stub pages for all main routes.

**Architecture:** BottomNav lives in `shared/components/ui/BottomNav/`. Each tab route maps to a module under `src/modules/`. App Router `(main)/layout.tsx` mounts the nav. Stub pages render "Coming Soon".

**Tech Stack:** Next.js 14 App Router, Tailwind CSS, Lucide React, `usePathname` (next/navigation)

## Global Constraints

- Next.js 14 App Router — `src/app/` routing
- Tailwind CSS only — no inline `style={{}}` ever
- All styles in co-located `*.styles.ts` files
- Colors: emerald (`#10B981`), zinc-500, zinc-900, white/10 only from tailwind.config.ts tokens
- Icon library: `lucide-react` (already installed)
- `"use client"` only on components that use hooks (BottomNav uses `usePathname`)
- Stub pages = Server Components (no `"use client"`)
- Module barrel = `index.ts` at module root, only cross-module entry point
- No commits until instructed
- Frontend root: `C:/Users/naats/Documents/github.com/ArdyJunata/RitualX/frontend`

---

### Task 1: BottomNav component

**Files:**
- Create: `src/shared/components/ui/BottomNav/BottomNav.styles.ts`
- Create: `src/shared/components/ui/BottomNav/BottomNav.tsx`
- Create: `src/shared/components/ui/BottomNav/index.ts`

**Interfaces:**
- Consumes: `usePathname()` from `next/navigation`, Lucide icons, styles from `BottomNav.styles.ts`
- Produces: `export { BottomNav }` via `index.ts` — used by `(main)/layout.tsx`

- [ ] **Step 1: Create `BottomNav.styles.ts`**

```ts
// src/shared/components/ui/BottomNav/BottomNav.styles.ts

export const bottomNavStyles = {
  container:
    "fixed bottom-6 left-1/2 -translate-x-1/2 z-50 flex items-center gap-6 px-6 py-3 rounded-full bg-zinc-900/70 backdrop-blur-xl border border-white/10 shadow-lg shadow-black/40",
  tab: "flex flex-col items-center gap-0.5 transition-colors duration-200",
  icon: "w-6 h-6",
  iconActive: "text-emerald-DEFAULT",
  iconInactive: "text-zinc-500",
  dot: "w-1.5 h-1.5 rounded-full",
  dotActive: "bg-emerald-DEFAULT",
  dotInactive: "bg-transparent",
} as const;
```

- [ ] **Step 2: Create `BottomNav.tsx`**

```tsx
// src/shared/components/ui/BottomNav/BottomNav.tsx
"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Home, CalendarDays, BarChart2, Sword, User } from "lucide-react";
import { bottomNavStyles as s } from "./BottomNav.styles";

const NAV_TABS = [
  { label: "Home",     href: "/",          Icon: Home        },
  { label: "Calendar", href: "/calendar",  Icon: CalendarDays },
  { label: "Stats",    href: "/stats",     Icon: BarChart2   },
  { label: "Quest",    href: "/quest",     Icon: Sword       },
  { label: "Profile",  href: "/profile",   Icon: User        },
] as const;

function isActive(pathname: string, href: string): boolean {
  if (href === "/") return pathname === "/";
  return pathname.startsWith(href);
}

export function BottomNav() {
  const pathname = usePathname();

  return (
    <nav aria-label="Main navigation" className={s.container}>
      {NAV_TABS.map(({ label, href, Icon }) => {
        const active = isActive(pathname, href);
        return (
          <Link
            key={href}
            href={href}
            aria-label={label}
            aria-current={active ? "page" : undefined}
            className={s.tab}
          >
            <Icon className={`${s.icon} ${active ? s.iconActive : s.iconInactive}`} />
            <span className={`${s.dot} ${active ? s.dotActive : s.dotInactive}`} />
          </Link>
        );
      })}
    </nav>
  );
}
```

- [ ] **Step 3: Create `index.ts` barrel**

```ts
// src/shared/components/ui/BottomNav/index.ts
export { BottomNav } from "./BottomNav";
```

- [ ] **Step 4: Verify no TypeScript errors**

Run from `frontend/`:
```bash
npx tsc --noEmit
```
Expected: no errors related to BottomNav files.

---

### Task 2: Home module stub

**Files:**
- Create: `src/modules/home/pages/home/HomePlaceholderPage.tsx`
- Create: `src/modules/home/pages/home/index.ts`
- Create: `src/modules/home/index.ts`

**Interfaces:**
- Consumes: nothing external
- Produces: `export { HomePlaceholderPage }` via module barrel

- [ ] **Step 1: Create `HomePlaceholderPage.tsx`**

```tsx
// src/modules/home/pages/home/HomePlaceholderPage.tsx

export function HomePlaceholderPage() {
  return (
    <main className="flex flex-col items-center justify-center min-h-[60vh] gap-2 p-4">
      <h1 className="font-heading text-2xl font-bold text-zinc-100">Home</h1>
      <p className="text-zinc-500 text-sm">Daily view — coming soon</p>
    </main>
  );
}
```

- [ ] **Step 2: Create page-level barrel**

```ts
// src/modules/home/pages/home/index.ts
export { HomePlaceholderPage } from "./HomePlaceholderPage";
```

- [ ] **Step 3: Create module barrel**

```ts
// src/modules/home/index.ts
export { HomePlaceholderPage } from "./pages/home";
```

---

### Task 3: Calendar, Stats, Quest, Profile module stubs

**Files:**
- Create: `src/modules/calendar/pages/calendar/CalendarPage.tsx`
- Create: `src/modules/calendar/pages/calendar/index.ts`
- Create: `src/modules/calendar/index.ts`
- Create: `src/modules/stats/pages/stats/StatsPage.tsx`
- Create: `src/modules/stats/pages/stats/index.ts`
- Create: `src/modules/stats/index.ts`
- Create: `src/modules/quest/pages/quest/QuestPage.tsx`
- Create: `src/modules/quest/pages/quest/index.ts`
- Create: `src/modules/quest/index.ts`
- Create: `src/modules/profile/pages/profile/ProfilePage.tsx`
- Create: `src/modules/profile/pages/profile/index.ts`
- Create: `src/modules/profile/index.ts`

**Interfaces:**
- Consumes: nothing external
- Produces: `CalendarPage`, `StatsPage`, `QuestPage`, `ProfilePage` via module barrels

- [ ] **Step 1: Calendar module**

```tsx
// src/modules/calendar/pages/calendar/CalendarPage.tsx
export function CalendarPage() {
  return (
    <main className="flex flex-col items-center justify-center min-h-[60vh] gap-2 p-4">
      <h1 className="font-heading text-2xl font-bold text-zinc-100">Calendar</h1>
      <p className="text-zinc-500 text-sm">Coming Soon</p>
    </main>
  );
}
```

```ts
// src/modules/calendar/pages/calendar/index.ts
export { CalendarPage } from "./CalendarPage";
```

```ts
// src/modules/calendar/index.ts
export { CalendarPage } from "./pages/calendar";
```

- [ ] **Step 2: Stats module**

```tsx
// src/modules/stats/pages/stats/StatsPage.tsx
export function StatsPage() {
  return (
    <main className="flex flex-col items-center justify-center min-h-[60vh] gap-2 p-4">
      <h1 className="font-heading text-2xl font-bold text-zinc-100">Stats</h1>
      <p className="text-zinc-500 text-sm">Coming Soon</p>
    </main>
  );
}
```

```ts
// src/modules/stats/pages/stats/index.ts
export { StatsPage } from "./StatsPage";
```

```ts
// src/modules/stats/index.ts
export { StatsPage } from "./pages/stats";
```

- [ ] **Step 3: Quest module**

```tsx
// src/modules/quest/pages/quest/QuestPage.tsx
export function QuestPage() {
  return (
    <main className="flex flex-col items-center justify-center min-h-[60vh] gap-2 p-4">
      <h1 className="font-heading text-2xl font-bold text-zinc-100">Quest</h1>
      <p className="text-zinc-500 text-sm">Coming Soon</p>
    </main>
  );
}
```

```ts
// src/modules/quest/pages/quest/index.ts
export { QuestPage } from "./QuestPage";
```

```ts
// src/modules/quest/index.ts
export { QuestPage } from "./pages/quest";
```

- [ ] **Step 4: Profile module**

```tsx
// src/modules/profile/pages/profile/ProfilePage.tsx
export function ProfilePage() {
  return (
    <main className="flex flex-col items-center justify-center min-h-[60vh] gap-2 p-4">
      <h1 className="font-heading text-2xl font-bold text-zinc-100">Profile</h1>
      <p className="text-zinc-500 text-sm">Coming Soon</p>
    </main>
  );
}
```

```ts
// src/modules/profile/pages/profile/index.ts
export { ProfilePage } from "./ProfilePage";
```

```ts
// src/modules/profile/index.ts
export { ProfilePage } from "./pages/profile";
```

---

### Task 4: App Router route files

**Files:**
- Create: `src/app/(main)/calendar/page.tsx`
- Create: `src/app/(main)/stats/page.tsx`
- Create: `src/app/(main)/quest/page.tsx`
- Create: `src/app/(main)/profile/page.tsx`
- Modify: `src/app/(main)/page.tsx`

**Interfaces:**
- Consumes: module barrels from Task 2 and Task 3
- Produces: Next.js route pages

- [ ] **Step 1: Calendar route**

```tsx
// src/app/(main)/calendar/page.tsx
import { CalendarPage } from "@/modules/calendar";
export default function Page() { return <CalendarPage />; }
```

- [ ] **Step 2: Stats route**

```tsx
// src/app/(main)/stats/page.tsx
import { StatsPage } from "@/modules/stats";
export default function Page() { return <StatsPage />; }
```

- [ ] **Step 3: Quest route**

```tsx
// src/app/(main)/quest/page.tsx
import { QuestPage } from "@/modules/quest";
export default function Page() { return <QuestPage />; }
```

- [ ] **Step 4: Profile route**

```tsx
// src/app/(main)/profile/page.tsx
import { ProfilePage } from "@/modules/profile";
export default function Page() { return <ProfilePage />; }
```

- [ ] **Step 5: Update Home route**

Replace entire content of `src/app/(main)/page.tsx`:

```tsx
// src/app/(main)/page.tsx
import { HomePlaceholderPage } from "@/modules/home";
export default function Page() { return <HomePlaceholderPage />; }
```

---

### Task 5: Wire layout — mount BottomNav

**Files:**
- Modify: `src/app/(main)/layout.tsx`

**Interfaces:**
- Consumes: `BottomNav` from `@/shared/components/ui/BottomNav`
- Produces: shell layout with nav rendered on all `(main)` routes

- [ ] **Step 1: Update `(main)/layout.tsx`**

Replace entire content:

```tsx
// src/app/(main)/layout.tsx
import { BottomNav } from "@/shared/components/ui/BottomNav";

export default function MainLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen pb-28">
      {children}
      <BottomNav />
    </div>
  );
}
```

- [ ] **Step 2: Run TypeScript check**

```bash
npx tsc --noEmit
```
Expected: 0 errors.

- [ ] **Step 3: Start dev server and manual verify**

```bash
npm run dev
```

Open `http://localhost:3000` in browser. Verify:
- [ ] Floating pill navbar visible at bottom
- [ ] Home tab active (emerald icon + dot) on `/`
- [ ] Clicking Calendar → `/calendar` → CalendarPage renders, Calendar tab active
- [ ] Clicking Stats → `/stats` → StatsPage renders, Stats tab active
- [ ] Clicking Quest → `/quest` → QuestPage renders, Quest tab active
- [ ] Clicking Profile → `/profile` → ProfilePage renders, Profile tab active
- [ ] Navbar does not appear on `/login` or `/register`
- [ ] Page content not hidden behind navbar (pb-28 clearance works)
