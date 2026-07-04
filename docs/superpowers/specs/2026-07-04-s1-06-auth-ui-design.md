# RitualX S1-06 — Auth UI Screens Design Spec

> **Date:** 2026-07-04
> **Task:** S1-06
> **Status:** Approved

## Overview

Implement Login and Register UI screens for RitualX. Screens live inside the existing `(auth)` Next.js route group. Minimal dark aesthetic (Linear/Vercel style). Calls existing backend auth endpoints. Access token stored in `localStorage`.

## Goals

- Login screen: email + password → calls `POST /api/v1/auth/login`
- Register screen: display name + username + email + password + confirm password → calls `POST /api/v1/auth/register`
- Store `access_token` in `localStorage` on success
- Redirect to `/` on success
- Client-side validation + server-side field error mapping
- Minimal dark UI: zinc-950 bg, glass-card form, emerald accents, mobile-first

## Non-Goals

- OAuth / social login
- Email verification flow
- "Forgot password" flow
- Logout UI (handled in S1-07+)
- Desktop-optimized layout (mobile-first only for now)

## Approach

Page-level forms (Approach A). Each auth screen = standalone Next.js page. Form logic in co-located hook. API calls in `modules/auth/api.ts`. Pages are thin wrappers over components.

## Design Details

### 1. File Structure

```
frontend/src/
├── app/
│   └── (auth)/
│       ├── layout.tsx               # existing
│       ├── login/
│       │   └── page.tsx             # Login page (thin wrapper)
│       └── register/
│           └── page.tsx             # Register page (thin wrapper)
│
└── modules/
    └── auth/
        ├── api.ts                   # login(), register() using apiClient
        ├── types.ts                 # LoginRequest, RegisterRequest, AuthUser, LoginResponse, RegisterResponse
        ├── hooks/
        │   ├── useLoginForm.ts      # form state, client validation, submit, error mapping, redirect
        │   └── useRegisterForm.ts   # form state, client validation, submit, error mapping, redirect
        └── components/
            ├── LoginForm.tsx        # pure UI, driven by useLoginForm hook
            └── RegisterForm.tsx     # pure UI, driven by useRegisterForm hook
```

### 2. API Contract & Types

**`auth/types.ts`**

```ts
// Requests
interface LoginRequest {
  email: string
  password: string
}

interface RegisterRequest {
  display_name: string
  username: string
  email: string
  password: string
}

// Domain
interface AuthUser {
  id: string
  display_name: string
  username: string
  email: string
  created_at: string
}

// Responses
interface LoginResponse {
  access_token: string
  user: AuthUser
}

interface RegisterResponse {
  access_token: string
  user: AuthUser
}
```

**`auth/api.ts`**

```ts
// POST /api/v1/auth/login
login(body: LoginRequest): Promise<LoginResponse>

// POST /api/v1/auth/register
register(body: RegisterRequest): Promise<RegisterResponse>
```

Both use `apiClient` from `@/shared/api-client`. Throw `ApiError` on failure.

**localStorage key:** `ritualx_access_token`

### 3. Validation Rules

**Client-side (pre-submit):**

| Field              | Rule                                     |
|--------------------|------------------------------------------|
| `display_name`     | required, 2–50 chars                     |
| `username`         | required, 3–30 chars, `/^[a-z0-9_]+$/i` |
| `email`            | required, valid email format             |
| `password`         | required, min 8 chars                    |
| `confirm_password` | must match `password`                    |

**Server-side error mapping:**

- Catch `ApiError` in hook
- Read `ApiError.fields` (Record<string, string[]>)
- Merge into per-field error state
- Client errors cleared on each resubmit attempt

**UX rules:**
- Inline errors below each field (`text-red-400 text-sm`)
- Submit button disabled while in-flight
- Loading spinner inside button during submit
- On success → `localStorage.setItem('ritualx_access_token', access_token)` → `router.push('/')`

### 4. UI Design

**Style:** Minimal dark, mobile-first (max-width 420px, centered).

**Color tokens (from existing Tailwind config):**
- Page bg: `bg-zinc-950`
- Card: `glass-card` class (zinc-900/80, backdrop-blur, rounded-2xl, border-zinc-800)
- Inputs: `bg-zinc-800`, `border-zinc-700`, focus ring → `ring-emerald`
- Button: `bg-emerald-DEFAULT`, hover → `bg-emerald-600`, disabled → `opacity-50`
- Error text: `text-red-400 text-sm`
- Logo "X": `text-emerald-DEFAULT`

**Typography:**
- Logo wordmark: `font-heading` (Outfit), bold
- Body/labels: `font-sans` (Inter)

**Login page layout:**
```
┌─────────────────────────────┐
│   RitualX  (logo)           │
│   "Track. Level up.         │
│    Dominate."               │
│                             │
│   ┌─────────────────────┐   │
│   │  Email              │   │
│   │  [error]            │   │
│   │  Password           │   │
│   │  [error]            │   │
│   │  [Login button]     │   │
│   └─────────────────────┘   │
│                             │
│  "Don't have an account?"   │
│  → Register                 │
└─────────────────────────────┘
```

**Register page layout:**
```
┌─────────────────────────────┐
│   RitualX  (logo)           │
│   "Your ritual starts here."│
│                             │
│   ┌─────────────────────┐   │
│   │  Display Name       │   │
│   │  Username           │   │
│   │  Email              │   │
│   │  Password           │   │
│   │  Confirm Password   │   │
│   │  [Register button]  │   │
│   └─────────────────────┘   │
│                             │
│  "Already have an account?" │
│  → Login                    │
└─────────────────────────────┘
```

**Animation:** page fade-in on mount — `opacity-0 → opacity-100`, 200ms ease-in.

## Open Questions

- None. All requirements are bounded and clear.
