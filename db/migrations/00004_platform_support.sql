-- Migration: 00004_platform_support
-- Purpose: Create the cross-cutting platform tables: PostgreSQL-backed rate limiting, idempotency keys and the append-only audit log.
-- Tables/indexes: rate_limit_buckets (UNLOGGED) + rate_limit_buckets_window_idx; idempotency_keys + idempotency_keys_expiry_idx; audit_log + audit_log_target_idx, audit_log_actor_idx, audit_log_occurred_idx.
-- Lock risk: none (new tables on an empty database). rate_limit_buckets is UNLOGGED: not replicated and truncated after a crash (accepted, ADR-0003).
-- Data volume: empty at creation; rate-limit and idempotency rows are purged by expiry jobs, audit_log grows with moderator actions only.
-- Transaction: single transaction (goose default).
-- Compatibility window: none (new tables). Append-only audit_log is enforced by role grants (INSERT, SELECT only for the application role) applied by ops, not by this migration.
-- Restore plan: forward-only. On an empty database, recreate it; otherwise restore the pre-migration pg_dump -Fc backup or redeploy the previous image by digest. UNLOGGED data is rebuildable and excluded from restore expectations.
-- Owner: Y4NN77
-- Design: docs/spec-kit/DATABASE_DESIGN.md 3.2 (rate_limit_buckets) and 3.8. Forward-only: no Down section.
-- +goose Up
CREATE UNLOGGED TABLE rate_limit_buckets (
  scope        text NOT NULL,        -- e.g. 'signin:ip', 'signin:account', 'reset:ip', 'write:user', 'github_import:user'
  key_hash     bytea NOT NULL,
  window_start timestamptz NOT NULL,
  hits         integer NOT NULL DEFAULT 1,
  CONSTRAINT rate_limit_buckets_pk PRIMARY KEY (scope, key_hash, window_start),
  CONSTRAINT rate_limit_buckets_scope_ck CHECK (scope ~ '^[a-z_]+(:[a-z_]+)?$' AND char_length(scope) <= 64),
  CONSTRAINT rate_limit_buckets_hits_ck CHECK (hits > 0)
);
CREATE INDEX rate_limit_buckets_window_idx ON rate_limit_buckets (window_start);   -- purge

CREATE TABLE idempotency_keys (
  user_id         uuid NOT NULL,
  idempotency_key text NOT NULL,
  method          text NOT NULL,
  route           text NOT NULL,            -- route template, not the concrete path
  request_hash    bytea NOT NULL,
  state           text NOT NULL DEFAULT 'in_progress',
  response_status smallint,
  response_body   jsonb,
  created_at      timestamptz NOT NULL DEFAULT now(),
  completed_at    timestamptz,
  expires_at      timestamptz NOT NULL,
  CONSTRAINT idempotency_keys_pk PRIMARY KEY (user_id, idempotency_key),
  CONSTRAINT idempotency_keys_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT idempotency_keys_key_ck CHECK (char_length(idempotency_key) BETWEEN 8 AND 128),
  CONSTRAINT idempotency_keys_method_ck CHECK (method IN ('POST','PATCH','DELETE')),
  CONSTRAINT idempotency_keys_route_ck CHECK (char_length(route) <= 200),
  CONSTRAINT idempotency_keys_hash_ck CHECK (octet_length(request_hash) = 32),
  CONSTRAINT idempotency_keys_state_ck CHECK (state IN ('in_progress','completed')),
  CONSTRAINT idempotency_keys_completed_ck CHECK (
    (state = 'completed') = (completed_at IS NOT NULL AND response_status IS NOT NULL)),
  CONSTRAINT idempotency_keys_body_ck CHECK (response_body IS NULL OR (jsonb_typeof(response_body) = 'object' AND octet_length(response_body::text) <= 16384)),
  CONSTRAINT idempotency_keys_expiry_ck CHECK (expires_at > created_at)
);
CREATE INDEX idempotency_keys_expiry_idx ON idempotency_keys (expires_at);   -- purge job; retention 24 h (ADR-0004)

CREATE TABLE audit_log (
  id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  occurred_at timestamptz NOT NULL DEFAULT now(),
  actor_id    uuid,
  actor_role  text,
  action      text NOT NULL,
  target_type text,
  target_id   uuid,                         -- no FK by design: the row must survive any future hard deletion
  outcome     text NOT NULL DEFAULT 'success',
  reason      text,
  request_id  text,
  details     jsonb NOT NULL DEFAULT '{}'::jsonb,
  CONSTRAINT audit_log_actor_fk FOREIGN KEY (actor_id) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT audit_log_actor_role_ck CHECK (actor_role IS NULL OR actor_role IN ('member','moderator','system')),
  CONSTRAINT audit_log_action_ck CHECK (action ~ '^[a-z_]+\.[a-z_]+$' AND char_length(action) <= 60),
  CONSTRAINT audit_log_target_type_ck CHECK (target_type IS NULL OR target_type IN
    ('user','question','answer','resource','provider','project','contribution_need','application','article','series','lesson','relation','report','session')),
  CONSTRAINT audit_log_outcome_ck CHECK (outcome IN ('success','denied','failed')),
  CONSTRAINT audit_log_reason_ck CHECK (reason IS NULL OR char_length(reason) <= 500),
  CONSTRAINT audit_log_request_id_ck CHECK (request_id IS NULL OR char_length(request_id) <= 64),
  CONSTRAINT audit_log_details_ck CHECK (jsonb_typeof(details) = 'object' AND octet_length(details::text) <= 4096)
);
CREATE INDEX audit_log_target_idx ON audit_log (target_type, target_id, occurred_at DESC) WHERE target_id IS NOT NULL;
CREATE INDEX audit_log_actor_idx ON audit_log (actor_id, occurred_at DESC) WHERE actor_id IS NOT NULL;
CREATE INDEX audit_log_occurred_idx ON audit_log (occurred_at DESC);
