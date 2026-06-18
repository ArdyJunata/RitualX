# RitualX — Product Requirements Document

> **Date:** 2026-06-19
> **Status:** Approved
> **Author:** Brainstorming Session
> **Tech Stack:** Next.js + Tailwind CSS (Frontend) / Go Fiber + PostgreSQL + GORM (Backend)

---

## Overview

RitualX is a mobile web application for recording and tracking activity routines. It combines a GitHub contribution-style heatmap calendar with deep gamification mechanics (XP, quests, rewards, punishments) and social accountability features (partners, leaderboards, nudges) to drive consistent habit formation.

The app targets users who are motivated by **competitive accountability** and **game mechanics** — people who respond to peer pressure, streaks, and virtual rewards as tools for self-discipline.

## Goals

1. **Make routine tracking visually satisfying** — GitHub-style contribution heatmaps give an at-a-glance view of consistency
2. **Drive retention through gamification** — XP, leveling, quests, and cosmetic rewards create a compelling progression loop
3. **Enforce accountability through social pressure** — Partner visibility, leaderboards, and social shame punishments add real stakes
4. **Support flexible routine types** — Daily, weekly, and monthly routines with configurable frequency targets
5. **Encourage habit stacking** — Ritual Chains let users link routines into sequences for bonus rewards

## Non-Goals

- Native mobile app (iOS/Android) — mobile web + PWA only
- OAuth social login (Google, Apple) — email/password only for now
- Push notifications — can be added via Web Push API later
- Data export/import
- Admin dashboard
- Paid tiers / monetization
- Desktop-optimized UI — mobile-first only

---

## Approach: Phased Modular

Build in 3 distinct phases, each independently deployable:

- **Phase 1 (Weeks 1–3):** Core — Auth, routines, heatmaps, streaks, daily goals, statistics
- **Phase 2 (Weeks 4–6):** Gamification — XP, quests, rewards, punishments, ritual chains
- **Phase 3 (Weeks 7–8):** Social — Partners, nudges, leaderboard, social shame

---

## Design Details

### 1. Architecture

```
┌─────────────────────────────┐
│   Mobile Web (Next.js)      │
│   Tailwind CSS              │
│   PWA-ready                 │
├─────────────────────────────┤
│   REST API (JSON)           │
├─────────────────────────────┤
│   Go Fiber API Server       │
│   ├── Auth (JWT)            │
│   ├── Routine Service       │
│   ├── Streak Engine         │
│   ├── Quest Engine          │
│   ├── Reward Service        │
│   ├── Punishment Service    │
│   ├── Social Service        │
│   └── Statistics Service    │
├─────────────────────────────┤
│   PostgreSQL + GORM         │
└─────────────────────────────┘
```

#### Frontend Structure (`/frontend`)

```
frontend/
├── src/
│   ├── app/                  # Next.js App Router
│   │   ├── (auth)/           # Login, Register
│   │   ├── (main)/           # Main app layout with bottom nav
│   │   │   ├── daily/        # Daily view
│   │   │   ├── weekly/       # Weekly view
│   │   │   ├── monthly/      # Monthly view
│   │   │   ├── stats/        # Statistics
│   │   │   ├── quests/       # Quests board
│   │   │   ├── rewards/      # Reward shop
│   │   │   └── social/       # Partners & leaderboard
│   │   └── layout.tsx
│   ├── components/
│   │   ├── ui/               # Base UI components
│   │   ├── heatmap/          # GitHub contribution calendar
│   │   ├── routine/          # Routine cards, forms
│   │   ├── chain/            # Ritual chain components
│   │   └── nav/              # Floating bottom navbar
│   ├── hooks/                # Custom React hooks
│   ├── lib/                  # API client, utils
│   └── types/                # TypeScript types
```

#### Backend Structure (`/backend`)

```
backend/
├── cmd/
│   └── server/main.go        # Entry point
├── internal/
│   ├── config/                # App config
│   ├── middleware/             # Auth, CORS, logging
│   ├── model/                 # GORM models
│   ├── handler/               # HTTP handlers
│   ├── service/               # Business logic
│   ├── repository/            # DB queries
│   └── engine/
│       ├── streak/            # Streak calculation
│       ├── quest/             # Quest generation
│       └── xp/                # XP & leveling
├── migrations/                # SQL migrations
└── pkg/                       # Shared utilities
```

#### Authentication

- JWT-based authentication (access + refresh tokens)
- Email/password registration
- Access token: short-lived (15 min)
- Refresh token: long-lived (7 days), stored server-side
- Bcrypt password hashing

---

### 2. Data Model

#### Core Tables

**users**
| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| email | VARCHAR | UNIQUE, NOT NULL |
| password_hash | VARCHAR | NOT NULL |
| username | VARCHAR | UNIQUE, NOT NULL |
| display_name | VARCHAR | |
| avatar_url | VARCHAR | |
| xp | INTEGER | DEFAULT 0 |
| level | INTEGER | DEFAULT 1 |
| coins | INTEGER | DEFAULT 0 |
| title | VARCHAR | DEFAULT 'Novice' |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

**routines**
| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| user_id | UUID | FK → users |
| title | VARCHAR | NOT NULL |
| description | TEXT | |
| period_type | ENUM | 'daily', 'weekly', 'monthly' |
| target_count | INTEGER | NOT NULL |
| icon | VARCHAR | |
| color | VARCHAR | |
| is_active | BOOLEAN | DEFAULT true |
| sort_order | INTEGER | DEFAULT 0 |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

**routine_logs**
| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| routine_id | UUID | FK → routines |
| user_id | UUID | FK → users |
| logged_at | DATE | NOT NULL |
| count | INTEGER | DEFAULT 1 |
| note | TEXT | |
| created_at | TIMESTAMP | |

> UNIQUE constraint on (routine_id, logged_at)

#### Streak & Goals

**streaks**
| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| user_id | UUID | FK → users |
| routine_id | UUID | FK → routines |
| current_streak | INTEGER | DEFAULT 0 |
| longest_streak | INTEGER | DEFAULT 0 |
| last_completed | DATE | |
| updated_at | TIMESTAMP | |

**daily_goals**
| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| user_id | UUID | FK → users |
| date | DATE | NOT NULL |
| total_routines | INTEGER | |
| completed | INTEGER | DEFAULT 0 |
| is_achieved | BOOLEAN | DEFAULT false |
| created_at | TIMESTAMP | |

