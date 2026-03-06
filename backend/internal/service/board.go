package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mer-prog/taskflow/internal/model"
)

type BoardData struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	ProjectID uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ColumnData struct {
	ID        uuid.UUID
	BoardID   uuid.UUID
	Name      string
	Position  int
	Color     string
	WIPLimit  *int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BoardTaskData struct {
	ID             uuid.UUID
	ColumnID       uuid.UUID
	Title          string
	Position       int
	Priority       string
	AssigneeID     *uuid.UUID
	DueDate        *time.Time
	AssigneeName   *string
	AssigneeAvatar *string
}

type BoardTaskLabelData struct {
	TaskID  uuid.UUID
	LabelID uuid.UUID
	Name    string
	Color   string
}

type BoardRepository interface {
	CreateBoard(ctx context.Context, tenantID, projectID uuid.UUID, name string) (BoardData, error)
	GetBoardByID(ctx context.Context, id, tenantID uuid.UUID) (BoardData, error)
	UpdateBoard(ctx context.Context, id, tenantID uuid.UUID, name *string) (BoardData, error)
	DeleteBoard(ctx context.Context, id, tenantID uuid.UUID) error
	GetColumnsByBoardID(ctx context.Context, boardID, tenantID uuid.UUID) ([]ColumnData, error)
	GetTasksByBoardID(ctx context.Context, boardID, tenantID uuid.UUID) ([]BoardTaskData, error)
	GetTaskLabelsByBoardID(ctx context.Context, boardID, tenantID uuid.UUID) ([]BoardTaskLabelData, error)
	CreateColumn(ctx context.Context, tenantID, boardID uuid.UUID, name, color string, wipLimit *int, position int) (ColumnData, error)
	UpdateColumn(ctx context.Context, id, tenantID uuid.UUID, name, color *string, wipLimit *int) (ColumnData, error)
	DeleteColumn(ctx context.Context, id, tenantID uuid.UUID) error
	UpdateColumnPosition(ctx context.Context, id, tenantID uuid.UUID, position int) error
	GetMaxColumnPosition(ctx context.Context, boardID, tenantID uuid.UUID) (int, error)
	GetColumnByID(ctx context.Context, id, tenantID uuid.UUID) (ColumnData, error)
}

type BoardService struct {
	repo BoardRepository
}

func NewBoardService(repo BoardRepository) *BoardService {
	return &BoardService{repo: repo}
}

func (s *BoardService) Create(ctx context.Context, tenantID uuid.UUID, req model.CreateBoardRequest) (*BoardData, error) {
	b, err := s.repo.CreateBoard(ctx, tenantID, req.ProjectID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("service.BoardCreate: %w", err)
	}
	return &b, nil
}

func (s *BoardService) Get(ctx context.Context, id, tenantID uuid.UUID) (*model.BoardDetailResponse, error) {
	board, err := s.repo.GetBoardByID(ctx, id, tenantID)
	if err != nil {
		return nil, ErrBoardNotFound
	}

	columns, err := s.repo.GetColumnsByBoardID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("service.BoardGet: %w", err)
	}

	tasks, err := s.repo.GetTasksByBoardID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("service.BoardGet: %w", err)
	}

	taskLabels, err := s.repo.GetTaskLabelsByBoardID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("service.BoardGet: %w", err)
	}

	// Index labels by task ID
	labelMap := make(map[uuid.UUID][]model.LabelInfo)
	for _, tl := range taskLabels {
		labelMap[tl.TaskID] = append(labelMap[tl.TaskID], model.LabelInfo{
			ID: tl.LabelID, Name: tl.Name, Color: tl.Color,
		})
	}

	// Index tasks by column ID
	taskMap := make(map[uuid.UUID][]model.TaskSummary)
	for _, t := range tasks {
		var assignee *model.AssigneeInfo
		if t.AssigneeID != nil {
			assignee = &model.AssigneeInfo{
				ID: *t.AssigneeID, DisplayName: derefStr(t.AssigneeName), AvatarURL: t.AssigneeAvatar,
			}
		}
		labels := labelMap[t.ID]
		if labels == nil {
			labels = []model.LabelInfo{}
		}
		taskMap[t.ColumnID] = append(taskMap[t.ColumnID], model.TaskSummary{
			ID: t.ID, Title: t.Title, Position: t.Position,
			Priority: t.Priority, DueDate: t.DueDate,
			Assignee: assignee, Labels: labels,
		})
	}

	cols := make([]model.ColumnWithTasks, len(columns))
	for i, c := range columns {
		colTasks := taskMap[c.ID]
		if colTasks == nil {
			colTasks = []model.TaskSummary{}
		}
		cols[i] = model.ColumnWithTasks{
			ID: c.ID, Name: c.Name, Position: c.Position,
			Color: c.Color, WIPLimit: c.WIPLimit, Tasks: colTasks,
		}
	}

	return &model.BoardDetailResponse{
		ID: board.ID, ProjectID: board.ProjectID, Name: board.Name,
		Columns: cols, CreatedAt: board.CreatedAt, UpdatedAt: board.UpdatedAt,
	}, nil
}

