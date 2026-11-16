-- name: CreateThing :one
INSERT INTO things (id, name) VALUES ($1, $2) RETURNING id;

-- name: ListThings :many
SELECT id, name FROM things ORDER BY name, id LIMIT 50;
