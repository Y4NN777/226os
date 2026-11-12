-- Minimal bootstrap query proving generation for the identity area (never selects credentials).
-- name: GetUserByEmail :one
SELECT id, email, email_verified_at, role, status, created_at FROM users WHERE email = $1;
