# Frontend Scaffolding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Initialize a Next.js 14 App Router project with Tailwind CSS, configure design tokens, and set up the foundational folder structure for RitualX.

**Architecture:** A React-based mobile-first web app using Next.js App Router for routing, Tailwind for styling, React Query for server state, and Zustand for client state.

**Tech Stack:** Next.js 14, React, Tailwind CSS, TanStack React Query, Zustand, Lucide React, TypeScript.

## Global Constraints

- Must use Next.js 14 App Router (`src/app`).
- Must use TypeScript and Tailwind CSS.
- Primary Accent: Emerald green (`#10B981`).
- Body Font: Inter, Heading Font: Outfit.

---

### Task 1: Initialize Next.js Project

**Files:**
- Create: `frontend/package.json` (and standard Next.js scaffolding)

**Interfaces:**
- Consumes: None
- Produces: Base Next.js project in the `frontend` directory.

- [ ] **Step 1: Write the failing test**

```powershell
Test-Path frontend/package.json
```

- [ ] **Step 2: Run test to verify it fails**

Run: `Test-Path frontend/package.json` in the terminal.
Expected: `False`

- [ ] **Step 3: Write minimal implementation**

```powershell
mkdir frontend
cd frontend
npx -y create-next-app@14 ./ --ts --tailwind --eslint --app --src-dir --import-alias "@/*" --use-npm
```
*Note: If prompted to proceed, accept defaults.*

- [ ] **Step 4: Run test to verify it passes**

Run: `Test-Path package.json` (assuming you are inside `frontend`)
Expected: `True`

- [ ] **Step 5: Commit**

```powershell
git add .
git commit -m "chore: initialize next.js app router project"
```

---

### Task 2: Install Additional Dependencies

**Files:**
- Modify: `frontend/package.json`

**Interfaces:**
- Consumes: Next.js project
- Produces: Installed dependencies for state management and UI utilities.

- [ ] **Step 1: Write the failing test**

```powershell
Select-String -Path package.json -Pattern "zustand" -Quiet
```

- [ ] **Step 2: Run test to verify it fails**

Run: `Select-String -Path package.json -Pattern "zustand" -Quiet` inside `frontend`.
Expected: `False` (or no output)

- [ ] **Step 3: Write minimal implementation**

```powershell
npm install @tanstack/react-query zustand lucide-react clsx tailwind-merge
```

- [ ] **Step 4: Run test to verify it passes**

Run: `Select-String -Path package.json -Pattern "zustand" -Quiet` inside `frontend`.
Expected: `True`

- [ ] **Step 5: Commit**

```powershell
git add package.json package-lock.json
git commit -m "chore: install react-query, zustand, lucide-react, and ui utils"
```

---

### Task 3: Configure Tailwind Design Tokens

**Files:**
- Modify: `frontend/tailwind.config.ts`

**Interfaces:**
- Consumes: Next.js default tailwind config
- Produces: Configured theme with RitualX colors and fonts.

- [ ] **Step 1: Write the failing test**

```powershell
Select-String -Path tailwind.config.ts -Pattern "10B981" -Quiet
```

- [ ] **Step 2: Run test to verify it fails**

Run: `Select-String -Path tailwind.config.ts -Pattern "10B981" -Quiet`
Expected: `False`

- [ ] **Step 3: Write minimal implementation**

Overwrite `frontend/tailwind.config.ts` with:
```typescript
import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        background: "var(--background)",
        foreground: "var(--foreground)",
        emerald: {
          300: "#6EE7B7",
          400: "#34D399",
          600: "#059669",
          900: "#064E3B",
          DEFAULT: "#10B981", // Primary Accent
        },
        amber: {
          DEFAULT: "#F59E0B", // Streak Color
        },
        red: {
          DEFAULT: "#EF4444", // Punishment Color
        },
        purple: {
          DEFAULT: "#8B5CF6", // Gamification/XP Color
        }
      },
      fontFamily: {
        sans: ["var(--font-inter)", "sans-serif"],
        heading: ["var(--font-outfit)", "sans-serif"],
      }
    },
  },
  plugins: [],
};
export default config;
```

- [ ] **Step 4: Run test to verify it passes**

Run: `Select-String -Path tailwind.config.ts -Pattern "10B981" -Quiet`
Expected: `True`

- [ ] **Step 5: Commit**

```powershell
git add tailwind.config.ts
git commit -m "style: configure tailwind design tokens for ritualx"
```