#### Ritual Chains

**ritual_chains**
| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| user_id | UUID | FK → users |
| name | VARCHAR | NOT NULL |
| bonus_xp | INTEGER | DEFAULT 10 |
| is_active | BOOLEAN | DEFAULT true |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

**ritual_chain_items**
| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| chain_id | UUID | FK → ritual_chains |
| routine_id | UUID | FK → routines |
| sort_order | INTEGER | NOT NULL |

**chain_completions**
| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| chain_id | UUID | FK → ritual_chains |
| user_id | UUID | FK → users |
| completed_at | DATE | NOT NULL |
| bonus_xp | INTEGER | |

#### Gamification

**quests**
| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| user_id | UUID | FK → users |
| title | VARCHAR | NOT NULL |
| description | TEXT | |
| quest_type | ENUM | 'daily', 'weekly', 'monthly', 'special' |
| criteria_type | VARCHAR | NOT NULL (e.g. 'streak_reach', 'complete_n', 'chain_complete', 'total_logs') |
| criteria_value | INTEGER | NOT NULL |
| criteria_meta | JSONB | |
| xp_reward | INTEGER | NOT NULL |
| coin_reward | INTEGER | DEFAULT 0 |
| status | ENUM | 'active', 'completed', 'expired', 'failed' |
| progress | INTEGER | DEFAULT 0 |
| expires_at | TIMESTAMP | |
| completed_at | TIMESTAMP | |
| created_at | TIMESTAMP | |

**rewards**
| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| name | VARCHAR | NOT NULL |
| description | TEXT | |
| category | ENUM | 'theme', 'avatar', 'icon', 'badge', 'graph_color' |
| asset_key | VARCHAR | NOT NULL |
| cost_coins | INTEGER | NOT NULL |
| required_lvl | INTEGER | DEFAULT 1 |
| created_at | TIMESTAMP | |

**user_rewards**
| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| user_id | UUID | FK → users |
| reward_id | UUID | FK → rewards |
| is_equipped | BOOLEAN | DEFAULT false |
| purchased_at | TIMESTAMP | |

#### Punishment

**punishments**
| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| user_id | UUID | FK → users |
| routine_id | UUID | FK → routines (nullable) |
| punishment_type | ENUM | 'stake', 'currency', 'social' |
| description | TEXT | |
| trigger_rule | VARCHAR | NOT NULL (e.g. 'miss_3_days', 'break_streak', 'miss_weekly_target') |
| stake_amount | DECIMAL | (for self-stakes) |
| coin_penalty | INTEGER | (for currency) |
| is_active | BOOLEAN | DEFAULT true |
| triggered_count | INTEGER | DEFAULT 0 |
| last_triggered | TIMESTAMP | |
| created_at | TIMESTAMP | |

**punishment_logs**
| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| punishment_id | UUID | FK → punishments |
| user_id | UUID | FK → users |
| triggered_at | TIMESTAMP | |
| details | JSONB | |
| notified_partners | BOOLEAN | DEFAULT false |

#### Social

**partnerships**
| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| user_id | UUID | FK → users |
| partner_id | UUID | FK → users |
| status | ENUM | 'pending', 'accepted', 'blocked' |
| created_at | TIMESTAMP | |

**nudges**
| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| sender_id | UUID | FK → users |
| receiver_id | UUID | FK → users |
| emoji | VARCHAR | NOT NULL |
| message | VARCHAR | |
| is_read | BOOLEAN | DEFAULT false |
| created_at | TIMESTAMP | |

**invite_codes**
| Column | Type | Constraints |
|--------|------|-------------|
| id | UUID | PK |
| user_id | UUID | FK → users |
| code | VARCHAR | UNIQUE, NOT NULL |
| max_uses | INTEGER | DEFAULT 1 |
| used_count | INTEGER | DEFAULT 0 |
| expires_at | TIMESTAMP | |
| created_at | TIMESTAMP | |

> **15 tables total.** JSONB fields on `criteria_meta` and `details` keep quest and punishment systems flexible.

---

### 3. API Design

All endpoints prefixed with `/api/v1`. Auth-protected routes require `Authorization: Bearer <JWT>`.

#### Auth
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register` | Register with email/password/username |
| POST | `/auth/login` | Login, returns access + refresh tokens |
| POST | `/auth/refresh` | Refresh access token |
| POST | `/auth/logout` | Invalidate refresh token |

#### Routines
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/routines` | List user's routines (filter: `?period=daily\|weekly\|monthly`) |
| POST | `/routines` | Create routine (title, period_type, target_count, icon, color) |
| PUT | `/routines/:id` | Update routine |
| DELETE | `/routines/:id` | Soft-delete routine (set is_active=false) |
| PATCH | `/routines/:id/reorder` | Update sort_order |

#### Routine Logs
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/routines/:id/log` | Log a completion (date, count, note) |
| DELETE | `/routines/:id/log/:logId` | Undo a log entry |
| GET | `/routines/:id/heatmap` | Get heatmap data (date→count map, filter: `?year=2026`) |
| GET | `/routines/:id/progress` | Get current period progress (completed/target) |

#### Streaks
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/streaks` | Get all routine streaks for user |
| GET | `/streaks/:routineId` | Get streak for specific routine |

> Streaks are calculated server-side on each log event — no manual CRUD.

#### Daily Goals
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/goals/daily` | Get today's goal summary (total, completed, achieved) |
| GET | `/goals/daily/history` | Get daily goal history (filter: `?from=&to=`) |

#### Ritual Chains
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/chains` | List user's chains |
| POST | `/chains` | Create chain (name, routine_ids in order) |
| PUT | `/chains/:id` | Update chain |
| DELETE | `/chains/:id` | Delete chain |
| POST | `/chains/:id/complete` | Mark chain completed for today |
| GET | `/chains/:id/history` | Get chain completion history |

#### Quests
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/quests` | List active quests (filter: `?status=active\|completed\|expired`) |
| GET | `/quests/history` | Past completed/failed quests |
| POST | `/quests/:id/claim` | Claim reward for completed quest |

> Quests are auto-generated by the quest engine — no user CRUD.

#### Rewards
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/rewards/shop` | List available rewards (filter: `?category=`) |
| POST | `/rewards/:id/purchase` | Buy reward with coins |
| GET | `/rewards/owned` | List user's owned rewards |
| PATCH | `/rewards/:id/equip` | Equip/unequip a reward |