func (s *BoardService) Update(ctx context.Context, id, tenantID uuid.UUID, req model.UpdateBoardRequest) (*BoardData, error) {
	b, err := s.repo.UpdateBoard(ctx, id, tenantID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("service.BoardUpdate: %w", err)
	}
	return &b, nil
}

func (s *BoardService) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	return s.repo.DeleteBoard(ctx, id, tenantID)
}

func (s *BoardService) CreateColumn(ctx context.Context, tenantID uuid.UUID, req model.CreateColumnRequest) (*model.ColumnResponse, error) {
	maxPos, err := s.repo.GetMaxColumnPosition(ctx, req.BoardID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("service.CreateColumn: %w", err)
	}

	color := req.Color
	if color == "" {
		color = "#6B7280"
	}

	col, err := s.repo.CreateColumn(ctx, tenantID, req.BoardID, req.Name, color, req.WIPLimit, maxPos+1)
	if err != nil {
		return nil, fmt.Errorf("service.CreateColumn: %w", err)
	}
	return toColumnResponse(col), nil
}

func (s *BoardService) UpdateColumn(ctx context.Context, id, tenantID uuid.UUID, req model.UpdateColumnRequest) (*model.ColumnResponse, error) {
	col, err := s.repo.UpdateColumn(ctx, id, tenantID, req.Name, req.Color, req.WIPLimit)
	if err != nil {
		return nil, fmt.Errorf("service.UpdateColumn: %w", err)
	}
	return toColumnResponse(col), nil
}

func (s *BoardService) DeleteColumn(ctx context.Context, id, tenantID uuid.UUID) error {
	return s.repo.DeleteColumn(ctx, id, tenantID)
}

func (s *BoardService) ReorderColumns(ctx context.Context, tenantID uuid.UUID, req model.ReorderColumnsRequest) error {
	for i, colID := range req.ColumnIDs {
		if err := s.repo.UpdateColumnPosition(ctx, colID, tenantID, i); err != nil {
			return fmt.Errorf("service.ReorderColumns: %w", err)
		}
	}
	return nil
}

func (s *BoardService) GetBoardIDByColumnID(ctx context.Context, columnID, tenantID uuid.UUID) (uuid.UUID, error) {
	col, err := s.repo.GetColumnByID(ctx, columnID, tenantID)
	if err != nil {
		return uuid.Nil, ErrColumnNotFound
	}
	return col.BoardID, nil
}

func toColumnResponse(c ColumnData) *model.ColumnResponse {
	return &model.ColumnResponse{
		ID: c.ID, BoardID: c.BoardID, Name: c.Name,
		Position: c.Position, Color: c.Color, WIPLimit: c.WIPLimit,
		CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
