-- name: CreateTenant :one
INSERT INTO tenants (name, slug)
VALUES ($1, $2)
RETURNING *;

-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1;

-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1;

-- name: UpdateTenant :one
UPDATE tenants
SET name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    updated_at = NOW()
WHERE id = @id
RETURNING *;

-- name: GetTenantsByUserID :many
SELECT t.*
FROM tenants t
JOIN tenant_members tm ON tm.tenant_id = t.id
WHERE tm.user_id = $1
ORDER BY t.created_at;

-- name: CreateTenantMember :exec
INSERT INTO tenant_members (tenant_id, user_id, role)
VALUES ($1, $2, $3);

-- name: GetTenantMemberByUserID :one
SELECT * FROM tenant_members WHERE user_id = $1 LIMIT 1;

-- name: GetTenantMemberByTenantAndUser :one
SELECT * FROM tenant_members WHERE tenant_id = $1 AND user_id = $2;

-- name: GetTenantMembers :many
SELECT
    tm.user_id,
    u.email,
    u.display_name,
    u.avatar_url,
    tm.role,
    tm.created_at AS joined_at
FROM tenant_members tm
JOIN users u ON u.id = tm.user_id
WHERE tm.tenant_id = $1
ORDER BY tm.created_at;

-- name: UpdateTenantMemberRole :exec
UPDATE tenant_members SET role = $3
WHERE tenant_id = $1 AND user_id = $2;

-- name: RemoveTenantMember :exec
DELETE FROM tenant_members WHERE tenant_id = $1 AND user_id = $2;
