# RitualX S1-05 — Login, Refresh, Logout Design Spec

> **Date:** 2026-06-23
> **Task:** S1-05
> **Status:** Approved

## Overview
This document outlines the design for the backend authentication endpoints: Login, Token Refresh, and Logout. This builds on the existing User Registration (S1-04) by enabling users to establish and manage secure, persistent sessions.

## Goals
- Provide secure login functionality validating email and password.
- Establish short-lived access tokens (15 minutes) and long-lived refresh tokens (7 days).
- Support multiple concurrent sessions per user (e.g., mobile + desktop) with device tracking.
- Protect against XSS by using HttpOnly cookies for refresh tokens.
- Secure downstream endpoints with an Auth Middleware.

## Non-Goals
- OAuth / Social Logins (Google, Apple).
- "Manage Devices" UI or "Log out of all devices" endpoint (API support for tracking exists, but endpoints for management are deferred).
- Email verification or password reset flows (handled in separate tasks if needed).

## Approach
We will use a database-backed refresh token approach. When a user logs in, a new refresh token is generated and stored in a new `refresh_tokens` table alongside their IP and User-Agent. The refresh token is returned to the client as an `HttpOnly` cookie, while a short-lived JWT access token is returned in the JSON body. A custom Auth middleware will validate the JWT for protected routes.

## Design Details

### 1. Data Model
A new GORM model `RefreshToken` will map to the `refresh_tokens` table.

**Table: `refresh_tokens`**
- `id`: UUID (Primary Key)
- `user_id`: UUID (Foreign Key to `users`)
- `token`: VARCHAR (Unique, secure random string)
- `user_agent`: VARCHAR
- `ip_address`: VARCHAR
- `expires_at`: TIMESTAMP (7 days from creation)
- `created_at`: TIMESTAMP

### 2. API Contract & Security

**`POST /api/v1/auth/login`**
- **Request Body:** `{ "email": "...", "password": "..." }`
- **Logic:** 
  - Validate credentials against `users` table.
  - Generate Access Token (JWT, 15m) and Refresh Token (Opaque string, 7d).
  - Insert Refresh Token into DB with IP/User-Agent from request context.
- **Response:**
  - `Set-Cookie`: `refresh_token=<token>; HttpOnly; Secure; SameSite=Strict; Max-Age=604800`
  - Body: `{ "success": true, "data": { "access_token": "...", "user": { ... } } }`

**`POST /api/v1/auth/refresh`**
- **Request:** Empty body. Browser sends `refresh_token` cookie.
- **Logic:**
  - Extract cookie.
  - Query DB for token. Check if `expires_at` is in the future.
  - If valid, generate a new Access Token.
- **Response:**
  - Body: `{ "success": true, "data": { "access_token": "..." } }`

**`POST /api/v1/auth/logout`**
- **Request:** Empty body. Browser sends `refresh_token` cookie.
- **Logic:**
  - Delete the specific token row from the DB.
- **Response:**
  - `Set-Cookie`: `refresh_token=; HttpOnly; Secure; SameSite=Strict; Max-Age=0`
  - Body: `{ "success": true }`

### 3. Auth Middleware
Located at `backend/internal/middleware/auth.go`.
- Extracts token from `Authorization: Bearer <token>`.
- Validates signature using server secret.
- Validates `exp` claim.
- Injects `user_id` into `c.Locals("user_id")` on success.
- Returns `401 Unauthorized` via `handleServiceError` on failure.

## Open Questions
- Should we run a cron job to clean up expired refresh tokens from the database, or just delete them when a user attempts to use an expired one? (Decision: Delete on attempted use for now, can add a cron later if table grows too large).
