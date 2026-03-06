-- name: CreateBoard :one
INSERT INTO boards (tenant_id, project_id, name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetBoardByID :one
SELECT * FROM boards WHERE id = $1 AND tenant_id = $2;

-- name: GetBoardsByProjectID :many
SELECT * FROM boards WHERE project_id = $1 AND tenant_id = $2 ORDER BY created_at;

-- name: UpdateBoard :one
UPDATE boards
SET name = COALESCE(sqlc.narg('name'), name),
    updated_at = NOW()
WHERE id = @id AND tenant_id = @tenant_id
RETURNING *;

-- name: DeleteBoard :exec
DELETE FROM boards WHERE id = $1 AND tenant_id = $2;
