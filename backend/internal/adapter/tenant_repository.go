package adapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/mer-prog/taskflow/internal/repository"
	"github.com/mer-prog/taskflow/internal/service"
)

type TenantRepositoryAdapter struct {
	q *repository.Queries
}

func NewTenantRepository(q *repository.Queries) *TenantRepositoryAdapter {
	return &TenantRepositoryAdapter{q: q}
}

func (a *TenantRepositoryAdapter) CreateTenant(ctx context.Context, name, slug string) (service.TenantData, error) {
	t, err := a.q.CreateTenant(ctx, repository.CreateTenantParams{Name: name, Slug: slug})
	if err != nil {
		return service.TenantData{}, err
	}
	return toTenantData(t), nil
}

func (a *TenantRepositoryAdapter) GetTenantByID(ctx context.Context, id uuid.UUID) (service.TenantData, error) {
	t, err := a.q.GetTenantByID(ctx, toPgUUID(id))
	if err != nil {
		return service.TenantData{}, err
	}
	return toTenantData(t), nil
}

func (a *TenantRepositoryAdapter) UpdateTenant(ctx context.Context, id uuid.UUID, name, slug *string) (service.TenantData, error) {
	t, err := a.q.UpdateTenant(ctx, repository.UpdateTenantParams{
		ID:   toPgUUID(id),
		Name: toPgText(name),
		Slug: toPgText(slug),
	})
	if err != nil {
		return service.TenantData{}, err
	}
	return toTenantData(t), nil
}

func (a *TenantRepositoryAdapter) GetTenantsByUserID(ctx context.Context, userID uuid.UUID) ([]service.TenantData, error) {
	tenants, err := a.q.GetTenantsByUserID(ctx, toPgUUID(userID))
	if err != nil {
		return nil, err
	}
	result := make([]service.TenantData, len(tenants))
	for i, t := range tenants {
		result[i] = toTenantData(t)
	}
	return result, nil
}

func (a *TenantRepositoryAdapter) AddMember(ctx context.Context, tenantID, userID uuid.UUID, role string) error {
	return a.q.CreateTenantMember(ctx, repository.CreateTenantMemberParams{
		TenantID: toPgUUID(tenantID),
		UserID:   toPgUUID(userID),
		Role:     role,
	})
}

func (a *TenantRepositoryAdapter) GetMemberRole(ctx context.Context, tenantID, userID uuid.UUID) (string, error) {
	tm, err := a.q.GetTenantMemberByTenantAndUser(ctx, repository.GetTenantMemberByTenantAndUserParams{
		TenantID: toPgUUID(tenantID),
		UserID:   toPgUUID(userID),
	})
	if err != nil {
		return "", err
	}
	return tm.Role, nil
}

func (a *TenantRepositoryAdapter) GetMembers(ctx context.Context, tenantID uuid.UUID) ([]service.TenantMember, error) {
	rows, err := a.q.GetTenantMembers(ctx, toPgUUID(tenantID))
	if err != nil {
		return nil, err
	}
	result := make([]service.TenantMember, len(rows))
	for i, r := range rows {
		result[i] = service.TenantMember{
			UserID:      fromPgUUID(r.UserID),
			Email:       r.Email,
			DisplayName: r.DisplayName,
			AvatarURL:   fromPgText(r.AvatarUrl),
			Role:        r.Role,
			JoinedAt:    r.JoinedAt.Time,
		}
	}
	return result, nil
}

func (a *TenantRepositoryAdapter) UpdateMemberRole(ctx context.Context, tenantID, userID uuid.UUID, role string) error {
	return a.q.UpdateTenantMemberRole(ctx, repository.UpdateTenantMemberRoleParams{
		TenantID: toPgUUID(tenantID),
		UserID:   toPgUUID(userID),
		Role:     role,
	})
}

func (a *TenantRepositoryAdapter) RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	return a.q.RemoveTenantMember(ctx, repository.RemoveTenantMemberParams{
		TenantID: toPgUUID(tenantID),
		UserID:   toPgUUID(userID),
	})
}

func toTenantData(t repository.Tenant) service.TenantData {
	return service.TenantData{
		ID:        fromPgUUID(t.ID),
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: t.CreatedAt.Time,
		UpdatedAt: t.UpdatedAt.Time,
	}
}

func toPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func fromPgText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}
