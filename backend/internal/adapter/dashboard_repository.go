package adapter

import (
	"context"

	"github.com/google/uuid"

	"github.com/mer-prog/taskflow/internal/repository"
	"github.com/mer-prog/taskflow/internal/service"
)

type DashboardRepositoryAdapter struct {
	q *repository.Queries
}

func NewDashboardRepository(q *repository.Queries) *DashboardRepositoryAdapter {
	return &DashboardRepositoryAdapter{q: q}
}

func (a *DashboardRepositoryAdapter) GetTenantTaskSummary(ctx context.Context, tenantID uuid.UUID) (totalTasks, overdueTasks int, err error) {
	row, err := a.q.GetTenantTaskSummary(ctx, toPgUUID(tenantID))
	if err != nil {
		return 0, 0, err
	}
	return int(row.TotalTasks), int(row.OverdueTasks), nil
}

func (a *DashboardRepositoryAdapter) GetTaskCountsByColumn(ctx context.Context, tenantID uuid.UUID) ([]service.ColumnCountData, error) {
	rows, err := a.q.GetTaskCountsByColumn(ctx, toPgUUID(tenantID))
	if err != nil {
		return nil, err
	}
	result := make([]service.ColumnCountData, len(rows))
	for i, r := range rows {
		result[i] = service.ColumnCountData{
			ColumnID:   fromPgUUID(r.ColumnID),
			ColumnName: r.ColumnName,
			BoardName:  r.BoardName,
			TaskCount:  int(r.TaskCount),
		}
	}
	return result, nil
}

func (a *DashboardRepositoryAdapter) GetOverdueTasks(ctx context.Context, tenantID uuid.UUID) ([]service.DashboardTaskData, error) {
	rows, err := a.q.GetOverdueTasks(ctx, toPgUUID(tenantID))
	if err != nil {
		return nil, err
	}
	result := make([]service.DashboardTaskData, len(rows))
	for i, r := range rows {
		result[i] = service.DashboardTaskData{
			ID:          fromPgUUID(r.ID),
			Title:       r.Title,
			Priority:    r.Priority,
			DueDate:     fromPgTimestamptzPtr(r.DueDate),
			ColumnName:  r.ColumnName,
			AssigneeName: fromPgText(r.AssigneeName),
		}
	}
	return result, nil
}

func (a *DashboardRepositoryAdapter) GetTasksByAssignee(ctx context.Context, assigneeID, tenantID uuid.UUID) ([]service.DashboardTaskData, error) {
	rows, err := a.q.GetTasksByAssignee(ctx, repository.GetTasksByAssigneeParams{
		AssigneeID: toPgUUID(assigneeID),
		TenantID:   toPgUUID(tenantID),
	})
	if err != nil {
		return nil, err
	}
	result := make([]service.DashboardTaskData, len(rows))
	for i, r := range rows {
		result[i] = service.DashboardTaskData{
			ID:         fromPgUUID(r.ID),
			Title:      r.Title,
			Priority:   r.Priority,
			DueDate:    fromPgTimestamptzPtr(r.DueDate),
			ColumnName: r.ColumnName,
			BoardName:  r.BoardName,
		}
	}
	return result, nil
}
