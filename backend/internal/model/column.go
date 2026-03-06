package model

import (
	"time"

	"github.com/google/uuid"
)

type CreateColumnRequest struct {
	BoardID  uuid.UUID `json:"board_id"`
	Name     string    `json:"name"`
	Color    string    `json:"color"`
	WIPLimit *int      `json:"wip_limit"`
}

type UpdateColumnRequest struct {
	Name     *string `json:"name"`
	Color    *string `json:"color"`
	WIPLimit *int    `json:"wip_limit"`
}

type ReorderColumnsRequest struct {
	ColumnIDs []uuid.UUID `json:"column_ids"`
}

type ColumnResponse struct {
	ID        uuid.UUID `json:"id"`
	BoardID   uuid.UUID `json:"board_id"`
	Name      string    `json:"name"`
	Position  int       `json:"position"`
	Color     string    `json:"color"`
	WIPLimit  *int      `json:"wip_limit"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
