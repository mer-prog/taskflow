-- name: CreateColumn :one
INSERT INTO columns (tenant_id, board_id, name, position, color, wip_limit)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetColumnsByBoardID :many
SELECT * FROM columns WHERE board_id = $1 AND tenant_id = $2 ORDER BY position;

-- name: UpdateColumn :one
UPDATE columns
SET name = COALESCE(sqlc.narg('name'), name),
    color = COALESCE(sqlc.narg('color'), color),
    wip_limit = COALESCE(sqlc.narg('wip_limit'), wip_limit),
    updated_at = NOW()
WHERE id = @id AND tenant_id = @tenant_id
RETURNING *;

-- name: DeleteColumn :exec
DELETE FROM columns WHERE id = $1 AND tenant_id = $2;

-- name: UpdateColumnPosition :exec
UPDATE columns SET position = $3, updated_at = NOW()
WHERE id = $1 AND tenant_id = $2;

-- name: GetMaxColumnPosition :one
SELECT COALESCE(MAX(position), -1)::int as max_position
FROM columns WHERE board_id = $1 AND tenant_id = $2;
