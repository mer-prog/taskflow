package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mer-prog/taskflow/internal/model"
)

type DashboardTaskData struct {
	ID           uuid.UUID
	Title        string
	Priority     string
	DueDate      *time.Time
	ColumnName   string
	BoardName    string
	AssigneeName *string
}

type ColumnCountData struct {
	ColumnID   uuid.UUID
	ColumnName string
	BoardName  string
	TaskCount  int
}

type DashboardRepository interface {
	GetTenantTaskSummary(ctx context.Context, tenantID uuid.UUID) (totalTasks, overdueTasks int, err error)
	GetTaskCountsByColumn(ctx context.Context, tenantID uuid.UUID) ([]ColumnCountData, error)
	GetOverdueTasks(ctx context.Context, tenantID uuid.UUID) ([]DashboardTaskData, error)
	GetTasksByAssignee(ctx context.Context, assigneeID, tenantID uuid.UUID) ([]DashboardTaskData, error)
}

type DashboardService struct {
	repo DashboardRepository
}

func NewDashboardService(repo DashboardRepository) *DashboardService {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) GetSummary(ctx context.Context, tenantID uuid.UUID) (*model.DashboardSummary, error) {
	total, overdue, err := s.repo.GetTenantTaskSummary(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("service.GetSummary: %w", err)
	}

	byCols, err := s.repo.GetTaskCountsByColumn(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("service.GetSummary: %w", err)
	}

	cols := make([]model.ColumnCount, len(byCols))
	for i, c := range byCols {
		cols[i] = model.ColumnCount{
			ColumnID: c.ColumnID, ColumnName: c.ColumnName,
			BoardName: c.BoardName, TaskCount: c.TaskCount,
		}
	}

	return &model.DashboardSummary{
		TotalTasks: total, OverdueTasks: overdue, ByColumn: cols,
	}, nil
}

func (s *DashboardService) GetOverdueTasks(ctx context.Context, tenantID uuid.UUID) ([]model.DashboardTask, error) {
	tasks, err := s.repo.GetOverdueTasks(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("service.GetOverdueTasks: %w", err)
	}
	return toDashboardTasks(tasks), nil
}

func (s *DashboardService) GetMyTasks(ctx context.Context, userID, tenantID uuid.UUID) ([]model.DashboardTask, error) {
	tasks, err := s.repo.GetTasksByAssignee(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("service.GetMyTasks: %w", err)
	}
	return toDashboardTasks(tasks), nil
}

func toDashboardTasks(tasks []DashboardTaskData) []model.DashboardTask {
	result := make([]model.DashboardTask, len(tasks))
	for i, t := range tasks {
		result[i] = model.DashboardTask{
			ID: t.ID, Title: t.Title, Priority: t.Priority,
			DueDate: t.DueDate, ColumnName: t.ColumnName,
			BoardName: t.BoardName, AssigneeName: t.AssigneeName,
		}
	}
	return result
}
