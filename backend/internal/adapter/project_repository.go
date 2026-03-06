package adapter

import (
	"context"

	"github.com/google/uuid"

	"github.com/mer-prog/taskflow/internal/repository"
	"github.com/mer-prog/taskflow/internal/service"
)

type ProjectRepositoryAdapter struct {
	q *repository.Queries
}

func NewProjectRepository(q *repository.Queries) *ProjectRepositoryAdapter {
	return &ProjectRepositoryAdapter{q: q}
}

func (a *ProjectRepositoryAdapter) CreateProject(ctx context.Context, tenantID uuid.UUID, name string, description *string) (service.ProjectData, error) {
	p, err := a.q.CreateProject(ctx, repository.CreateProjectParams{
		TenantID:    toPgUUID(tenantID),
		Name:        name,
		Description: toPgText(description),
	})
	if err != nil {
		return service.ProjectData{}, err
	}
	return toProjectData(p), nil
}

func (a *ProjectRepositoryAdapter) GetProjectByID(ctx context.Context, id, tenantID uuid.UUID) (service.ProjectData, error) {
	p, err := a.q.GetProjectByID(ctx, repository.GetProjectByIDParams{
		ID:       toPgUUID(id),
		TenantID: toPgUUID(tenantID),
	})
	if err != nil {
		return service.ProjectData{}, err
	}
	return toProjectData(p), nil
}

func (a *ProjectRepositoryAdapter) GetProjectsByTenantID(ctx context.Context, tenantID uuid.UUID) ([]service.ProjectData, error) {
	projects, err := a.q.GetProjectsByTenantID(ctx, toPgUUID(tenantID))
	if err != nil {
		return nil, err
	}
	result := make([]service.ProjectData, len(projects))
	for i, p := range projects {
		result[i] = toProjectData(p)
	}
	return result, nil
}

func (a *ProjectRepositoryAdapter) UpdateProject(ctx context.Context, id, tenantID uuid.UUID, name, description *string) (service.ProjectData, error) {
	p, err := a.q.UpdateProject(ctx, repository.UpdateProjectParams{
		ID:          toPgUUID(id),
		TenantID:    toPgUUID(tenantID),
		Name:        toPgText(name),
		Description: toPgText(description),
	})
	if err != nil {
		return service.ProjectData{}, err
	}
	return toProjectData(p), nil
}

func (a *ProjectRepositoryAdapter) ArchiveProject(ctx context.Context, id, tenantID uuid.UUID) error {
	return a.q.ArchiveProject(ctx, repository.ArchiveProjectParams{
		ID:       toPgUUID(id),
		TenantID: toPgUUID(tenantID),
	})
}

func (a *ProjectRepositoryAdapter) AddMember(ctx context.Context, projectID, userID uuid.UUID, role string) error {
	return a.q.AddProjectMember(ctx, repository.AddProjectMemberParams{
		ProjectID: toPgUUID(projectID),
		UserID:    toPgUUID(userID),
		Role:      role,
	})
}

func (a *ProjectRepositoryAdapter) GetMembers(ctx context.Context, projectID uuid.UUID) ([]service.ProjectMember, error) {
	rows, err := a.q.GetProjectMembers(ctx, toPgUUID(projectID))
	if err != nil {
		return nil, err
	}
	result := make([]service.ProjectMember, len(rows))
	for i, r := range rows {
		result[i] = service.ProjectMember{
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

func (a *ProjectRepositoryAdapter) RemoveMember(ctx context.Context, projectID, userID uuid.UUID) error {
	return a.q.RemoveProjectMember(ctx, repository.RemoveProjectMemberParams{
		ProjectID: toPgUUID(projectID),
		UserID:    toPgUUID(userID),
	})
}

func toProjectData(p repository.Project) service.ProjectData {
	return service.ProjectData{
		ID:          fromPgUUID(p.ID),
		TenantID:    fromPgUUID(p.TenantID),
		Name:        p.Name,
		Description: fromPgText(p.Description),
		CreatedAt:   p.CreatedAt.Time,
		UpdatedAt:   p.UpdatedAt.Time,
	}
}
