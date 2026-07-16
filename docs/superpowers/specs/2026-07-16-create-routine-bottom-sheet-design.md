# S2-05 — Create Routine Bottom Sheet: Design Spec

> **Date:** 2026-07-16
> **Status:** Approved
> **Sprint:** S2

---

## Overview

A 5-step wizard bottom sheet that lets users create a new routine.
Triggered by the FAB `+` button. Slides up from the bottom of the screen.
Each step shows one decision at a time.

---

## Goals

- User can create a routine via a mobile-friendly step-by-step wizard.
- Sheet is reusable (generic `BottomSheet` component).
- Preview card (Step 5) reusable in S2-06 daily view.

## Non-Goals

- Edit routine (future task).
- Drag-to-reorder inside the sheet.

---

## Approach

Multi-step wizard inside a draggable bottom sheet.
State lives in `useCreateRoutine` hook.
Step components are pure UI — receive props, emit callbacks.

---

## Design Details

### Sheet Behavior

- Opens when user taps the FAB `+` in the center of `BottomNav`.
- Covers ~85% screen height.
- Backdrop (semi-transparent) behind sheet; tap backdrop or drag down to close.
- If any field has been filled, close shows a discard confirmation (`window.confirm`).
- Progress bar (5 filled/empty circles) at top shows current step.

### Step Sequence

| Step | Screen | API field |
|------|--------|-----------|
| 1 | Period type (Daily / Weekly / Monthly) | `period_type` |
| 2 | Title (text input) + Icon (emoji grid) | `title`, `icon` |
| 3 | Target count (stepper) | `target_count` |
| 4 | Color (12 preset swatches) | `color` |
| 5 | Preview card + Create button | submits POST |

### Step 1 — Period Type

Three large tappable option cards (full-width).
Selected option gets an emerald ring + checkmark.
"Next" button disabled until selection made.

Options: `daily` / `weekly` / `monthly`

### Step 2 — Title & Icon

- Text input (autofocused on step enter) for routine title.
- 6×4 emoji grid below (24 emojis). Selected emoji gets emerald ring.
- "Next" disabled until title.trim().length > 0 AND icon selected.

Emoji list (24):
`🏃 🏋️ 📚 💧 🧘 🎯 🎮 🍎 ✍️ 🎵 🧹 😴 🚴 🧠 💊 🌿 🔥 ⚡ 🎨 📷 🛁 🍳 🧘 💪`

### Step 3 — Target Count

- Label: "How many times per [period_type]?"
- Number stepper: `[ − ]  N  [ + ]`
- Min = 1, Max = 99, Default = 1.
- "Next" always enabled (default is valid).

### Step 4 — Color

- 12 preset color swatches in a 6×2 grid.
- Selected color gets a white ring + scale-up.
- "Next" disabled until color selected.

Color values (hex):
`#ef4444` `#f97316` `#eab308` `#22c55e` `#10b981` `#06b6d4`
`#3b82f6` `#8b5cf6` `#ec4899` `#f43f5e` `#a3e635` `#ffffff`

### Step 5 — Confirm

- Heading: "Ready to go!"
- `RoutinePreviewCard` showing filled-in data.
- "Back" button + "Create Routine" primary button.
- On submit: call `POST /api/v1/routines`, show loading spinner on button.
- On success: close sheet.
- On error: show error message below button (do not close sheet).

### RoutinePreviewCard

Displays:
- Emoji icon (left)
- Routine title (bold)
- Color swatch dot (right, small circle)
- Period type + target count (subtitle row)
- Empty progress bar (0 / target_count today)

Reused in S2-06.

---

## API Contract

```
POST /api/v1/routines
Authorization: Bearer <access_token>
Content-Type: application/json

Body:
{
  "title": string,
  "period_type": "daily" | "weekly" | "monthly",
  "target_count": number,
  "icon": string,
  "color": string
}

Success 201:
{
  "data": {
    "id": string,
    "user_id": string,
    "title": string,
    "period_type": string,
    "target_count": number,
    "icon": string,
    "color": string,
    "sort_order": number,
    "is_active": boolean,
    "created_at": string,
    "updated_at": string
  }
}
```

---

## File Locations

```
frontend/src/modules/routines/
  ├── index.ts
  ├── types.ts
  ├── api.ts
  ├── hooks/
  │   └── useCreateRoutine.ts
  └── components/
      ├── CreateRoutineSheet.tsx
      ├── RoutinePreviewCard.tsx
      └── steps/
          ├── StepPeriod.tsx
          ├── StepTitleIcon.tsx
          ├── StepTarget.tsx
          ├── StepColor.tsx
          └── StepConfirm.tsx

frontend/src/shared/components/ui/
  └── BottomSheet.tsx

frontend/src/shared/components/ui/BottomNav/
  └── BottomNav.tsx   ← MODIFY: add FAB + button, wire sheet open state
```

---

## Animations

- Sheet: CSS `transform: translateY` (300ms ease-out open, 250ms close).
- Step: slide left (forward) / right (back) via CSS transition.
- Stepper dots: `transition: background-color 150ms`.
- Create button: brief scale pulse on success.
