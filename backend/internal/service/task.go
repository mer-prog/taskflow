package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mer-prog/taskflow/internal/model"
)

type TaskData struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	ColumnID       uuid.UUID
	Title          string
	Description    *string
	Position       int
	Priority       string
	AssigneeID     *uuid.UUID
	DueDate        *time.Time
	AssigneeName   *string
	AssigneeAvatar *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type LabelData struct {
	ID    uuid.UUID
	Name  string
	Color string
}

type CommentData struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	DisplayName string
	AvatarURL   *string
	Content     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TaskRepository interface {
	CreateTask(ctx context.Context, tenantID, columnID uuid.UUID, title string, description *string, position int, priority string, assigneeID *uuid.UUID, dueDate *time.Time) (TaskData, error)
	GetTaskByID(ctx context.Context, id, tenantID uuid.UUID) (TaskData, error)
	UpdateTask(ctx context.Context, id, tenantID uuid.UUID, title, description, priority *string, assigneeID *uuid.UUID, dueDate *time.Time) (TaskData, error)
	DeleteTask(ctx context.Context, id, tenantID uuid.UUID) error
	MoveTask(ctx context.Context, id uuid.UUID, toColumnID uuid.UUID, position int, tenantID uuid.UUID) error
	ShiftTaskPositionsUp(ctx context.Context, columnID, tenantID uuid.UUID, fromPosition int) error
	ShiftTaskPositionsDown(ctx context.Context, columnID, tenantID uuid.UUID, afterPosition int) error
	GetMaxTaskPosition(ctx context.Context, columnID, tenantID uuid.UUID) (int, error)
	GetTaskLabels(ctx context.Context, taskID uuid.UUID) ([]LabelData, error)
	AddTaskLabel(ctx context.Context, taskID, labelID uuid.UUID) error
	RemoveTaskLabel(ctx context.Context, taskID, labelID uuid.UUID) error
	CreateLabel(ctx context.Context, tenantID uuid.UUID, name, color string) (LabelData, error)
	GetLabelsByTenantID(ctx context.Context, tenantID uuid.UUID) ([]LabelData, error)
	DeleteLabel(ctx context.Context, id, tenantID uuid.UUID) error
	CreateComment(ctx context.Context, tenantID, taskID, userID uuid.UUID, content string) (CommentData, error)
	GetCommentsByTaskID(ctx context.Context, taskID, tenantID uuid.UUID) ([]CommentData, error)
	RunInTx(ctx context.Context, fn func(TaskRepository) error) error
}

type TaskService struct {
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) Create(ctx context.Context, tenantID uuid.UUID, req model.CreateTaskRequest) (*model.TaskResponse, error) {
	maxPos, err := s.repo.GetMaxTaskPosition(ctx, req.ColumnID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("service.TaskCreate: %w", err)
	}

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	var desc *string
	if req.Description != "" {
		desc = &req.Description
	}

	t, err := s.repo.CreateTask(ctx, tenantID, req.ColumnID, req.Title, desc, maxPos+1, priority, req.AssigneeID, req.DueDate)
	if err != nil {
		return nil, fmt.Errorf("service.TaskCreate: %w", err)
	}

	return s.buildTaskResponse(ctx, t)
}

func (s *TaskService) Get(ctx context.Context, id, tenantID uuid.UUID) (*model.TaskResponse, error) {
	t, err := s.repo.GetTaskByID(ctx, id, tenantID)
	if err != nil {
		return nil, ErrTaskNotFound
	}
	return s.buildTaskResponse(ctx, t)
}

func (s *TaskService) Update(ctx context.Context, id, tenantID uuid.UUID, req model.UpdateTaskRequest) (*model.TaskResponse, error) {
	t, err := s.repo.UpdateTask(ctx, id, tenantID, req.Title, req.Description, req.Priority, req.AssigneeID, req.DueDate)
	if err != nil {
		return nil, fmt.Errorf("service.TaskUpdate: %w", err)
	}
	return s.buildTaskResponse(ctx, t)
}

func (s *TaskService) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	return s.repo.DeleteTask(ctx, id, tenantID)
}

