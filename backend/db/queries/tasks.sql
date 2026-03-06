-- name: CreateTask :one
INSERT INTO tasks (tenant_id, column_id, title, description, position, priority, assignee_id, due_date)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetTaskByID :one
SELECT t.id, t.tenant_id, t.column_id, t.title, t.description,
       t.position, t.priority, t.assignee_id, t.due_date,
       t.created_at, t.updated_at,
       u.display_name as assignee_name, u.avatar_url as assignee_avatar
FROM tasks t
LEFT JOIN users u ON u.id = t.assignee_id
WHERE t.id = $1 AND t.tenant_id = $2;

-- name: GetTasksByColumnID :many
SELECT * FROM tasks WHERE column_id = $1 AND tenant_id = $2 ORDER BY position;

-- name: GetTasksByBoardID :many
SELECT t.id, t.column_id, t.title, t.position, t.priority,
       t.assignee_id, t.due_date,
       u.display_name as assignee_name, u.avatar_url as assignee_avatar
FROM tasks t
LEFT JOIN users u ON u.id = t.assignee_id
JOIN columns c ON c.id = t.column_id
WHERE c.board_id = $1 AND t.tenant_id = $2
ORDER BY t.column_id, t.position;

-- name: UpdateTask :one
UPDATE tasks
SET title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    priority = COALESCE(sqlc.narg('priority'), priority),
    assignee_id = COALESCE(sqlc.narg('assignee_id'), assignee_id),
    due_date = COALESCE(sqlc.narg('due_date'), due_date),
    updated_at = NOW()
WHERE id = @id AND tenant_id = @tenant_id
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = $1 AND tenant_id = $2;

-- name: MoveTask :exec
UPDATE tasks SET column_id = $2, position = $3, updated_at = NOW()
WHERE id = $1 AND tenant_id = $4;

-- name: ShiftTaskPositionsUp :exec
UPDATE tasks SET position = position + 1, updated_at = NOW()
WHERE column_id = $1 AND tenant_id = $2 AND position >= $3;

-- name: ShiftTaskPositionsDown :exec
UPDATE tasks SET position = position - 1, updated_at = NOW()
WHERE column_id = $1 AND tenant_id = $2 AND position > $3;

-- name: GetMaxTaskPosition :one
SELECT COALESCE(MAX(position), -1)::int as max_position
FROM tasks WHERE column_id = $1 AND tenant_id = $2;

-- name: GetTasksByAssignee :many
SELECT t.id, t.title, t.priority, t.due_date,
       c.name as column_name, b.name as board_name
FROM tasks t
JOIN columns c ON c.id = t.column_id
JOIN boards b ON b.id = c.board_id
WHERE t.assignee_id = $1 AND t.tenant_id = $2
ORDER BY t.due_date NULLS LAST, t.created_at;

-- name: GetOverdueTasks :many
SELECT t.id, t.title, t.priority, t.due_date,
       c.name as column_name,
       u.display_name as assignee_name
FROM tasks t
JOIN columns c ON c.id = t.column_id
LEFT JOIN users u ON u.id = t.assignee_id
WHERE t.tenant_id = $1 AND t.due_date < NOW()
ORDER BY t.due_date;

-- name: GetTenantTaskSummary :one
SELECT
    COUNT(*)::int as total_tasks,
    COUNT(*) FILTER (WHERE t.due_date < NOW())::int as overdue_tasks
FROM tasks t WHERE t.tenant_id = $1;

-- name: GetTaskCountsByColumn :many
SELECT c.id as column_id, c.name as column_name,
       b.name as board_name, COUNT(t.id)::int as task_count
FROM columns c
JOIN boards b ON b.id = c.board_id
LEFT JOIN tasks t ON t.column_id = c.id
WHERE c.tenant_id = $1
GROUP BY c.id, c.name, b.name, c.position
ORDER BY b.name, c.position;
