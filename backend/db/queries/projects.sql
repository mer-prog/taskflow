-- name: CreateProject :one
INSERT INTO projects (tenant_id, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetProjectByID :one
SELECT * FROM projects WHERE id = $1 AND tenant_id = $2;

-- name: GetProjectsByTenantID :many
SELECT * FROM projects
WHERE tenant_id = $1 AND archived_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateProject :one
UPDATE projects
SET name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    updated_at = NOW()
WHERE id = @id AND tenant_id = @tenant_id
RETURNING *;

-- name: ArchiveProject :exec
UPDATE projects
SET archived_at = NOW(), updated_at = NOW()
WHERE id = $1 AND tenant_id = $2;

-- name: AddProjectMember :exec
INSERT INTO project_members (project_id, user_id, role)
VALUES ($1, $2, $3);

-- name: GetProjectMembers :many
SELECT
    pm.user_id,
    u.email,
    u.display_name,
    u.avatar_url,
    pm.role,
    pm.created_at AS joined_at
FROM project_members pm
JOIN users u ON u.id = pm.user_id
WHERE pm.project_id = $1
ORDER BY pm.created_at;

-- name: RemoveProjectMember :exec
DELETE FROM project_members WHERE project_id = $1 AND user_id = $2;
