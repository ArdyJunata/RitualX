# S2-03 — Routines CRUD + Reorder API Design

> **Date:** 2026-07-14

## Overview

Add GET (list + single), PUT (update), DELETE (soft-delete), and PATCH reorder endpoints to the existing routines resource.

- `GET    /api/v1/routines`           — list user active routines ordered by sort_order
- `GET    /api/v1/routines/:id`       — get single routine (must belong to user)
- `PUT    /api/v1/routines/:id`       — update routine fields
- `DELETE /api/v1/routines/:id`       — soft-delete (set is_active = false)
- `PATCH  /api/v1/routines/reorder`   — update sort_order for a list of IDs

All routes behind RequireAuth middleware. User can only operate on their own routines.

## Goals

- Repository: FindAllByUserID, FindByIDAndUserID, Update, SoftDelete, Reorder methods
- Service: List, GetByID, Update, Delete, Reorder methods with ownership checks
- Handler: ListRoutines, GetRoutine, UpdateRoutine, DeleteRoutine, ReorderRoutines handlers
- Routes wired in main.go
- Unit tests for service validation logic

## Non-Goals

- Log completion (S2-04)
- Hard delete
- Filter by period_type (future)

## Endpoints

### GET /api/v1/routines
Returns all active routines for the authenticated user, sorted by sort_order ASC.

**Response 200:**
```json
{ "success": true, "data": [ { "id": "uuid", "user_id": "uuid", "title": "Morning Run", "description": "Run 5km", "period_type": "daily", "target_count": 1, "icon": "🏃", "color": "#FF6B6B", "is_active": true, "sort_order": 0, "created_at": "...", "updated_at": "..." } ] }
```

### GET /api/v1/routines/:id
Returns a single routine by ID (must belong to authenticated user).

**Error Responses:**
| Code | HTTP | Meaning |
|------|------|---------|
| NOT_FOUND | 404 | routine not found or belongs to another user |
| UNAUTHORIZED | 401 | missing/invalid token |

### PUT /api/v1/routines/:id
Update any subset of mutable fields. All fields optional — only provided fields updated.

**Request:**
```json
{ "title": "Evening Run", "description": "Updated desc", "period_type": "daily", "target_count": 2, "icon": "🌙", "color": "#4ECDC4" }
```

**Validation (applied only to provided fields):**
| Field | Rule |
|-------|------|
| title | if present: 1–100 chars |
| period_type | if present: must be daily, weekly, or monthly |
| target_count | if present: integer >= 1 |

### DELETE /api/v1/routines/:id
Soft-delete: sets is_active = false.

**Response 200:** `{ "success": true, "data": null }`

### PATCH /api/v1/routines/reorder
Update sort_order for multiple routines in a single transaction.

**Request:**
```json
{ "order": [ { "id": "uuid-1", "sort_order": 0 }, { "id": "uuid-2", "sort_order": 1 } ] }
```

**Response 200:** `{ "success": true, "data": null }`

## Architecture

```
handler.ListRoutines → service.List(userID) → repo.FindAllByUserID(userID) → 200
handler.GetRoutine → service.GetByID(userID, id) → repo.FindByIDAndUserID → nil=NOT_FOUND → 200
handler.UpdateRoutine → service.Update(userID, id, req) → ownership check → validate → repo.Update → 200
handler.DeleteRoutine → service.Delete(userID, id) → ownership check → repo.SoftDelete → 200
handler.ReorderRoutines → service.Reorder(userID, req) → validate → repo.Reorder (transaction) → 200
```

## Key Design Decisions

- Ownership check via FindByIDAndUserID — single query with id + user_id filter; nil = 404
- Soft delete only — is_active=false; FindAllByUserID filters is_active=true
- Reorder is transactional — all updates in one db.Transaction; any ID not owned = rollback = NOT_FOUND
- PUT uses pointer fields — UpdateRoutineRequest uses *string/*int so nil = not provided; only non-nil fields applied

## Files

| Action | File |
|--------|------|
| Modify | backend/internal/repository/routine.go |
| Modify | backend/internal/service/routine.go |
| Modify | backend/internal/handler/routine.go |
| Modify | backend/cmd/server/main.go |
| Modify | backend/internal/service/routine_test.go |

## Open Questions

- None