---

### Task 4: Configure Global CSS & Typography

**Files:**
- Modify: `frontend/src/app/globals.css`
- Modify: `frontend/src/app/layout.tsx`

**Interfaces:**
- Consumes: Tailwind config
- Produces: Global styles, CSS effects (.glass-card), and loaded Google fonts (Inter, Outfit).

- [ ] **Step 1: Write the failing test**

```powershell
Select-String -Path src/app/globals.css -Pattern "glass-card" -Quiet
```

- [ ] **Step 2: Run test to verify it fails**

Run: `Select-String -Path src/app/globals.css -Pattern "glass-card" -Quiet`
Expected: `False`

- [ ] **Step 3: Write minimal implementation**

Overwrite `frontend/src/app/globals.css`:
```css
@tailwind base;
@tailwind components;
@tailwind utilities;

:root {
  --background: #09090b; /* zinc-950 */
  --foreground: #fafafa; /* zinc-50 */
}

body {
  color: var(--foreground);
  background: var(--background);
}

@layer components {
  .glass-card {
    @apply bg-zinc-900/50 backdrop-blur-md border border-white/10 rounded-xl;
  }
  .neon-glow {
    @apply shadow-[0_0_15px_rgba(16,185,129,0.5)];
  }
}
```

Overwrite `frontend/src/app/layout.tsx`:
```tsx
import type { Metadata } from "next";
import { Inter, Outfit } from "next/font/google";
import "./globals.css";

const inter = Inter({ subsets: ["latin"], variable: "--font-inter" });
const outfit = Outfit({ subsets: ["latin"], variable: "--font-outfit" });

export const metadata: Metadata = {
  title: "RitualX",
  description: "Gamified habit tracker",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={`${inter.variable} ${outfit.variable} font-sans antialiased`}>
        {children}
      </body>
    </html>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `npm run build`
Expected: Next.js builds successfully.

- [ ] **Step 5: Commit**

```powershell
git add src/app/globals.css src/app/layout.tsx
git commit -m "style: add global css effects and load fonts"
```

---

### Task 5: Setup Route Groups and Providers Folder

**Files:**
- Create: `frontend/src/app/(auth)/layout.tsx`
- Create: `frontend/src/app/(main)/layout.tsx`
- Create: `frontend/src/app/(main)/page.tsx`
- Modify: `frontend/src/app/page.tsx` (Delete it)
- Create: `frontend/src/components/providers/QueryProvider.tsx`

**Interfaces:**
- Consumes: Root layout
- Produces: App route scaffolding and React Query setup.

- [ ] **Step 1: Write the failing test**

```powershell
Test-Path src/app/(main)/layout.tsx
```

- [ ] **Step 2: Run test to verify it fails**

Run: `Test-Path src/app/(main)/layout.tsx`
Expected: `False`

- [ ] **Step 3: Write minimal implementation**

Remove default page:
```powershell
Remove-Item src/app/page.tsx -ErrorAction SilentlyContinue
```

Create `frontend/src/components/providers/QueryProvider.tsx`:
```tsx
"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState } from "react";

export default function QueryProvider({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(() => new QueryClient());
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}
```

Create folders:
```powershell
mkdir src/app/(auth)
mkdir src/app/(main)
```

Create `frontend/src/app/(auth)/layout.tsx`:
```tsx
export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return <div className="min-h-screen flex items-center justify-center">{children}</div>;
}
```

Create `frontend/src/app/(main)/layout.tsx`:
```tsx
import QueryProvider from "@/components/providers/QueryProvider";

export default function MainLayout({ children }: { children: React.ReactNode }) {
  return (
    <QueryProvider>
      <div className="min-h-screen pb-20">
        {children}
        {/* Bottom Nav will go here later */}
      </div>
    </QueryProvider>
  );
}
```

Create `frontend/src/app/(main)/page.tsx`:
```tsx
export default function Home() {
  return (
    <main className="p-4">
      <h1 className="font-heading text-3xl font-bold text-emerald-DEFAULT">RitualX</h1>
      <p className="mt-2 text-zinc-400">Daily View</p>
      <div className="mt-4 p-4 glass-card neon-glow">
        <p>Test Card</p>
      </div>
    </main>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `npm run build`
Expected: Next.js builds successfully.

- [ ] **Step 5: Commit**

```powershell
git add src/
git commit -m "feat: scaffold route groups and query provider"
```
