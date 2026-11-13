-- name: AddSkill :exec
INSERT INTO user_skills (user_id, tag_id) VALUES ($1, $2);

-- name: AddNote :exec
INSERT INTO audit_notes (msg) VALUES ($1);

-- name: AddMember :exec
INSERT INTO members (email) VALUES ($1);
