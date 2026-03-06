package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/mer-prog/taskflow/internal/repository"
	"github.com/mer-prog/taskflow/internal/service"
)

type TaskRepositoryAdapter struct {
	q *repository.Queries
}

func NewTaskRepository(q *repository.Queries) *TaskRepositoryAdapter {
	return &TaskRepositoryAdapter{q: q}
}

func (a *TaskRepositoryAdapter) CreateTask(ctx context.Context, tenantID, columnID uuid.UUID, title string, description *string, position int, priority string, assigneeID *uuid.UUID, dueDate *time.Time) (service.TaskData, error) {
	t, err := a.q.CreateTask(ctx, repository.CreateTaskParams{
		TenantID:    toPgUUID(tenantID),
		ColumnID:    toPgUUID(columnID),
		Title:       title,
		Description: toPgText(description),
		Position:    int32(position),
		Priority:    priority,
		AssigneeID:  toPgUUIDPtr(assigneeID),
		DueDate:     toPgTimestamptz(dueDate),
	})
	if err != nil {
		return service.TaskData{}, err
	}
	return taskToTaskData(t), nil
}

func (a *TaskRepositoryAdapter) GetTaskByID(ctx context.Context, id, tenantID uuid.UUID) (service.TaskData, error) {
	r, err := a.q.GetTaskByID(ctx, repository.GetTaskByIDParams{
		ID:       toPgUUID(id),
		TenantID: toPgUUID(tenantID),
	})
	if err != nil {
		return service.TaskData{}, err
	}
	return service.TaskData{
		ID:             fromPgUUID(r.ID),
		TenantID:       fromPgUUID(r.TenantID),
		ColumnID:       fromPgUUID(r.ColumnID),
		Title:          r.Title,
		Description:    fromPgText(r.Description),
		Position:       int(r.Position),
		Priority:       r.Priority,
		AssigneeID:     fromPgUUIDPtr(r.AssigneeID),
		DueDate:        fromPgTimestamptzPtr(r.DueDate),
		AssigneeName:   fromPgText(r.AssigneeName),
		AssigneeAvatar: fromPgText(r.AssigneeAvatar),
		CreatedAt:      r.CreatedAt.Time,
		UpdatedAt:      r.UpdatedAt.Time,
	}, nil
}

func (a *TaskRepositoryAdapter) UpdateTask(ctx context.Context, id, tenantID uuid.UUID, title, description, priority *string, assigneeID *uuid.UUID, dueDate *time.Time) (service.TaskData, error) {
	t, err := a.q.UpdateTask(ctx, repository.UpdateTaskParams{
		Title:       toPgText(title),
		Description: toPgText(description),
		Priority:    toPgText(priority),
		AssigneeID:  toPgUUIDPtr(assigneeID),
		DueDate:     toPgTimestamptz(dueDate),
		ID:          toPgUUID(id),
		TenantID:    toPgUUID(tenantID),
	})
	if err != nil {
		return service.TaskData{}, err
	}
	return taskToTaskData(t), nil
}

func (a *TaskRepositoryAdapter) DeleteTask(ctx context.Context, id, tenantID uuid.UUID) error {
	return a.q.DeleteTask(ctx, repository.DeleteTaskParams{
		ID:       toPgUUID(id),
		TenantID: toPgUUID(tenantID),
	})
}

func (a *TaskRepositoryAdapter) MoveTask(ctx context.Context, id uuid.UUID, toColumnID uuid.UUID, position int, tenantID uuid.UUID) error {
	return a.q.MoveTask(ctx, repository.MoveTaskParams{
		ID:       toPgUUID(id),
		ColumnID: toPgUUID(toColumnID),
		Position: int32(position),
		TenantID: toPgUUID(tenantID),
	})
}

func (a *TaskRepositoryAdapter) ShiftTaskPositionsUp(ctx context.Context, columnID, tenantID uuid.UUID, fromPosition int) error {
	return a.q.ShiftTaskPositionsUp(ctx, repository.ShiftTaskPositionsUpParams{
		ColumnID: toPgUUID(columnID),
		TenantID: toPgUUID(tenantID),
		Position: int32(fromPosition),
	})
}

func (a *TaskRepositoryAdapter) ShiftTaskPositionsDown(ctx context.Context, columnID, tenantID uuid.UUID, afterPosition int) error {
	return a.q.ShiftTaskPositionsDown(ctx, repository.ShiftTaskPositionsDownParams{
		ColumnID: toPgUUID(columnID),
		TenantID: toPgUUID(tenantID),
		Position: int32(afterPosition),
	})
}

func (a *TaskRepositoryAdapter) GetMaxTaskPosition(ctx context.Context, columnID, tenantID uuid.UUID) (int, error) {
	pos, err := a.q.GetMaxTaskPosition(ctx, repository.GetMaxTaskPositionParams{
		ColumnID: toPgUUID(columnID),
		TenantID: toPgUUID(tenantID),
	})
	if err != nil {
		return 0, err
	}
	return int(pos), nil
}

