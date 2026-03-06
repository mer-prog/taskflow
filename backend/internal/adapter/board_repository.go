package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/mer-prog/taskflow/internal/repository"
	"github.com/mer-prog/taskflow/internal/service"
)

type BoardRepositoryAdapter struct {
	q *repository.Queries
}

func NewBoardRepository(q *repository.Queries) *BoardRepositoryAdapter {
	return &BoardRepositoryAdapter{q: q}
}

func (a *BoardRepositoryAdapter) CreateBoard(ctx context.Context, tenantID, projectID uuid.UUID, name string) (service.BoardData, error) {
	b, err := a.q.CreateBoard(ctx, repository.CreateBoardParams{
		TenantID:  toPgUUID(tenantID),
		ProjectID: toPgUUID(projectID),
		Name:      name,
	})
	if err != nil {
		return service.BoardData{}, err
	}
	return toBoardData(b), nil
}

func (a *BoardRepositoryAdapter) GetBoardByID(ctx context.Context, id, tenantID uuid.UUID) (service.BoardData, error) {
	b, err := a.q.GetBoardByID(ctx, repository.GetBoardByIDParams{
		ID:       toPgUUID(id),
		TenantID: toPgUUID(tenantID),
	})
	if err != nil {
		return service.BoardData{}, err
	}
	return toBoardData(b), nil
}

func (a *BoardRepositoryAdapter) UpdateBoard(ctx context.Context, id, tenantID uuid.UUID, name *string) (service.BoardData, error) {
	b, err := a.q.UpdateBoard(ctx, repository.UpdateBoardParams{
		Name:     toPgText(name),
		ID:       toPgUUID(id),
		TenantID: toPgUUID(tenantID),
	})
	if err != nil {
		return service.BoardData{}, err
	}
	return toBoardData(b), nil
}

func (a *BoardRepositoryAdapter) DeleteBoard(ctx context.Context, id, tenantID uuid.UUID) error {
	return a.q.DeleteBoard(ctx, repository.DeleteBoardParams{
		ID:       toPgUUID(id),
		TenantID: toPgUUID(tenantID),
	})
}

func (a *BoardRepositoryAdapter) GetColumnsByBoardID(ctx context.Context, boardID, tenantID uuid.UUID) ([]service.ColumnData, error) {
	cols, err := a.q.GetColumnsByBoardID(ctx, repository.GetColumnsByBoardIDParams{
		BoardID:  toPgUUID(boardID),
		TenantID: toPgUUID(tenantID),
	})
	if err != nil {
		return nil, err
	}
	result := make([]service.ColumnData, len(cols))
	for i, c := range cols {
		result[i] = toColumnData(c)
	}
	return result, nil
}

func (a *BoardRepositoryAdapter) GetTasksByBoardID(ctx context.Context, boardID, tenantID uuid.UUID) ([]service.BoardTaskData, error) {
	rows, err := a.q.GetTasksByBoardID(ctx, repository.GetTasksByBoardIDParams{
		BoardID:  toPgUUID(boardID),
		TenantID: toPgUUID(tenantID),
	})
	if err != nil {
		return nil, err
	}
	result := make([]service.BoardTaskData, len(rows))
	for i, r := range rows {
		result[i] = service.BoardTaskData{
			ID:             fromPgUUID(r.ID),
			ColumnID:       fromPgUUID(r.ColumnID),
			Title:          r.Title,
			Position:       int(r.Position),
			Priority:       r.Priority,
			AssigneeID:     fromPgUUIDPtr(r.AssigneeID),
			DueDate:        fromPgTimestamptzPtr(r.DueDate),
			AssigneeName:   fromPgText(r.AssigneeName),
			AssigneeAvatar: fromPgText(r.AssigneeAvatar),
		}
	}
	return result, nil
}

func (a *BoardRepositoryAdapter) GetTaskLabelsByBoardID(ctx context.Context, boardID, tenantID uuid.UUID) ([]service.BoardTaskLabelData, error) {
	rows, err := a.q.GetTaskLabelsByBoardID(ctx, repository.GetTaskLabelsByBoardIDParams{
		BoardID:  toPgUUID(boardID),
		TenantID: toPgUUID(tenantID),
	})
	if err != nil {
		return nil, err
	}
	result := make([]service.BoardTaskLabelData, len(rows))
	for i, r := range rows {
		result[i] = service.BoardTaskLabelData{
			TaskID:  fromPgUUID(r.TaskID),
			LabelID: fromPgUUID(r.LabelID),
			Name:    r.Name,
			Color:   r.Color,
		}
	}
	return result, nil
}

