-- name: CreateLabel :one
INSERT INTO labels (tenant_id, name, color)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetLabelsByTenantID :many
SELECT * FROM labels WHERE tenant_id = $1 ORDER BY name;

-- name: DeleteLabel :exec
DELETE FROM labels WHERE id = $1 AND tenant_id = $2;

-- name: AddTaskLabel :exec
INSERT INTO task_labels (task_id, label_id) VALUES ($1, $2);

-- name: RemoveTaskLabel :exec
DELETE FROM task_labels WHERE task_id = $1 AND label_id = $2;

-- name: GetTaskLabels :many
SELECT l.id, l.name, l.color
FROM labels l
JOIN task_labels tl ON tl.label_id = l.id
WHERE tl.task_id = $1
ORDER BY l.name;

-- name: GetTaskLabelsByBoardID :many
SELECT tl.task_id, l.id as label_id, l.name, l.color
FROM task_labels tl
JOIN labels l ON l.id = tl.label_id
JOIN tasks t ON t.id = tl.task_id
JOIN columns c ON c.id = t.column_id
WHERE c.board_id = $1 AND t.tenant_id = $2;