func (a *TaskRepositoryAdapter) GetTaskLabels(ctx context.Context, taskID uuid.UUID) ([]service.LabelData, error) {
	rows, err := a.q.GetTaskLabels(ctx, toPgUUID(taskID))
	if err != nil {
		return nil, err
	}
	result := make([]service.LabelData, len(rows))
	for i, r := range rows {
		result[i] = service.LabelData{
			ID:    fromPgUUID(r.ID),
			Name:  r.Name,
			Color: r.Color,
		}
	}
	return result, nil
}

func (a *TaskRepositoryAdapter) AddTaskLabel(ctx context.Context, taskID, labelID uuid.UUID) error {
	return a.q.AddTaskLabel(ctx, repository.AddTaskLabelParams{
		TaskID:  toPgUUID(taskID),
		LabelID: toPgUUID(labelID),
	})
}

func (a *TaskRepositoryAdapter) RemoveTaskLabel(ctx context.Context, taskID, labelID uuid.UUID) error {
	return a.q.RemoveTaskLabel(ctx, repository.RemoveTaskLabelParams{
		TaskID:  toPgUUID(taskID),
		LabelID: toPgUUID(labelID),
	})
}

func (a *TaskRepositoryAdapter) CreateLabel(ctx context.Context, tenantID uuid.UUID, name, color string) (service.LabelData, error) {
	l, err := a.q.CreateLabel(ctx, repository.CreateLabelParams{
		TenantID: toPgUUID(tenantID),
		Name:     name,
		Color:    color,
	})
	if err != nil {
		return service.LabelData{}, err
	}
	return service.LabelData{
		ID:    fromPgUUID(l.ID),
		Name:  l.Name,
		Color: l.Color,
	}, nil
}

func (a *TaskRepositoryAdapter) GetLabelsByTenantID(ctx context.Context, tenantID uuid.UUID) ([]service.LabelData, error) {
	labels, err := a.q.GetLabelsByTenantID(ctx, toPgUUID(tenantID))
	if err != nil {
		return nil, err
	}
	result := make([]service.LabelData, len(labels))
	for i, l := range labels {
		result[i] = service.LabelData{
			ID:    fromPgUUID(l.ID),
			Name:  l.Name,
			Color: l.Color,
		}
	}
	return result, nil
}

func (a *TaskRepositoryAdapter) DeleteLabel(ctx context.Context, id, tenantID uuid.UUID) error {
	return a.q.DeleteLabel(ctx, repository.DeleteLabelParams{
		ID:       toPgUUID(id),
		TenantID: toPgUUID(tenantID),
	})
}

func (a *TaskRepositoryAdapter) CreateComment(ctx context.Context, tenantID, taskID, userID uuid.UUID, content string) (service.CommentData, error) {
	tc, err := a.q.CreateComment(ctx, repository.CreateCommentParams{
		TenantID: toPgUUID(tenantID),
		TaskID:   toPgUUID(taskID),
		UserID:   toPgUUID(userID),
		Content:  content,
	})
	if err != nil {
		return service.CommentData{}, err
	}

	user, err := a.q.GetUserByID(ctx, toPgUUID(userID))
	if err != nil {
		return service.CommentData{}, err
	}

	return service.CommentData{
		ID:          fromPgUUID(tc.ID),
		UserID:      fromPgUUID(tc.UserID),
		DisplayName: user.DisplayName,
		AvatarURL:   fromPgText(user.AvatarUrl),
		Content:     tc.Content,
		CreatedAt:   tc.CreatedAt.Time,
		UpdatedAt:   tc.UpdatedAt.Time,
	}, nil
}

func (a *TaskRepositoryAdapter) GetCommentsByTaskID(ctx context.Context, taskID, tenantID uuid.UUID) ([]service.CommentData, error) {
	rows, err := a.q.GetCommentsByTaskID(ctx, repository.GetCommentsByTaskIDParams{
		TaskID:   toPgUUID(taskID),
		TenantID: toPgUUID(tenantID),
	})
	if err != nil {
		return nil, err
	}
	result := make([]service.CommentData, len(rows))
	for i, r := range rows {
		result[i] = service.CommentData{
			ID:          fromPgUUID(r.ID),
			UserID:      fromPgUUID(r.UserID),
			DisplayName: r.DisplayName,
			AvatarURL:   fromPgText(r.AvatarUrl),
			Content:     r.Content,
			CreatedAt:   r.CreatedAt.Time,
			UpdatedAt:   r.UpdatedAt.Time,
		}
	}
	return result, nil
}

// Helper conversions

func toPgUUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return toPgUUID(*id)
}

func taskToTaskData(t repository.Task) service.TaskData {
	return service.TaskData{
		ID:             fromPgUUID(t.ID),
		TenantID:       fromPgUUID(t.TenantID),
		ColumnID:       fromPgUUID(t.ColumnID),
		Title:          t.Title,
		Description:    fromPgText(t.Description),
		Position:       int(t.Position),
		Priority:       t.Priority,
		AssigneeID:     fromPgUUIDPtr(t.AssigneeID),
		DueDate:        fromPgTimestamptzPtr(t.DueDate),
		AssigneeName:   nil,
		AssigneeAvatar: nil,
		CreatedAt:      t.CreatedAt.Time,
		UpdatedAt:      t.UpdatedAt.Time,
	}
}