func (a *BoardRepositoryAdapter) CreateColumn(ctx context.Context, tenantID, boardID uuid.UUID, name, color string, wipLimit *int, position int) (service.ColumnData, error) {
	c, err := a.q.CreateColumn(ctx, repository.CreateColumnParams{
		TenantID: toPgUUID(tenantID),
		BoardID:  toPgUUID(boardID),
		Name:     name,
		Position: int32(position),
		Color:    color,
		WipLimit: toPgInt4(wipLimit),
	})
	if err != nil {
		return service.ColumnData{}, err
	}
	return toColumnData(c), nil
}

func (a *BoardRepositoryAdapter) UpdateColumn(ctx context.Context, id, tenantID uuid.UUID, name, color *string, wipLimit *int) (service.ColumnData, error) {
	c, err := a.q.UpdateColumn(ctx, repository.UpdateColumnParams{
		Name:     toPgText(name),
		Color:    toPgText(color),
		WipLimit: toPgInt4(wipLimit),
		ID:       toPgUUID(id),
		TenantID: toPgUUID(tenantID),
	})
	if err != nil {
		return service.ColumnData{}, err
	}
	return toColumnData(c), nil
}

func (a *BoardRepositoryAdapter) DeleteColumn(ctx context.Context, id, tenantID uuid.UUID) error {
	return a.q.DeleteColumn(ctx, repository.DeleteColumnParams{
		ID:       toPgUUID(id),
		TenantID: toPgUUID(tenantID),
	})
}

func (a *BoardRepositoryAdapter) UpdateColumnPosition(ctx context.Context, id, tenantID uuid.UUID, position int) error {
	return a.q.UpdateColumnPosition(ctx, repository.UpdateColumnPositionParams{
		ID:       toPgUUID(id),
		TenantID: toPgUUID(tenantID),
		Position: int32(position),
	})
}

func (a *BoardRepositoryAdapter) GetMaxColumnPosition(ctx context.Context, boardID, tenantID uuid.UUID) (int, error) {
	pos, err := a.q.GetMaxColumnPosition(ctx, repository.GetMaxColumnPositionParams{
		BoardID:  toPgUUID(boardID),
		TenantID: toPgUUID(tenantID),
	})
	if err != nil {
		return 0, err
	}
	return int(pos), nil
}

func (a *BoardRepositoryAdapter) GetColumnByID(ctx context.Context, id, tenantID uuid.UUID) (service.ColumnData, error) {
	c, err := a.q.GetColumnByID(ctx, repository.GetColumnByIDParams{
		ID:       toPgUUID(id),
		TenantID: toPgUUID(tenantID),
	})
	if err != nil {
		return service.ColumnData{}, err
	}
	return toColumnData(c), nil
}

// Helper conversions

func toBoardData(b repository.Board) service.BoardData {
	return service.BoardData{
		ID:        fromPgUUID(b.ID),
		TenantID:  fromPgUUID(b.TenantID),
		ProjectID: fromPgUUID(b.ProjectID),
		Name:      b.Name,
		CreatedAt: b.CreatedAt.Time,
		UpdatedAt: b.UpdatedAt.Time,
	}
}

func toColumnData(c repository.Column) service.ColumnData {
	return service.ColumnData{
		ID:        fromPgUUID(c.ID),
		BoardID:   fromPgUUID(c.BoardID),
		Name:      c.Name,
		Position:  int(c.Position),
		Color:     c.Color,
		WIPLimit:  fromPgInt4(c.WipLimit),
		CreatedAt: c.CreatedAt.Time,
		UpdatedAt: c.UpdatedAt.Time,
	}
}

func toPgInt4(v *int) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(*v), Valid: true}
}

func fromPgInt4(v pgtype.Int4) *int {
	if !v.Valid {
		return nil
	}
	i := int(v.Int32)
	return &i
}

func fromPgUUIDPtr(id pgtype.UUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}
	u := uuid.UUID(id.Bytes)
	return &u
}

func toPgTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func fromPgTimestamptzPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
