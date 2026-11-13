-- +goose Up
CREATE TABLE user_skills (
  user_id uuid NOT NULL,
  tag_id uuid NOT NULL,
  CONSTRAINT user_skills_pk PRIMARY KEY (user_id, tag_id)
);

CREATE TABLE audit_notes (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  msg text NOT NULL
);

CREATE TABLE members (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email text NOT NULL,
  CONSTRAINT members_email_uq UNIQUE (email)
);
