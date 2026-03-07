package model

import (
	"time"

	"github.com/google/uuid"
)

type CreateBoardRequest struct {
	ProjectID uuid.UUID `json:"project_id"`
	Name      string    `json:"name"`
}

type UpdateBoardRequest struct {
	Name *string `json:"name"`
}

type BoardResponse struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BoardDetailResponse struct {
	ID        uuid.UUID          `json:"id"`
	ProjectID uuid.UUID          `json:"project_id"`
	Name      string             `json:"name"`
	Columns   []ColumnWithTasks  `json:"columns"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type ColumnWithTasks struct {
	ID       uuid.UUID     `json:"id"`
	Name     string        `json:"name"`
	Position int           `json:"position"`
	Color    string        `json:"color"`
	WIPLimit *int          `json:"wip_limit"`
	Tasks    []TaskSummary `json:"tasks"`
}

type TaskSummary struct {
	ID       uuid.UUID     `json:"id"`
	Title    string        `json:"title"`
	Position int           `json:"position"`
	Priority string        `json:"priority"`
	DueDate  *time.Time    `json:"due_date"`
	Assignee *AssigneeInfo `json:"assignee"`
	Labels   []LabelInfo   `json:"labels"`
}

type AssigneeInfo struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
}

type LabelInfo struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
}