func (s *TaskService) Move(ctx context.Context, tenantID uuid.UUID, req model.MoveTaskRequest) error {
	return s.repo.RunInTx(ctx, func(txRepo TaskRepository) error {
		task, err := txRepo.GetTaskByID(ctx, req.TaskID, tenantID)
		if err != nil {
			return ErrTaskNotFound
		}

		// Close gap in source column
		if err := txRepo.ShiftTaskPositionsDown(ctx, task.ColumnID, tenantID, task.Position); err != nil {
			return fmt.Errorf("service.TaskMove: %w", err)
		}

		// Make room in target column
		if err := txRepo.ShiftTaskPositionsUp(ctx, req.ToColumnID, tenantID, req.NewPosition); err != nil {
			return fmt.Errorf("service.TaskMove: %w", err)
		}

		// Move the task
		if err := txRepo.MoveTask(ctx, req.TaskID, req.ToColumnID, req.NewPosition, tenantID); err != nil {
			return fmt.Errorf("service.TaskMove: %w", err)
		}

		return nil
	})
}

func (s *TaskService) AddLabel(ctx context.Context, taskID, labelID uuid.UUID) error {
	return s.repo.AddTaskLabel(ctx, taskID, labelID)
}

func (s *TaskService) RemoveLabel(ctx context.Context, taskID, labelID uuid.UUID) error {
	return s.repo.RemoveTaskLabel(ctx, taskID, labelID)
}

func (s *TaskService) CreateLabel(ctx context.Context, tenantID uuid.UUID, req model.CreateLabelRequest) (*model.LabelResponse, error) {
	l, err := s.repo.CreateLabel(ctx, tenantID, req.Name, req.Color)
	if err != nil {
		return nil, fmt.Errorf("service.CreateLabel: %w", err)
	}
	return &model.LabelResponse{ID: l.ID, Name: l.Name, Color: l.Color}, nil
}

func (s *TaskService) ListLabels(ctx context.Context, tenantID uuid.UUID) ([]model.LabelResponse, error) {
	labels, err := s.repo.GetLabelsByTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("service.ListLabels: %w", err)
	}
	result := make([]model.LabelResponse, len(labels))
	for i, l := range labels {
		result[i] = model.LabelResponse{ID: l.ID, Name: l.Name, Color: l.Color}
	}
	return result, nil
}

func (s *TaskService) DeleteLabel(ctx context.Context, id, tenantID uuid.UUID) error {
	return s.repo.DeleteLabel(ctx, id, tenantID)
}

func (s *TaskService) CreateComment(ctx context.Context, tenantID, taskID, userID uuid.UUID, req model.CreateCommentRequest) (*model.CommentResponse, error) {
	c, err := s.repo.CreateComment(ctx, tenantID, taskID, userID, req.Content)
	if err != nil {
		return nil, fmt.Errorf("service.CreateComment: %w", err)
	}
	return toCommentResponse(c), nil
}

func (s *TaskService) GetComments(ctx context.Context, taskID, tenantID uuid.UUID) ([]model.CommentResponse, error) {
	comments, err := s.repo.GetCommentsByTaskID(ctx, taskID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("service.GetComments: %w", err)
	}
	result := make([]model.CommentResponse, len(comments))
	for i, c := range comments {
		result[i] = *toCommentResponse(c)
	}
	return result, nil
}

func (s *TaskService) buildTaskResponse(ctx context.Context, t TaskData) (*model.TaskResponse, error) {
	labels, _ := s.repo.GetTaskLabels(ctx, t.ID)
	labelInfos := make([]model.LabelInfo, len(labels))
	for i, l := range labels {
		labelInfos[i] = model.LabelInfo{ID: l.ID, Name: l.Name, Color: l.Color}
	}

	var assignee *model.AssigneeInfo
	if t.AssigneeID != nil {
		assignee = &model.AssigneeInfo{
			ID: *t.AssigneeID, DisplayName: derefStr(t.AssigneeName), AvatarURL: t.AssigneeAvatar,
		}
	}

	return &model.TaskResponse{
		ID: t.ID, ColumnID: t.ColumnID, Title: t.Title,
		Description: t.Description, Position: t.Position,
		Priority: t.Priority, Assignee: assignee, DueDate: t.DueDate,
		Labels: labelInfos, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}, nil
}

func toCommentResponse(c CommentData) *model.CommentResponse {
	return &model.CommentResponse{
		ID: c.ID, UserID: c.UserID, DisplayName: c.DisplayName,
		AvatarURL: c.AvatarURL, Content: c.Content,
		CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}
