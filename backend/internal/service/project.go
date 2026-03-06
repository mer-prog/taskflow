package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mer-prog/taskflow/internal/model"
)

type ProjectData struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProjectMember struct {
	UserID      uuid.UUID
	Email       string
	DisplayName string
	AvatarURL   *string
	Role        string
	JoinedAt    time.Time
}

type ProjectRepository interface {
	CreateProject(ctx context.Context, tenantID uuid.UUID, name string, description *string) (ProjectData, error)
	GetProjectByID(ctx context.Context, id, tenantID uuid.UUID) (ProjectData, error)
	GetProjectsByTenantID(ctx context.Context, tenantID uuid.UUID) ([]ProjectData, error)
	UpdateProject(ctx context.Context, id, tenantID uuid.UUID, name, description *string) (ProjectData, error)
	ArchiveProject(ctx context.Context, id, tenantID uuid.UUID) error
	AddMember(ctx context.Context, projectID, userID uuid.UUID, role string) error
	GetMembers(ctx context.Context, projectID uuid.UUID) ([]ProjectMember, error)
	RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error
}

type ProjectService struct {
	repo ProjectRepository
}

func NewProjectService(repo ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) Create(ctx context.Context, tenantID, creatorID uuid.UUID, req model.CreateProjectRequest) (*model.ProjectResponse, error) {
	var desc *string
	if req.Description != "" {
		desc = &req.Description
	}

	p, err := s.repo.CreateProject(ctx, tenantID, req.Name, desc)
	if err != nil {
		return nil, fmt.Errorf("service.ProjectCreate: %w", err)
	}

	if err := s.repo.AddMember(ctx, p.ID, creatorID, "manager"); err != nil {
		return nil, fmt.Errorf("service.ProjectCreate: %w", err)
	}

	return toProjectResponse(p), nil
}

func (s *ProjectService) Get(ctx context.Context, id, tenantID uuid.UUID) (*model.ProjectResponse, error) {
	p, err := s.repo.GetProjectByID(ctx, id, tenantID)
	if err != nil {
		return nil, ErrProjectNotFound
	}
	return toProjectResponse(p), nil
}

func (s *ProjectService) List(ctx context.Context, tenantID uuid.UUID) ([]model.ProjectResponse, error) {
	projects, err := s.repo.GetProjectsByTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("service.ProjectList: %w", err)
	}
	result := make([]model.ProjectResponse, len(projects))
	for i, p := range projects {
		result[i] = *toProjectResponse(p)
	}
	return result, nil
}

func (s *ProjectService) Update(ctx context.Context, id, tenantID uuid.UUID, req model.UpdateProjectRequest) (*model.ProjectResponse, error) {
	p, err := s.repo.UpdateProject(ctx, id, tenantID, req.Name, req.Description)
	if err != nil {
		return nil, fmt.Errorf("service.ProjectUpdate: %w", err)
	}
	return toProjectResponse(p), nil
}

func (s *ProjectService) Archive(ctx context.Context, id, tenantID uuid.UUID) error {
	if err := s.repo.ArchiveProject(ctx, id, tenantID); err != nil {
		return fmt.Errorf("service.ProjectArchive: %w", err)
	}
	return nil
}

func (s *ProjectService) AddMember(ctx context.Context, projectID, userID uuid.UUID, role string) error {
	return s.repo.AddMember(ctx, projectID, userID, role)
}

func (s *ProjectService) GetMembers(ctx context.Context, projectID uuid.UUID) ([]model.MemberResponse, error) {
	members, err := s.repo.GetMembers(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("service.ProjectGetMembers: %w", err)
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

func (s *ProjectService) RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error {
	return s.repo.RemoveMember(ctx, projectID, userID)
}

func toProjectResponse(p ProjectData) *model.ProjectResponse {
	return &model.ProjectResponse{
		ID:          p.ID,
		TenantID:    p.TenantID,
		Name:        p.Name,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