#### Punishments
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/punishments` | List user's punishment rules |
| POST | `/punishments` | Create punishment rule |
| PUT | `/punishments/:id` | Update punishment rule |
| DELETE | `/punishments/:id` | Delete punishment rule |
| GET | `/punishments/logs` | Get triggered punishment history |

#### Social
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/social/invite` | Generate invite code |
| POST | `/social/join/:code` | Accept invite, create partnership |
| GET | `/social/partners` | List accountability partners |
| DELETE | `/social/partners/:id` | Remove partner |
| POST | `/social/nudge` | Send nudge/reaction to partner |
| GET | `/social/nudges` | Get received nudges (unread first) |
| PATCH | `/social/nudges/:id/read` | Mark nudge as read |
| GET | `/social/leaderboard` | Partner circle leaderboard |

#### Statistics
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/stats/overview` | Overall stats (total completions, active streaks, level, XP) |
| GET | `/stats/routines/:id` | Per-routine stats (completion rate, best streak, trend) |
| GET | `/stats/weekly-summary` | Weekly summary (this week vs last week) |
| GET | `/stats/monthly-summary` | Monthly summary with trends |

#### User Profile
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/profile` | Get own profile |
| PUT | `/profile` | Update display_name, avatar |
| GET | `/profile/:userId` | View partner's public profile |

> **~40 endpoints total.** All responses follow consistent format:

```json
// Success
{
  "success": true,
  "data": { ... },
  "meta": { "page": 1, "total": 50 }
}

// Error
{
  "success": false,
  "error": {
    "code": "ROUTINE_NOT_FOUND",
    "message": "Routine not found"
  }
}
```

---

### 4. UI/UX Design

#### Design Language

