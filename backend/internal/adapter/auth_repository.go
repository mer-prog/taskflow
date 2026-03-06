package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/mer-prog/taskflow/internal/repository"
	"github.com/mer-prog/taskflow/internal/service"
)

type AuthRepositoryAdapter struct {
	q *repository.Queries
}

func NewAuthRepository(q *repository.Queries) *AuthRepositoryAdapter {
	return &AuthRepositoryAdapter{q: q}
}

func (a *AuthRepositoryAdapter) CreateUser(ctx context.Context, email, passwordHash, displayName string) (service.RepositoryUser, error) {
	u, err := a.q.CreateUser(ctx, repository.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
	})
	if err != nil {
		return service.RepositoryUser{}, err
	}
	return toServiceUser(u), nil
}

func (a *AuthRepositoryAdapter) GetUserByEmail(ctx context.Context, email string) (service.RepositoryUser, error) {
	u, err := a.q.GetUserByEmail(ctx, email)
	if err != nil {
		return service.RepositoryUser{}, err
	}
	return toServiceUser(u), nil
}

func (a *AuthRepositoryAdapter) GetUserByID(ctx context.Context, id uuid.UUID) (service.RepositoryUser, error) {
	u, err := a.q.GetUserByID(ctx, toPgUUID(id))
	if err != nil {
		return service.RepositoryUser{}, err
	}
	return toServiceUser(u), nil
}

func (a *AuthRepositoryAdapter) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	return a.q.UserExistsByEmail(ctx, email)
}

func (a *AuthRepositoryAdapter) CreateRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	return a.q.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
		UserID:    toPgUUID(userID),
		TokenHash: tokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
}

func (a *AuthRepositoryAdapter) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (service.RepositoryRefreshToken, error) {
	rt, err := a.q.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return service.RepositoryRefreshToken{}, err
	}
	return service.RepositoryRefreshToken{
		ID:        fromPgUUID(rt.ID),
		UserID:    fromPgUUID(rt.UserID),
		TokenHash: rt.TokenHash,
		ExpiresAt: rt.ExpiresAt.Time,
	}, nil
}

func (a *AuthRepositoryAdapter) DeleteRefreshToken(ctx context.Context, id uuid.UUID) error {
	return a.q.DeleteRefreshToken(ctx, toPgUUID(id))
}

func (a *AuthRepositoryAdapter) DeleteRefreshTokensByUser(ctx context.Context, userID uuid.UUID) error {
	return a.q.DeleteRefreshTokensByUser(ctx, toPgUUID(userID))
}

func (a *AuthRepositoryAdapter) CreateTenant(ctx context.Context, name, slug string) (uuid.UUID, error) {
	t, err := a.q.CreateTenant(ctx, repository.CreateTenantParams{
		Name: name,
		Slug: slug,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return fromPgUUID(t.ID), nil
}

func (a *AuthRepositoryAdapter) CreateTenantMember(ctx context.Context, tenantID, userID uuid.UUID, role string) error {
	return a.q.CreateTenantMember(ctx, repository.CreateTenantMemberParams{
		TenantID: toPgUUID(tenantID),
		UserID:   toPgUUID(userID),
		Role:     role,
	})
}

func (a *AuthRepositoryAdapter) GetTenantMemberByUserID(ctx context.Context, userID uuid.UUID) (service.RepositoryTenantMember, error) {
	tm, err := a.q.GetTenantMemberByUserID(ctx, toPgUUID(userID))
	if err != nil {
		return service.RepositoryTenantMember{}, err
	}
	return service.RepositoryTenantMember{
		TenantID: fromPgUUID(tm.TenantID),
		UserID:   fromPgUUID(tm.UserID),
		Role:     tm.Role,
	}, nil
}

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func fromPgUUID(id pgtype.UUID) uuid.UUID {
	return uuid.UUID(id.Bytes)
}

func toServiceUser(u repository.User) service.RepositoryUser {
	var avatarURL *string
	if u.AvatarUrl.Valid {
		avatarURL = &u.AvatarUrl.String
	}
	return service.RepositoryUser{
		ID:           fromPgUUID(u.ID),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		DisplayName:  u.DisplayName,
		AvatarURL:    avatarURL,
		CreatedAt:    u.CreatedAt.Time,
	}
}
