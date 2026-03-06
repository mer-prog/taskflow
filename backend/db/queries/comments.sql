-- name: CreateComment :one
INSERT INTO task_comments (tenant_id, task_id, user_id, content)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCommentsByTaskID :many
SELECT tc.id, tc.task_id, tc.user_id, tc.content, tc.created_at, tc.updated_at,
       u.display_name, u.avatar_url
FROM task_comments tc
JOIN users u ON u.id = tc.user_id
WHERE tc.task_id = $1 AND tc.tenant_id = $2
ORDER BY tc.created_at;

-- name: UpdateComment :one
UPDATE task_comments SET content = $2, updated_at = NOW()
WHERE id = $1 AND tenant_id = $3
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM task_comments WHERE id = $1 AND tenant_id = $2;