| Property | Value |
|----------|-------|
| **Theme** | Dark mode primary (deep blacks/grays with neon accents) |
| **Aesthetic** | Glassmorphism cards, subtle glow effects, smooth micro-animations |
| **Body Font** | Inter |
| **Heading Font** | Outfit |
| **Primary Accent** | Emerald green (#10B981) |
| **Streak Color** | Amber (#F59E0B) |
| **Punishment Color** | Red (#EF4444) |
| **XP/Quest Color** | Purple (#8B5CF6) |
| **Heatmap Scale** | 5-level: gray → light green → emerald → dark green → neon green |

#### Screen Flow

```
Splash/Auth ──→ Main App
                  │
     ┌────────────┼────────────┬──────────┐
     │            │            │          │
   Daily       Weekly      Monthly    Profile
   View         View        View      /More
     │            │            │          │
     └────────────┼────────────┘          │
                  │                       │
            Routine Detail          ┌─────┼─────┐
            (heatmap, stats,        │     │     │
             log action)          Stats Quests Social
                                        │       │
                                    Rewards  Leaderboard
                                     Shop    & Partners
```

#### Floating Bottom Navbar

Persistent across all main views. 5 items:

```
┌─────────────────────────────────────────────┐
│  📅 Daily  │  📊 Weekly  │  ＋  │  📆 Monthly  │  👤 Me  │
└─────────────────────────────────────────────┘
```

- Center **＋** button elevated (FAB style) — opens "Create Routine" or "Quick Log" bottom sheet
- Active tab has glowing dot indicator + icon fill change
- Smooth slide animation when switching tabs

#### Screen Breakdown

**1. Auth Screens (Login / Register)**
- Minimal dark screen with logo + tagline
- Email/password form with glassmorphism card
- Smooth fade-in transitions

**2. Daily View (Home)**
- Top bar: Date (today), streak fire icon 🔥 with count, XP progress bar
- Daily goal progress: Circular ring showing "5/8 routines done today"
- Routine cards list — each card shows:
  - Icon + title
  - Mini progress bar (e.g., "2/3 times today")
  - Tap to log (+1), long-press for details
  - Completion state: glowing border when done
- Ritual chains section: Collapsible, shows linked routines as horizontal pipeline with checkmarks
- Active quest banner: Slides in at top, shows nearest quest deadline

**3. Weekly View**
- 7-day column chart at top showing completion rate per day
- Per-routine heatmap rows — each routine gets a mini contribution strip (7 cells, Mon-Sun)
- Tap a routine row to expand into full heatmap detail

**4. Monthly View**
- Full GitHub contribution calendar — 4-5 week grid per routine
- Toggle between routines with horizontal swipe tabs
- Month stats summary below: completion rate %, best streak, total logs
- Color intensity = completions that day vs target

**5. Routine Detail Screen**
- Full-year heatmap (GitHub-style, scrollable)
- Streak card: Current streak 🔥, longest streak 🏆
- Period progress: "This week: 4/5 completed"
- Completion history: Scrollable log list
- Punishment rules attached to this routine (if any)
- Edit/Delete actions

**6. Create Routine (Bottom Sheet)**

Step-by-step flow:
1. Period type — Tap to select: Daily / Weekly / Monthly
2. Title & icon — Text input + icon picker grid
3. Target count — "How many times per [period]?" — number stepper
4. Color — Color picker for heatmap
5. Confirm — Preview card, save button

**7. Quests Board**
- Active quests as cards with progress bars
- Quest types badged: 🗓️ Daily, 📅 Weekly, 🏔️ Monthly, ⭐ Special
- Completed quests have a "Claim" button with reward preview
- Expired/failed quests grayed out in collapsible "Past" section

**8. Rewards Shop**
- Grid layout of purchasable items organized by category tabs
- Each item shows: preview, name, coin cost, level requirement
- Owned items marked with checkmark
- "Equip" toggle on owned items

**9. Punishments Screen**
- List of active punishment rules as cards
- Each card shows: type icon (💰 stake / 🪙 currency / 👥 social), trigger rule in plain English, penalty amount
- "Triggered X times" counter with shame color (red intensity increases)
- "Add Punishment" FAB button

**10. Social / Partners**
- Partner list with mini profile cards (avatar, username, streak count)
- Nudge buttons — quick emoji reactions per partner
- Invite code — Generate / share button
- Leaderboard tab: Ranked list by XP with weekly delta arrows (↑↓)

**11. Statistics**
- Overview cards: Total completions, active days, total XP earned
- Completion rate chart: Line graph over weeks/months
- Routine comparison: Bar chart comparing completion rates across routines
- Best/worst days: Heatmap showing which days of the week are strongest

**12. Profile / Me**
- Avatar, display name, username
- Level badge + XP progress to next level
- Title display
- Equipped cosmetics preview
- Settings (notifications, account, logout)

#### Key Interactions

| Interaction | Behavior |
|------------|----------|
| Log routine | Single tap → +1 count with satisfying haptic + confetti micro-animation |
| Complete daily goal | Celebration animation (particle burst) when ring fills |
| Chain complete | Sequential checkmark cascade animation → bonus XP toast |
| Streak milestone | Special animation at 7, 30, 100, 365 days |
| Punishment triggered | Screen edge briefly flashes red + notification |

---

### 5. Gamification Engine

#### XP Sources

| Action | XP Earned |
|--------|-----------|
| Log a routine completion | +10 XP |
| Complete daily goal (all routines done) | +25 XP |
| Complete a ritual chain | +15 XP (configurable per chain) |
| Streak milestone — 7 days | +50 XP |
| Streak milestone — 30 days | +200 XP |
| Streak milestone — 100 days | +500 XP |
| Streak milestone — 365 days | +2000 XP |
| Complete a quest | Quest-defined (50–500 XP) |

#### Leveling Formula

```
XP required for level N = 100 * N^1.5 (rounded)
```

| Level | Total XP Required |
|-------|-------------------|
| 1 | 0 |
| 2 | 100 |
| 3 | 283 |
| 5 | 559 |
| 10 | 1,585 |
| 20 | 4,472 |
| 50 | 17,678 |

#### Title Progression

| Level Range | Title |
|-------------|-------|
| 1–4 | Novice |
| 5–9 | Apprentice |
| 10–19 | Disciple |
| 20–34 | Ritualist |
| 35–49 | Master |
| 50+ | Ascended |

#### Coin Economy

**Earning:**
| Action | Coins |
|--------|-------|
| Complete daily goal | +5 coins |
| Complete quest | +10–50 coins (quest-defined) |
| 7-day streak milestone | +15 coins |
| 30-day streak milestone | +50 coins |

**Spending:**
| Item | Cost |
|------|------|
| Theme packs | 200–500 coins |
| Avatar items | 50–150 coins |
| Custom routine icons | 100 coins |
| Graph color schemes | 150 coins |
| Profile badges | 75–300 coins |

> Balanced so a consistent user earns roughly 1 shop item per week.

#### Quest Generation Engine

The backend runs a quest scheduler (cron job or triggered on login).

**Daily Quests** (3 generated per day, refresh at midnight):
- "Complete all your daily routines today" → 30 XP, 5 coins
- "Log [routine_name] 3 times today" → 20 XP
- "Complete a ritual chain" → 25 XP, 5 coins

**Weekly Quests** (2 generated on Monday):
- "Achieve daily goal 5 out of 7 days" → 100 XP, 20 coins
- "Maintain streak on [routine_name] for 7 days" → 75 XP, 15 coins

**Monthly Quests** (1 generated on 1st of month):
- "Log 100 total completions this month" → 300 XP, 50 coins
- "Complete all ritual chains 20 times" → 250 XP, 40 coins

**Special Quests** (triggered by milestones):
- "Reach level 10" → 200 XP, 30 coins
- "Add 5 accountability partners" → 100 XP, 20 coins
- "Create your first ritual chain" → 50 XP, 10 coins

**Adaptive Difficulty:**
- Quest engine reads user's average completion rate over last 14 days
- Rate > 80%: quests require higher targets (stretch goals)
- Rate < 40%: quests use lower targets (achievable wins)
- Never generates quests for inactive/paused routines

**Quest State Machine:**

```
  [Generated] → active
       │
  ┌────┴────┐
  │         │
completed  expired (past deadline, not met)
  │         │
claim      failed (logged in punishment_logs if linked)
  │
[Rewards distributed]
```

#### Punishment Engine

**Trigger Rules:**

| Rule Key | Description |
|----------|-------------|
| `miss_days:N` | Miss logging a routine for N consecutive days |
| `break_streak:N` | Break a streak that was ≥ N days |
| `miss_weekly_target` | Fail to hit weekly target_count |
| `miss_monthly_target` | Fail to hit monthly target_count |
| `miss_daily_goal:N` | Miss daily goal for N days in a row |

**Execution by Type:**

1. **Self-imposed Stakes (`stake`):**
   - User defines the stake text + dollar amount when creating
   - When triggered: push notification + prominent banner on home screen
   - Stays visible until user acknowledges ("I paid my stake" button)
   - Logged in punishment_logs for accountability partner visibility

2. **Currency Loss (`currency`):**
   - Auto-deducts coins from user balance
   - If balance goes negative, earned coins are garnished until repaid
   - Toast notification: "💀 Lost 20 coins — missed 3 days of Meditation"

3. **Social Shame (`social`):**
   - When triggered: all accountability partners receive a notification
   - Partner's view of user shows a "shame badge" for 24 hours
   - Visible on leaderboard as a penalty marker
   - Partners can send 💀 nudges in response

**Evaluation Schedule:**
- Daily triggers: checked at midnight (server timezone per user)
- Weekly/monthly triggers: checked at period boundary
- Streak-break triggers: checked in real-time on every log event

#### Streak Calculation Logic

```
On each routine log:
  1. Get last_completed date from streaks table
  2. Based on period_type:
     - daily:   if last_completed == yesterday → current_streak++
                if last_completed == today → no change
                else → current_streak = 1 (streak broken)
     - weekly:  if last_completed is in previous week → current_streak++
                if last_completed is in current week → no change
                else → current_streak = 1
     - monthly: same logic with months
  3. Update longest_streak if current > longest
  4. Update last_completed = today
  5. Check streak milestones → award XP/coins
  6. Check punishment triggers → execute if broken
```

---

### 6. SDLC: Sprint Agile Methodology

#### 6.1 Methodology Overview

| Property | Value |
|----------|-------|
| **Framework** | Scrum (adapted for solo/small team) |
| **Sprint Duration** | 1 week (Monday–Sunday) |
| **Total Sprints** | 8 sprints across 3 phases |
| **Story Point Scale** | Fibonacci (1, 2, 3, 5, 8, 13) |
| **Sprint Velocity Target** | ~20–25 story points per sprint |

#### 6.2 Sprint Ceremonies

| Ceremony | When | Duration | Purpose |
|----------|------|----------|---------|
| **Sprint Planning** | Monday morning | 30 min | Select stories from backlog, define sprint goal, assign points |
| **Daily Standup** | Every day | 10 min | What I did, what I'll do, blockers |
| **Sprint Review** | Sunday | 30 min | Demo completed features, verify acceptance criteria |
| **Sprint Retrospective** | Sunday (after review) | 15 min | What went well, what to improve, action items |

#### 6.3 Definition of Done (DoD)

A user story is considered **Done** when ALL of the following are met:

- [ ] Code is written and compiles without errors
- [ ] All acceptance criteria are verified
- [ ] API endpoints return correct responses for happy path and error cases
- [ ] UI is responsive on mobile viewports (360px–428px)
- [ ] Code follows project conventions (Go: handler→service→repository, Next.js: App Router)
- [ ] No console errors or warnings in browser
- [ ] Database migrations run cleanly
- [ ] Feature is manually tested end-to-end

#### 6.4 Product Backlog & Sprint Plan

---

### PHASE 1: CORE (Sprints 1–3)

---

#### Sprint 1 — Foundation & Auth
**Sprint Goal:** Project scaffolding complete, user can register and login.

| ID | User Story | Acceptance Criteria | Points |
|----|-----------|---------------------|--------|
| S1-01 | As a developer, I want the backend project scaffolded with Go Fiber, GORM, and PostgreSQL so that I have a working development environment | • Go Fiber server starts on configured port • GORM connects to PostgreSQL • Health check endpoint `/api/v1/health` returns 200 • Project follows the defined folder structure (`cmd/`, `internal/`, `pkg/`) • Environment config loads from `.env` | 3 |
| S1-02 | As a developer, I want the frontend project scaffolded with Next.js and Tailwind so that I have a working development environment | • Next.js app runs with App Router • Tailwind CSS is configured with the design token colors • Project follows the defined folder structure (`app/`, `components/`, `lib/`) • Google Fonts (Inter, Outfit) loaded | 3 |
| S1-03 | As a developer, I want database migrations for users table so that user data can be persisted | • `users` table created with all columns per data model • Migration runs forward and rolls back cleanly • UUID generation works | 2 |
| S1-04 | As a new user, I want to register with email, username, and password so that I can create an account | • `POST /api/v1/auth/register` creates user and returns access + refresh JWT tokens • Password is hashed with bcrypt • Duplicate email/username returns 409 error • Validation: email format, password min 8 chars, username 3-20 chars | 5 |
| S1-05 | As a returning user, I want to login with email and password so that I can access my account | • `POST /api/v1/auth/login` returns access + refresh tokens on valid credentials • Invalid credentials return 401 • Access token expires in 15 minutes • `POST /api/v1/auth/refresh` returns new access token with valid refresh token • `POST /api/v1/auth/logout` invalidates refresh token | 5 |
| S1-06 | As a user, I want to see login and register screens so that I can authenticate in the app | • Login screen: email + password fields, submit button, link to register • Register screen: email + username + password fields, submit button, link to login • Dark theme with glassmorphism card styling • Form validation with inline error messages • Successful auth redirects to daily view | 5 |
| S1-07 | As a user, I want a main app layout with floating bottom navbar so that I can navigate between views | • Bottom navbar with 5 items: Daily, Weekly, ＋ (FAB), Monthly, Me • Active tab shows glowing dot indicator • Center ＋ button is elevated FAB style • Smooth slide animation on tab switch • Navbar persists across all main views • Routing works for `/daily`, `/weekly`, `/monthly`, `/me` | 5 |

**Sprint 1 Total: 28 points**

---

#### Sprint 2 — Routines & Daily View
**Sprint Goal:** User can create routines, log completions, and see their daily view.

| ID | User Story | Acceptance Criteria | Points |
|----|-----------|---------------------|--------|
| S2-01 | As a developer, I want database migrations for routines and routine_logs tables | • `routines` and `routine_logs` tables created per data model • UNIQUE constraint on `(routine_id, logged_at)` in routine_logs • Foreign keys to users table | 2 |
| S2-02 | As a user, I want to create a new routine so that I can start tracking an activity | • `POST /api/v1/routines` creates routine with title, period_type, target_count, icon, color • Validation: title required, period_type must be daily/weekly/monthly, target_count ≥ 1 • Returns created routine with UUID | 3 |
| S2-03 | As a user, I want to view, update, and delete my routines | • `GET /api/v1/routines` returns user's active routines, filterable by `?period=` • `PUT /api/v1/routines/:id` updates routine fields • `DELETE /api/v1/routines/:id` soft-deletes (sets is_active=false) • `PATCH /api/v1/routines/:id/reorder` updates sort_order • Users can only access their own routines (403 otherwise) | 5 |
| S2-04 | As a user, I want to log a routine completion so that my progress is recorded | • `POST /api/v1/routines/:id/log` creates a log entry with date and count • Duplicate log for same routine+date increments count (or returns conflict based on design) • `DELETE /api/v1/routines/:id/log/:logId` removes a log entry • Returns updated progress for the current period | 5 |
| S2-05 | As a user, I want a create routine bottom sheet with step-by-step flow | • ＋ FAB opens bottom sheet overlay • Step 1: Select period type (Daily/Weekly/Monthly) with tap selection • Step 2: Enter title + pick icon from grid • Step 3: Set target count with number stepper • Step 4: Pick heatmap color • Step 5: Preview card + confirm button • Smooth transitions between steps • Success closes sheet and shows new routine in list | 8 |
| S2-06 | As a user, I want to see my daily view with routine cards so that I can track today's progress | • Daily view shows today's date in top bar • Routine cards list showing: icon, title, mini progress bar ("2/3 times today") • Tap card → logs +1 completion with confetti micro-animation • Completed routines show glowing border • Cards ordered by sort_order • Empty state shown when no routines exist | 8 |

**Sprint 2 Total: 31 points** *(stretch sprint — S2-05 and S2-06 can overflow to Sprint 3)*

---

#### Sprint 3 — Heatmap, Streaks & Statistics
**Sprint Goal:** GitHub contribution heatmaps, streak tracking, and statistics are live.

| ID | User Story | Acceptance Criteria | Points |
|----|-----------|---------------------|--------|
| S3-01 | As a developer, I want database migrations for streaks and daily_goals tables | • Tables created per data model • Indexes on `(user_id, routine_id)` for streaks | 1 |
| S3-02 | As a user, I want my streaks calculated automatically when I log a routine | • Streak engine runs on each log event • Daily: consecutive days increment streak, gap resets to 1 • Weekly: consecutive weeks increment streak • Monthly: consecutive months increment streak • `longest_streak` updated when `current_streak` exceeds it • `GET /api/v1/streaks` returns all routine streaks • `GET /api/v1/streaks/:routineId` returns specific streak | 5 |
| S3-03 | As a user, I want daily goal tracking so that I can see my overall daily progress | • System calculates total active routines and completed count per day • `GET /api/v1/goals/daily` returns today's summary • `GET /api/v1/goals/daily/history` returns date range history • `is_achieved` set to true when completed == total_routines | 3 |
| S3-04 | As a user, I want to see heatmap data for my routines so the contribution calendar can be rendered | • `GET /api/v1/routines/:id/heatmap?year=2026` returns `{ "2026-01-15": 3, "2026-01-16": 1, ... }` • `GET /api/v1/routines/:id/progress` returns `{ completed: 4, target: 5, period: "weekly" }` • Data scoped to authenticated user only | 3 |
| S3-05 | As a user, I want basic statistics endpoints | • `GET /api/v1/stats/overview` returns total completions, active streak count, current level, total XP • `GET /api/v1/stats/routines/:id` returns completion rate, best streak, weekly trend • `GET /api/v1/stats/weekly-summary` returns this week vs last week comparison • `GET /api/v1/stats/monthly-summary` returns monthly trends | 5 |
| S3-06 | As a user, I want to see a GitHub-style contribution heatmap for each routine | • Heatmap component renders a 52-week × 7-day grid (or current year to date) • 5-level color intensity based on completions vs target • Tooltip on cell shows date and count • Scrollable for full-year view • Renders in < 100ms | 8 |
| S3-07 | As a user, I want weekly and monthly views showing my routines | • Weekly view: 7-day column chart at top + per-routine mini heatmap strips (Mon-Sun) • Monthly view: full 4-5 week contribution grid per routine, horizontal swipe tabs to switch routines • Month stats summary: completion rate %, best streak, total logs | 5 |
| S3-08 | As a user, I want a routine detail screen with full heatmap and streak info | • Full-year heatmap (GitHub-style) • Streak card: current 🔥 and longest 🏆 • Period progress bar • Completion history log list (scrollable) • Edit/Delete actions | 5 |
| S3-09 | As a user, I want a statistics screen with charts and analytics | • Overview cards: total completions, active days, total XP • Completion rate line graph over weeks/months • Routine comparison bar chart • Best/worst day-of-week heatmap | 5 |
| S3-10 | As a user, I want daily goal progress shown in the daily view | • Circular progress ring at top of daily view ("5/8 routines done") • Ring fills with animation as routines are completed • Celebration particle burst animation when ring reaches 100% | 3 |

**Sprint 3 Total: 43 points** *(large sprint — lower-priority items S3-07, S3-09 can overflow to Sprint 4)*

---

### PHASE 2: GAMIFICATION (Sprints 4–6)

---

#### Sprint 4 — XP, Leveling & Ritual Chains
**Sprint Goal:** XP/leveling system is live, users can create and complete ritual chains.

| ID | User Story | Acceptance Criteria | Points |
|----|-----------|---------------------|--------|
| S4-01 | As a developer, I want database migrations for ritual_chains, ritual_chain_items, and chain_completions tables | • Tables created per data model • Foreign keys to routines and users | 2 |
| S4-02 | As a user, I want to earn XP for completing routines and milestones | • XP engine awards XP per the defined XP Sources table • User's `xp` and `level` fields update in real-time • Level calculated using formula: `100 * N^1.5` • Title updates automatically based on level range • XP changes return in API responses for toast display | 5 |
| S4-03 | As a user, I want to earn coins for daily goals and streak milestones | • Coin engine awards coins per the Coin Economy table • User's `coins` field updates on earn events • Coin balance cannot go below negative (unless punishment garnishment) • Coin changes return in API responses | 3 |
| S4-04 | As a user, I want to create ritual chains linking my routines together | • `POST /api/v1/chains` creates chain with name + ordered routine_ids • `GET /api/v1/chains` lists user's active chains • `PUT /api/v1/chains/:id` updates chain name/routines/order • `DELETE /api/v1/chains/:id` deletes chain • Validation: minimum 2 routines per chain, all routines must belong to user | 5 |
| S4-05 | As a user, I want to complete a ritual chain and earn bonus XP | • `POST /api/v1/chains/:id/complete` marks chain completed for today • System verifies all routines in chain have been logged today before allowing completion • Awards chain's `bonus_xp` to user • `GET /api/v1/chains/:id/history` returns completion history • Prevents duplicate completion on same day | 5 |
| S4-06 | As a user, I want to see an XP progress bar and level badge in the UI | • XP progress bar shown in daily view top bar • Shows current level + XP progress to next level (e.g., "Level 7 — 340/500 XP") • Level-up animation when threshold crossed • Title displayed on profile | 3 |
| S4-07 | As a user, I want a ritual chain UI to create and complete chains | • Chain creation: select routines from list, drag to reorder, name the chain • Daily view: collapsible chains section showing horizontal pipeline • Each routine in pipeline shows checkmark when logged today • "Complete Chain" button activates when all routines logged • Sequential checkmark cascade animation on completion + bonus XP toast | 8 |

**Sprint 4 Total: 31 points**

---

#### Sprint 5 — Quests & Rewards
**Sprint Goal:** System-generated quests and cosmetic reward shop are live.

| ID | User Story | Acceptance Criteria | Points |
|----|-----------|---------------------|--------|
| S5-01 | As a developer, I want database migrations for quests, rewards, and user_rewards tables | • Tables created per data model • JSONB column for criteria_meta • Indexes on `(user_id, status)` for quests | 2 |
| S5-02 | As a developer, I want a quest generation engine that creates quests automatically | • Cron job (or login trigger) generates: 3 daily quests at midnight, 2 weekly quests on Monday, 1 monthly quest on 1st • Quest templates use user's actual routine names and data • Adaptive difficulty: reads 14-day average completion rate, adjusts targets up (>80%) or down (<40%) • Skips inactive/paused routines • Expired quests auto-marked as 'expired' or 'failed' | 8 |
| S5-03 | As a user, I want to view my active and past quests | • `GET /api/v1/quests` returns active quests with progress • `GET /api/v1/quests?status=completed` returns completed quests • `GET /api/v1/quests/history` returns all past quests • Each quest shows: title, description, type badge, progress/criteria_value, XP/coin reward, expiry | 3 |
| S5-04 | As a user, I want quest progress to update automatically as I complete routines | • Quest progress increments automatically when matching criteria events occur • E.g., "Complete 5 daily routines" → progress increments each daily goal achievement • Quest status changes to 'completed' when progress ≥ criteria_value • Completed quests show "Claim" option | 5 |
| S5-05 | As a user, I want to claim quest rewards | • `POST /api/v1/quests/:id/claim` awards XP and coins to user • Can only claim quests with status 'completed' • Quest status changes to 'claimed' after successful claim • Returns updated user XP/coins/level | 3 |
| S5-06 | As a developer, I want reward shop data seeded in the database | • Seed script populates rewards table with: 5+ themes, 10+ avatar items, 5+ icon packs, 5+ badges, 5+ graph color schemes • Each reward has name, description, category, asset_key, cost_coins, required_lvl • Costs balanced per coin economy design | 3 |
| S5-07 | As a user, I want to browse, purchase, and equip cosmetic rewards | • `GET /api/v1/rewards/shop` lists all rewards, filterable by category • `POST /api/v1/rewards/:id/purchase` deducts coins and adds to user_rewards • Insufficient coins returns 400 error • Level requirement enforced • `GET /api/v1/rewards/owned` lists purchased rewards • `PATCH /api/v1/rewards/:id/equip` toggles equip state • Only one item per category can be equipped at a time | 5 |
| S5-08 | As a user, I want a quests board screen in the app | • Active quests displayed as cards with progress bars • Quest type badges: 🗓️ Daily, 📅 Weekly, 🏔️ Monthly, ⭐ Special • "Claim" button with reward preview on completed quests • Collapsible "Past" section for expired/failed quests (grayed out) • Quest completion celebration animation | 5 |
| S5-09 | As a user, I want a rewards shop screen in the app | • Grid layout organized by category tabs (themes, avatars, icons, badges, graph colors) • Each item shows: visual preview, name, coin cost, level requirement • Owned items marked with checkmark • "Equip"/"Unequip" toggle on owned items • Purchase confirmation modal showing coin balance | 5 |

**Sprint 5 Total: 39 points** *(large sprint — S5-08 and S5-09 can overflow to Sprint 6)*

---

#### Sprint 6 — Punishments & Polish
**Sprint Goal:** Punishment system is live, all gamification animations polished.

| ID | User Story | Acceptance Criteria | Points |
|----|-----------|---------------------|--------|
| S6-01 | As a developer, I want database migrations for punishments and punishment_logs tables | • Tables created per data model • JSONB column for details in punishment_logs | 2 |
| S6-02 | As a user, I want to create punishment rules for my routines | • `POST /api/v1/punishments` creates rule with: type (stake/currency/social), trigger_rule, amounts • `GET /api/v1/punishments` lists active rules • `PUT /api/v1/punishments/:id` updates rule • `DELETE /api/v1/punishments/:id` removes rule • Validation: coin_penalty must be positive, trigger_rule must be valid format | 5 |
| S6-03 | As a developer, I want a punishment evaluation engine that triggers punishments automatically | • Daily cron evaluates all active punishment rules at midnight • `miss_days:N` checks N consecutive days without logs • `break_streak:N` checks in real-time when streak ≥ N is broken • `miss_weekly_target` checks at end of week • `miss_monthly_target` checks at end of month • `miss_daily_goal:N` checks N consecutive days without achieving daily goal • Creates punishment_log entry on trigger • Increments triggered_count on punishment | 8 |
| S6-04 | As a user, I want stake punishments to show a prominent banner until acknowledged | • When stake punishment triggers: banner appears on home screen • Banner shows stake description + amount • "I paid my stake" acknowledgment button dismisses banner • Logged in punishment_logs for partner visibility | 3 |
| S6-05 | As a user, I want currency punishments to auto-deduct coins | • When currency punishment triggers: coins auto-deducted from balance • If balance goes negative: future earned coins garnished until repaid • Toast notification: "💀 Lost X coins — [reason]" | 3 |
| S6-06 | As a user, I want to view my punishment history | • `GET /api/v1/punishments/logs` returns triggered punishment history • Each entry shows: punishment type, trigger rule, when triggered, details | 2 |
| S6-07 | As a user, I want a punishments screen in the app | • List of active punishment rules as cards • Each card shows: type icon (💰/🪙/👥), trigger rule in plain English, penalty amount • "Triggered X times" counter with red intensity scaling • "Add Punishment" FAB → creation form (select type, routine, trigger, amounts) | 5 |
| S6-08 | As a user, I want celebration and shame animations throughout the app | • Streak milestone: special animation at 7, 30, 100, 365 days • Punishment triggered: screen edge flashes red briefly • Level up: burst animation + title change announcement • All animations are smooth (60fps) and non-blocking | 5 |

**Sprint 6 Total: 33 points**

---

### PHASE 3: SOCIAL (Sprints 7–8)

---

#### Sprint 7 — Partners & Nudges
**Sprint Goal:** Users can invite accountability partners, view their progress, and send reactions.

| ID | User Story | Acceptance Criteria | Points |
|----|-----------|---------------------|--------|
| S7-01 | As a developer, I want database migrations for partnerships, nudges, and invite_codes tables | • Tables created per data model • Index on `(user_id, partner_id)` for partnerships | 2 |
| S7-02 | As a user, I want to generate an invite code to share with potential partners | • `POST /api/v1/social/invite` generates unique code with configurable max_uses and expiry • Code is short (6-8 characters), alphanumeric, URL-safe • Returns shareable link/code | 3 |
| S7-03 | As a user, I want to join a partner's circle using their invite code | • `POST /api/v1/social/join/:code` creates partnership (status: 'accepted') • Invalid/expired/maxed-out codes return appropriate errors • Prevents duplicate partnerships • Cannot partner with yourself | 3 |
| S7-04 | As a user, I want to view and manage my accountability partners | • `GET /api/v1/social/partners` returns partner list with profile info (avatar, username, current streak, level) • `DELETE /api/v1/social/partners/:id` removes partnership (both directions) • `GET /api/v1/profile/:userId` returns partner's public profile (heatmap, streaks, level, title) | 5 |
| S7-05 | As a user, I want to send nudges/reactions to my partners | • `POST /api/v1/social/nudge` sends nudge with emoji + optional message • `GET /api/v1/social/nudges` returns received nudges (unread first) • `PATCH /api/v1/social/nudges/:id/read` marks as read • Rate limit: max 10 nudges per partner per day | 3 |
| S7-06 | As a user, I want social shame punishments to notify my partners | • When social punishment triggers: `notified_partners` set to true in punishment_log • Partners can see shame events on the user's profile • Shame badge visible on partner's card for 24 hours • Partners can send 💀 nudges in response | 5 |
| S7-07 | As a user, I want a social/partners screen in the app | • Partner list with mini profile cards (avatar, username, streak 🔥, level badge) • Quick nudge buttons: row of emoji reactions per partner (🔥💪💀👏) • "Invite" button → generate code / share link • Tap partner card → view their public profile (heatmaps, streaks, stats) • Notification badge for unread nudges • Emoji picker for custom nudge messages | 8 |

**Sprint 7 Total: 29 points**

---

#### Sprint 8 — Leaderboard, Profile & Launch Polish
**Sprint Goal:** Competitive leaderboard is live, profile screen complete, app polished for launch.

| ID | User Story | Acceptance Criteria | Points |
|----|-----------|---------------------|--------|
| S8-01 | As a user, I want a leaderboard showing rankings among my partners | • `GET /api/v1/social/leaderboard` returns ranked list of user + all partners • Ranked by total XP • Each entry shows: rank, avatar, username, XP, level, current longest streak • Weekly delta arrows (↑↓) showing rank change from last week • Current user highlighted in the list | 5 |
| S8-02 | As a user, I want leaderboard penalty markers for punished users | • Users who had punishments triggered in the last 7 days show a penalty marker • Shame badges visible next to their rank • Hover/tap shows punishment details | 3 |
| S8-03 | As a user, I want a profile / "Me" screen | • Shows: avatar, display name, username • Level badge + XP progress bar to next level • Title display (e.g., "Ritualist") • Equipped cosmetics preview (active theme, avatar items, badge) • `PUT /api/v1/profile` updates display_name and avatar | 5 |
| S8-04 | As a user, I want a settings section in my profile | • Notification preferences (toggle nudge notifications) • Account settings (change password, change email) • Logout button (invalidates tokens, redirects to login) • "Delete Account" with confirmation modal | 5 |
| S8-05 | As a user, I want a notification center for nudges and system events | • Nudges from partners shown with sender avatar + emoji + timestamp • Punishment trigger notifications • Quest completion/expiry notifications • Unread count badge on "Me" tab • Mark all as read action | 5 |
| S8-06 | As a user, I want the leaderboard integrated into the social screen | • Leaderboard as a tab within the social screen (Partners | Leaderboard) • Animated rank transitions on load • Pull-to-refresh updates rankings • Empty state for users with no partners | 3 |
| S8-07 | As a developer, I want end-to-end testing and launch polish | • All 12 screens render without errors • All API endpoints tested for happy path + error cases • Responsive on 360px–428px viewports • No console errors in production build • PWA manifest + service worker for add-to-homescreen • Performance: API < 200ms (p95), heatmap render < 100ms | 8 |

**Sprint 8 Total: 34 points**

---

#### 6.5 Sprint Summary

| Sprint | Phase | Goal | Story Points |
|--------|-------|------|-------------|
| Sprint 1 | Core | Foundation & Auth | 28 |
| Sprint 2 | Core | Routines & Daily View | 31 |
| Sprint 3 | Core | Heatmap, Streaks & Statistics | 43 |
| Sprint 4 | Gamification | XP, Leveling & Ritual Chains | 31 |
| Sprint 5 | Gamification | Quests & Rewards | 39 |
| Sprint 6 | Gamification | Punishments & Polish | 33 |
| Sprint 7 | Social | Partners & Nudges | 29 |
| Sprint 8 | Social | Leaderboard, Profile & Launch | 34 |
| **Total** | | | **268 points** |

> **Note:** Sprints 3 and 5 are overloaded. If velocity falls short, lower-priority stories (marked in sprint descriptions) should overflow to the next sprint. The sprint retrospective will adjust velocity targets based on actual throughput.

#### 6.6 Release Milestones

| Milestone | After Sprint | What's Live |
|-----------|-------------|-------------|
| **Alpha Release** | Sprint 3 | Core routine tracking with heatmaps, streaks, statistics |
| **Beta Release** | Sprint 6 | Full gamification: XP, quests, rewards, punishments, chains |
| **v1.0 Launch** | Sprint 8 | Complete app: social, leaderboard, polished UI |

---

### 7. Non-Functional Requirements

| Area | Target |
|------|--------|
| Performance | API response < 200ms (p95), heatmap render < 100ms |
| Mobile responsive | Optimized for 360px–428px viewport (iPhone SE → iPhone Pro Max) |
| PWA | Service worker for offline log queueing, add-to-homescreen |
| Security | Bcrypt password hashing, JWT rotation, rate limiting on auth endpoints |
| Database indexes | `routine_logs(routine_id, logged_at)`, `streaks(user_id, routine_id)`, `quests(user_id, status)`, `partnerships(user_id, partner_id)` |
| Timezone | All dates stored as UTC, converted to user's local timezone on display |
| Cron Jobs | Quest generation (daily midnight), punishment evaluation (daily midnight), streak verification (daily midnight) |

---

## Open Questions

*None — all design decisions have been resolved during brainstorming.*
