# Code Review — S1-07: Main App Layout + Bottom Navbar

> **Date:** 2026-07-05
> **Reviewer:** Senior Code Reviewer (AI)
> **Status:** Complete

---

## Strengths

- **`"use client"` placement correct** — only `BottomNav.tsx` carries the directive; all stub pages are clean Server Components.
- **Accessibility solid** — `<nav aria-label="Main navigation">`, per-tab `aria-label={label}`, `aria-current="page"` on active tab. Meets WCAG nav landmark requirements.
- **Auth routes safe** — `(auth)` group has its own layout with no BottomNav; `(main)` layout is isolated. `/login` and `/register` will never render the navbar.
- **Active-detection logic correct** — `pathname === "/"` guard for Home prevents `/calendar` etc. from matching the root prefix.
- **Architecture fully aligned** — module barrels, page-level barrels, `*.styles.ts` co-location, thin route files, no deep cross-module imports — all correct per frontend-architecture skill.
- **Zero TypeScript errors** — `npx tsc --noEmit` passes clean.
- **Tailwind content globs cover new paths** — `tailwind.config.ts` already includes `src/modules/**` and `src/shared/**`.
- **Plan/spec alignment complete** — all 22 new files created, both modified files updated exactly as specified, nothing missing.

---

## Issues

### Critical (Must Fix)

#### 1. `text-emerald-DEFAULT` / `bg-emerald-DEFAULT` — invalid Tailwind utility

**File:** `frontend/src/shared/components/ui/BottomNav/BottomNav.styles.ts` lines 6 & 9

```ts
iconActive: "text-emerald-DEFAULT",   // WRONG
dotActive:  "bg-emerald-DEFAULT",     // WRONG
```

Tailwind drops the `-DEFAULT` suffix when generating utilities from `tailwind.config.ts`. The correct generated classes are `text-emerald` and `bg-emerald`. `text-emerald-DEFAULT` is never generated — JIT silently ignores it, icons render with no active colour. Active state is **visually identical to inactive** (broken UX).

Every other file in codebase already uses correct form: `Button.tsx:18` → `bg-emerald`, `LoginForm.tsx:18` → `text-emerald`.

**Fix:**
```diff
- iconActive: "text-emerald-DEFAULT",
+ iconActive: "text-emerald",
- dotActive:  "bg-emerald-DEFAULT",
+ dotActive:  "bg-emerald",
```

> Silent visual regression — no TS error, no runtime error, just a broken indicator.

---

### Important (Should Fix)

#### 2. Encoding corruption in `HomePlaceholderPage.tsx`

**File:** `frontend/src/modules/home/pages/home/HomePlaceholderPage.tsx` line 4

The em-dash `—` is corrupted (Windows encoding mismatch). Renders as mojibake in the browser.

**Fix:**
```diff
- <p ...>Daily view — coming soon</p>
+ <p ...>Daily view - coming soon</p>
```

#### 3. `React` type referenced but not imported in `layout.tsx`

**File:** `frontend/src/app/(main)/layout.tsx` line 3

`React.ReactNode` used without `import React` or `import type { ReactNode }`. Works with Next.js 14 JSX transform but fragile.

**Fix:**
```diff
+ import type { ReactNode } from "react";
- export default function MainLayout({ children }: { children: React.ReactNode }) {
+ export default function MainLayout({ children }: { children: ReactNode }) {
```

---

### Minor (Nice to Have)

#### 4. No `title` tooltip on nav links

`frontend/src/shared/components/ui/BottomNav/BottomNav.tsx` line 31 — icon-only nav; add `title={label}` for desktop hover discoverability.

#### 5. Home stub description inconsistent with peers

`HomePlaceholderPage.tsx` says "Daily view - coming soon"; Calendar/Stats/Quest/Profile say "Coming Soon". Standardize copy.

#### 6. No `loading.tsx` / `error.tsx` at `(main)` level

Out of scope for S1-07 (per spec Non-Goals), but ticket before real content ships.

---

## Assessment

**Ready to commit?** NO — fix Critical #1 first.

**Reasoning:** `emerald-DEFAULT` is a silent visual bug; active tab indicator is invisible at runtime. Fix lines 6 & 9 of `BottomNav.styles.ts` (2-line change) and fix encoding in `HomePlaceholderPage.tsx`, then commit.
