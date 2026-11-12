-- Minimal bootstrap query proving generation for the reference/taxonomy area.
-- name: GetCountry :one
SELECT code, name_fr, name_en FROM countries WHERE code = $1;
