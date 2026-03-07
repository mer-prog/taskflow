package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mer-prog/taskflow/internal/model"
)

// mockTaskRepo implements TaskRepository for testing.
type mockTaskRepo struct {
	tasks    map[uuid.UUID]TaskData
	labels   map[uuid.UUID][]LabelData
	moveErr  error
	shiftErr error
	txFn     func(TaskRepository) error
}

func newMockTaskRepo() *mockTaskRepo {
	return &mockTaskRepo{
		tasks:  make(map[uuid.UUID]TaskData),
		labels: make(map[uuid.UUID][]LabelData),
	}
}

func (m *mockTaskRepo) CreateTask(_ context.Context, tenantID, columnID uuid.UUID, title string, description *string, position int, priority string, assigneeID *uuid.UUID, dueDate *time.Time) (TaskData, error) {
	id := uuid.New()
	t := TaskData{
		ID: id, TenantID: tenantID, ColumnID: columnID,
		Title: title, Description: description,
		Position: position, Priority: priority,
		AssigneeID: assigneeID, DueDate: dueDate,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	m.tasks[id] = t
	return t, nil
}

func (m *mockTaskRepo) GetTaskByID(_ context.Context, id, _ uuid.UUID) (TaskData, error) {
	t, ok := m.tasks[id]
	if !ok {
		return TaskData{}, ErrTaskNotFound
	}
	return t, nil
}

func (m *mockTaskRepo) UpdateTask(_ context.Context, id, _ uuid.UUID, title, description, priority *string, _ *uuid.UUID, _ *time.Time) (TaskData, error) {
	t, ok := m.tasks[id]
	if !ok {
		return TaskData{}, ErrTaskNotFound
	}
	if title != nil {
		t.Title = *title
	}
	if description != nil {
		t.Description = description
	}
	if priority != nil {
		t.Priority = *priority
	}
	t.UpdatedAt = time.Now()
	m.tasks[id] = t
	return t, nil
}

func (m *mockTaskRepo) DeleteTask(_ context.Context, id, _ uuid.UUID) error {
	delete(m.tasks, id)
	return nil
}

func (m *mockTaskRepo) MoveTask(_ context.Context, id uuid.UUID, toColumnID uuid.UUID, position int, _ uuid.UUID) error {
	if m.moveErr != nil {
		return m.moveErr
	}
	t := m.tasks[id]
	t.ColumnID = toColumnID
	t.Position = position
	m.tasks[id] = t
	return nil
}

func (m *mockTaskRepo) ShiftTaskPositionsUp(_ context.Context, _, _ uuid.UUID, _ int) error {
	return m.shiftErr
}

func (m *mockTaskRepo) ShiftTaskPositionsDown(_ context.Context, _, _ uuid.UUID, _ int) error {
	return m.shiftErr
}

func (m *mockTaskRepo) GetMaxTaskPosition(_ context.Context, _, _ uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockTaskRepo) GetTaskLabels(_ context.Context, taskID uuid.UUID) ([]LabelData, error) {
	return m.labels[taskID], nil
}

func (m *mockTaskRepo) AddTaskLabel(_ context.Context, taskID, labelID uuid.UUID) error {
	m.labels[taskID] = append(m.labels[taskID], LabelData{ID: labelID})
	return nil
}

func (m *mockTaskRepo) RemoveTaskLabel(_ context.Context, taskID, labelID uuid.UUID) error {
	labels := m.labels[taskID]
	for i, l := range labels {
		if l.ID == labelID {
			m.labels[taskID] = append(labels[:i], labels[i+1:]...)
			break
		}
	}
	return nil
}

func (m *mockTaskRepo) CreateLabel(_ context.Context, _ uuid.UUID, name, color string) (LabelData, error) {
	return LabelData{ID: uuid.New(), Name: name, Color: color}, nil
}

func (m *mockTaskRepo) GetLabelsByTenantID(_ context.Context, _ uuid.UUID) ([]LabelData, error) {
	return nil, nil
}

func (m *mockTaskRepo) DeleteLabel(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (m *mockTaskRepo) CreateComment(_ context.Context, _, _, _ uuid.UUID, content string) (CommentData, error) {
	return CommentData{ID: uuid.New(), Content: content, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}

func (m *mockTaskRepo) GetCommentsByTaskID(_ context.Context, _, _ uuid.UUID) ([]CommentData, error) {
	return nil, nil
}

func (m *mockTaskRepo) RunInTx(_ context.Context, fn func(TaskRepository) error) error {
	return fn(m)
}

func TestTaskService_Create(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)

	tenantID := uuid.New()
	columnID := uuid.New()

	resp, err := svc.Create(context.Background(), tenantID, model.CreateTaskRequest{
		ColumnID: columnID,
		Title:    "Test task",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Title != "Test task" {
		t.Errorf("expected title 'Test task', got %q", resp.Title)
	}
	if resp.Priority != "medium" {
		t.Errorf("expected default priority 'medium', got %q", resp.Priority)
	}
	if resp.ColumnID != columnID {
		t.Errorf("expected column_id %s, got %s", columnID, resp.ColumnID)
	}
}

func TestTaskService_Move(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)

	tenantID := uuid.New()
	fromCol := uuid.New()
	toCol := uuid.New()

	// Create a task
	taskData := TaskData{
		ID: uuid.New(), TenantID: tenantID, ColumnID: fromCol,
		Title: "Move me", Position: 0, Priority: "medium",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.tasks[taskData.ID] = taskData

	err := svc.Move(context.Background(), tenantID, model.MoveTaskRequest{
		TaskID:      taskData.ID,
		ToColumnID:  toCol,
		NewPosition: 0,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	moved := repo.tasks[taskData.ID]
	if moved.ColumnID != toCol {
		t.Errorf("expected column %s, got %s", toCol, moved.ColumnID)
	}
}

func TestTaskService_Move_NotFound(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)

	err := svc.Move(context.Background(), uuid.New(), model.MoveTaskRequest{
		TaskID:      uuid.New(),
		ToColumnID:  uuid.New(),
		NewPosition: 0,
	})

	if err != ErrTaskNotFound {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskService_Delete(t *testing.T) {
	repo := newMockTaskRepo()
	svc := NewTaskService(repo)

	tenantID := uuid.New()
	taskID := uuid.New()
	repo.tasks[taskID] = TaskData{ID: taskID, TenantID: tenantID}

	err := svc.Delete(context.Background(), taskID, tenantID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := repo.tasks[taskID]; ok {
		t.Error("task should have been deleted")
	}
}
