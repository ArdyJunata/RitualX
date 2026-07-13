# S2-02 — Create Routine API Design

> **Date:** 2026-07-13

## Overview

`POST /api/v1/routines` — authenticated user creates a new routine. Returns 201 with the created routine.

## Goals

- Repository: `Create`, `FindByID` methods on `RoutineRepository`
- Service: `RoutineService.Create` with validation
- Handler: `CreateRoutine` fiber handler
- Route wired behind `RequireAuth` middleware in `main.go`
- Error codes mapped in `response.go`

## Non-Goals

- GET / PUT / DELETE (S2-03)
- Log completion (S2-04)

## Request / Response

### Request
`POST /api/v1/routines`
`Authorization: Bearer <access_token>`

```json
{
  "title": "Morning Run",
  "description": "Run 5km every morning",
  "period_type": "daily",
  "target_count": 1,
  "icon": "🏃",
  "color": "#FF6B6B"
}
```

### Validation Rules
| Field | Rule |
|-------|------|
| `title` | required, 1–100 chars |
| `period_type` | required, one of: `daily`, `weekly`, `monthly` |
| `target_count` | required, integer ≥ 1 |
| `description` | optional |
| `icon` | optional |
| `color` | optional |

### Success Response — 201
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "title": "Morning Run",
    "description": "Run 5km every morning",
    "period_type": "daily",
    "target_count": 1,
    "icon": "🏃",
    "color": "#FF6B6B",
    "is_active": true,
    "sort_order": 0,
    "created_at": "...",
    "updated_at": "..."
  }
}
```

### Error Responses
| Code | HTTP | Meaning |
|------|------|---------|
| `VALIDATION_ERROR` | 400 | field validation failed |
| `INVALID_REQUEST` | 400 | malformed JSON body |
| `UNAUTHORIZED` | 401 | missing/invalid token |
| `INTERNAL_ERROR` | 500 | db error |

## Architecture

```
handler.CreateRoutine
  → service.RoutineService.Create(userID, req)
    → validate fields
    → repository.RoutineRepository.Create(&routine)
  → 201 + routine
```

## Files

| Action | File |
|--------|------|
| Create | `backend/internal/repository/routine.go` |
| Create | `backend/internal/service/routine.go` |
| Create | `backend/internal/handler/routine.go` |
| Modify | `backend/internal/handler/response.go` (add error codes) |
| Modify | `backend/cmd/server/main.go` (wire route) |

## Open Questions

- None
