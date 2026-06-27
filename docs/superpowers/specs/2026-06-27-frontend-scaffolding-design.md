# RitualX Frontend Scaffolding Design

> **Date:** 2026-06-27
> **Topic:** Frontend Scaffolding
> **Status:** Approved

## Overview
This specification details the foundational setup, tooling, and folder structure for the RitualX frontend application. It serves as the implementation guide for task S1-02. The frontend will be a mobile-first, highly interactive PWA designed to communicate with the existing Go Fiber backend.

## Goals
- Initialize a modern Next.js environment that supports server-side rendering and client-side interactivity.
- Establish a robust state management and data fetching strategy.
- Define a strict folder structure to maintain organization as the codebase grows.
- Implement the baseline design tokens (colors, typography, effects) specified in the main product PRD.

## Non-Goals
- Implementing the actual UI screens (Login, Home, Quests, etc.) - this spec only covers scaffolding the environment and layout wrappers.
- Backend modifications.
- Deployment configuration.

## Approach & Tooling
We will use the following technology stack for the frontend:
- **Framework:** Next.js 14+ (App Router) initialized with TypeScript.
- **Styling:** Tailwind CSS.
- **Data Fetching:** TanStack React Query (for API communication and caching).
- **Client State:** Zustand (for simple UI state).
- **Icons:** `lucide-react`.
- **Formatting/Linting:** ESLint + Prettier (with `prettier-plugin-tailwindcss`).

## Design Details

### 1. Folder Structure
The Next.js application will reside in a `frontend` directory at the root of the repository. The internal structure will utilize route groups to separate layouts.

```text
frontend/
├── src/
│   ├── app/                    # Next.js App Router
│   │   ├── (auth)/             # Route group for auth flows (no nav bar)
│   │   ├── (main)/             # Route group for main app (with bottom nav)
│   │   │   └── page.tsx        # Daily View (Home)
│   │   ├── layout.tsx          # Root layout (Providers, Fonts)
│   │   └── globals.css         # Global Tailwind imports & CSS variables
│   ├── components/             # Reusable React components
│   │   ├── ui/                 # Base elements (buttons, inputs)
│   │   ├── layout/             # Navbar, headers
│   │   └── providers/          # React Query & Zustand wrappers
│   ├── hooks/                  # Custom React hooks
│   ├── lib/                    # Utilities and configuration
│   │   ├── api.ts              # Axios/Fetch setup for Go backend
│   │   ├── utils.ts            # Helper functions
│   │   └── store.ts            # Zustand stores
│   └── types/                  # TypeScript interfaces matching backend
├── tailwind.config.ts          # Tailwind theme config
└── package.json
```

### 2. Design Tokens (Tailwind Config)
The `tailwind.config.ts` will be extended to support the specific aesthetic requirements:

**Typography:**
- Loaded via `next/font/google` in `layout.tsx`.
- Body: `Inter` (sans-serif default).
- Headings: `Outfit` (sans-serif display).

**Color Palette:**
- Backgrounds: `zinc-900` to `zinc-950`.
- Primary: Emerald `#10B981` (for primary actions and active states).
- Streak/Warning: Amber `#F59E0B`.
- Punishment: Red `#EF4444`.
- Gamification/XP: Purple `#8B5CF6`.

**Heatmap Scale Custom Classes:**
- Level 0: `bg-zinc-800`
- Level 1: `bg-emerald-900`
- Level 2: `bg-emerald-600`
- Level 3: `bg-emerald-400`
- Level 4: `bg-emerald-300`

### 3. Global CSS Effects
Defined in `globals.css` for easy reuse:
- `.glass-card`: Semi-transparent background, backdrop blur, and light inner border.
- `.neon-glow`: Emerald drop-shadow for active states.

### 4. Dependencies
The following core packages must be installed:
- `next`, `react`, `react-dom`
- `tailwindcss`, `postcss`, `autoprefixer`
- `@tanstack/react-query`
- `zustand`
- `lucide-react`
- `clsx`, `tailwind-merge` (for dynamic class utilities)

## Open Questions
- None. Requirements are clear and bounded to scaffolding.
