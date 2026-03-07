package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mer-prog/taskflow/internal/model"
)

type mockBoardRepo struct {
	boards     map[uuid.UUID]BoardData
	columns    map[uuid.UUID][]ColumnData
	tasks      map[uuid.UUID][]BoardTaskData
	taskLabels map[uuid.UUID][]BoardTaskLabelData
}

func newMockBoardRepo() *mockBoardRepo {
	return &mockBoardRepo{
		boards:     make(map[uuid.UUID]BoardData),
		columns:    make(map[uuid.UUID][]ColumnData),
		tasks:      make(map[uuid.UUID][]BoardTaskData),
		taskLabels: make(map[uuid.UUID][]BoardTaskLabelData),
	}
}

func (m *mockBoardRepo) CreateBoard(_ context.Context, tenantID, projectID uuid.UUID, name string) (BoardData, error) {
	id := uuid.New()
	b := BoardData{ID: id, TenantID: tenantID, ProjectID: projectID, Name: name, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	m.boards[id] = b
	return b, nil
}

func (m *mockBoardRepo) GetBoardByID(_ context.Context, id, _ uuid.UUID) (BoardData, error) {
	b, ok := m.boards[id]
	if !ok {
		return BoardData{}, ErrBoardNotFound
	}
	return b, nil
}

func (m *mockBoardRepo) UpdateBoard(_ context.Context, id, _ uuid.UUID, name *string) (BoardData, error) {
	b, ok := m.boards[id]
	if !ok {
		return BoardData{}, ErrBoardNotFound
	}
	if name != nil {
		b.Name = *name
	}
	m.boards[id] = b
	return b, nil
}

func (m *mockBoardRepo) DeleteBoard(_ context.Context, id, _ uuid.UUID) error {
	delete(m.boards, id)
	return nil
}

func (m *mockBoardRepo) GetColumnsByBoardID(_ context.Context, boardID, _ uuid.UUID) ([]ColumnData, error) {
	return m.columns[boardID], nil
}

func (m *mockBoardRepo) GetTasksByBoardID(_ context.Context, boardID, _ uuid.UUID) ([]BoardTaskData, error) {
	return m.tasks[boardID], nil
}

func (m *mockBoardRepo) GetTaskLabelsByBoardID(_ context.Context, boardID, _ uuid.UUID) ([]BoardTaskLabelData, error) {
	return m.taskLabels[boardID], nil
}

func (m *mockBoardRepo) CreateColumn(_ context.Context, _, boardID uuid.UUID, name, color string, wipLimit *int, position int) (ColumnData, error) {
	col := ColumnData{ID: uuid.New(), BoardID: boardID, Name: name, Color: color, WIPLimit: wipLimit, Position: position, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	m.columns[boardID] = append(m.columns[boardID], col)
	return col, nil
}

func (m *mockBoardRepo) UpdateColumn(_ context.Context, id, _ uuid.UUID, name, color *string, wipLimit *int) (ColumnData, error) {
	for bID, cols := range m.columns {
		for i, c := range cols {
			if c.ID == id {
				if name != nil {
					c.Name = *name
				}
				if color != nil {
					c.Color = *color
				}
				c.WIPLimit = wipLimit
				m.columns[bID][i] = c
				return c, nil
			}
		}
	}
	return ColumnData{}, ErrColumnNotFound
}

func (m *mockBoardRepo) DeleteColumn(_ context.Context, id, _ uuid.UUID) error {
	for bID, cols := range m.columns {
		for i, c := range cols {
			if c.ID == id {
				m.columns[bID] = append(cols[:i], cols[i+1:]...)
				return nil
			}
		}
	}
	return nil
}

func (m *mockBoardRepo) UpdateColumnPosition(_ context.Context, id, _ uuid.UUID, position int) error {
	for bID, cols := range m.columns {
		for i, c := range cols {
			if c.ID == id {
				c.Position = position
				m.columns[bID][i] = c
				return nil
			}
		}
	}
	return nil
}

func (m *mockBoardRepo) GetMaxColumnPosition(_ context.Context, _, _ uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockBoardRepo) GetColumnByID(_ context.Context, id, _ uuid.UUID) (ColumnData, error) {
	for _, cols := range m.columns {
		for _, c := range cols {
			if c.ID == id {
				return c, nil
			}
		}
	}
	return ColumnData{}, ErrColumnNotFound
}

func (m *mockBoardRepo) GetBoardsByProjectID(_ context.Context, projectID, _ uuid.UUID) ([]BoardData, error) {
	var result []BoardData
	for _, b := range m.boards {
		if b.ProjectID == projectID {
			result = append(result, b)
		}
	}
	return result, nil
}

func (m *mockBoardRepo) RunInTx(_ context.Context, fn func(BoardRepository) error) error {
	return fn(m)
}

func TestBoardService_Create(t *testing.T) {
	repo := newMockBoardRepo()
	svc := NewBoardService(repo)

	tenantID := uuid.New()
	projectID := uuid.New()

	board, err := svc.Create(context.Background(), tenantID, model.CreateBoardRequest{
		ProjectID: projectID,
		Name:      "Test Board",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if board.Name != "Test Board" {
		t.Errorf("expected name 'Test Board', got %q", board.Name)
	}
	if board.ProjectID != projectID {
		t.Errorf("expected project_id %s, got %s", projectID, board.ProjectID)
	}
}

func TestBoardService_Get(t *testing.T) {
	repo := newMockBoardRepo()
	svc := NewBoardService(repo)

	tenantID := uuid.New()
	projectID := uuid.New()
	boardID := uuid.New()

	repo.boards[boardID] = BoardData{
		ID: boardID, TenantID: tenantID, ProjectID: projectID,
		Name: "My Board", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.columns[boardID] = []ColumnData{
		{ID: uuid.New(), BoardID: boardID, Name: "Todo", Position: 0, Color: "#ccc"},
	}

	resp, err := svc.Get(context.Background(), boardID, tenantID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Name != "My Board" {
		t.Errorf("expected name 'My Board', got %q", resp.Name)
	}
	if len(resp.Columns) != 1 {
		t.Errorf("expected 1 column, got %d", len(resp.Columns))
	}
}

func TestBoardService_Get_NotFound(t *testing.T) {
	repo := newMockBoardRepo()
	svc := NewBoardService(repo)

	_, err := svc.Get(context.Background(), uuid.New(), uuid.New())
	if err != ErrBoardNotFound {
		t.Errorf("expected ErrBoardNotFound, got %v", err)
	}
}

func TestBoardService_ReorderColumns(t *testing.T) {
	repo := newMockBoardRepo()
	svc := NewBoardService(repo)

	tenantID := uuid.New()
	boardID := uuid.New()

	col1 := ColumnData{ID: uuid.New(), BoardID: boardID, Name: "A", Position: 0}
	col2 := ColumnData{ID: uuid.New(), BoardID: boardID, Name: "B", Position: 1}
	repo.columns[boardID] = []ColumnData{col1, col2}

	// Reverse order
	err := svc.ReorderColumns(context.Background(), tenantID, model.ReorderColumnsRequest{
		ColumnIDs: []uuid.UUID{col2.ID, col1.ID},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, c := range repo.columns[boardID] {
		if c.ID == col2.ID && c.Position != 0 {
			t.Errorf("expected col2 at position 0, got %d", c.Position)
		}
		if c.ID == col1.ID && c.Position != 1 {
			t.Errorf("expected col1 at position 1, got %d", c.Position)
		}
	}
}
