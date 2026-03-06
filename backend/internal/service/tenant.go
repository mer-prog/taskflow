package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mer-prog/taskflow/internal/model"
)

type TenantMember struct {
	UserID      uuid.UUID
	Email       string
	DisplayName string
	AvatarURL   *string
	Role        string
	JoinedAt    time.Time
}

type TenantData struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TenantRepository interface {
	CreateTenant(ctx context.Context, name, slug string) (TenantData, error)
	GetTenantByID(ctx context.Context, id uuid.UUID) (TenantData, error)
	UpdateTenant(ctx context.Context, id uuid.UUID, name, slug *string) (TenantData, error)
	GetTenantsByUserID(ctx context.Context, userID uuid.UUID) ([]TenantData, error)
	AddMember(ctx context.Context, tenantID, userID uuid.UUID, role string) error
	GetMemberRole(ctx context.Context, tenantID, userID uuid.UUID) (string, error)
	GetMembers(ctx context.Context, tenantID uuid.UUID) ([]TenantMember, error)
	UpdateMemberRole(ctx context.Context, tenantID, userID uuid.UUID, role string) error
	RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error
}

type TenantService struct {
	repo TenantRepository
}

func NewTenantService(repo TenantRepository) *TenantService {
	return &TenantService{repo: repo}
}

func (s *TenantService) Create(ctx context.Context, req model.CreateTenantRequest, creatorID uuid.UUID) (*model.TenantResponse, error) {
	t, err := s.repo.CreateTenant(ctx, req.Name, req.Slug)
	if err != nil {
		return nil, fmt.Errorf("service.TenantCreate: %w", err)
	}

	if err := s.repo.AddMember(ctx, t.ID, creatorID, "owner"); err != nil {
		return nil, fmt.Errorf("service.TenantCreate: %w", err)
	}

	return toTenantResponse(t), nil
}

func (s *TenantService) Get(ctx context.Context, id uuid.UUID) (*model.TenantResponse, error) {
	t, err := s.repo.GetTenantByID(ctx, id)
	if err != nil {
		return nil, ErrTenantNotFound
	}
	return toTenantResponse(t), nil
}

func (s *TenantService) Update(ctx context.Context, id uuid.UUID, req model.UpdateTenantRequest) (*model.TenantResponse, error) {
	t, err := s.repo.UpdateTenant(ctx, id, req.Name, req.Slug)
	if err != nil {
		return nil, fmt.Errorf("service.TenantUpdate: %w", err)
	}
	return toTenantResponse(t), nil
}

func (s *TenantService) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.TenantResponse, error) {
	tenants, err := s.repo.GetTenantsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service.TenantListByUser: %w", err)
	}
	result := make([]model.TenantResponse, len(tenants))
	for i, t := range tenants {
		result[i] = *toTenantResponse(t)
	}
	return result, nil
}

func (s *TenantService) CheckMembership(ctx context.Context, tenantID, userID uuid.UUID) (string, error) {
	role, err := s.repo.GetMemberRole(ctx, tenantID, userID)
	if err != nil {
		return "", ErrNotTenantMember
	}
	return role, nil
}

func (s *TenantService) AddMember(ctx context.Context, tenantID, userID uuid.UUID, role string) error {
	if err := s.repo.AddMember(ctx, tenantID, userID, role); err != nil {
		return fmt.Errorf("service.TenantAddMember: %w", err)
	}
	return nil
}

func (s *TenantService) GetMembers(ctx context.Context, tenantID uuid.UUID) ([]model.MemberResponse, error) {
	members, err := s.repo.GetMembers(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("service.TenantGetMembers: %w", err)
	}
	result := make([]model.MemberResponse, len(members))
	for i, m := range members {
		result[i] = model.MemberResponse{
			UserID:      m.UserID,
			Email:       m.Email,
			DisplayName: m.DisplayName,
			AvatarURL:   m.AvatarURL,
			Role:        m.Role,
			JoinedAt:    m.JoinedAt,
		}
	}
	return result, nil
}

func (s *TenantService) UpdateMemberRole(ctx context.Context, tenantID, userID uuid.UUID, role string) error {
	targetRole, err := s.repo.GetMemberRole(ctx, tenantID, userID)
	if err != nil {
		return ErrNotTenantMember
	}
	if targetRole == "owner" {
		return ErrCannotRemoveOwner
	}
	return s.repo.UpdateMemberRole(ctx, tenantID, userID, role)
}

func (s *TenantService) RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	role, err := s.repo.GetMemberRole(ctx, tenantID, userID)
	if err != nil {
		return ErrNotTenantMember
	}
	if role == "owner" {
		return ErrCannotRemoveOwner
	}
	return s.repo.RemoveMember(ctx, tenantID, userID)
}

func toTenantResponse(t TenantData) *model.TenantResponse {
	return &model.TenantResponse{
		ID:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
