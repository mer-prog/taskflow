package model

import (
	"time"

	"github.com/google/uuid"
)

type CreateTaskRequest struct {
	ColumnID    uuid.UUID  `json:"column_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	AssigneeID  *uuid.UUID `json:"assignee_id"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Priority    *string    `json:"priority"`
	AssigneeID  *uuid.UUID `json:"assignee_id"`
	DueDate     *time.Time `json:"due_date"`
}

type MoveTaskRequest struct {
	TaskID      uuid.UUID `json:"task_id"`
	ToColumnID  uuid.UUID `json:"to_column_id"`
	NewPosition int       `json:"new_position"`
}

type TaskResponse struct {
	ID          uuid.UUID     `json:"id"`
	ColumnID    uuid.UUID     `json:"column_id"`
	Title       string        `json:"title"`
	Description *string       `json:"description,omitempty"`
	Position    int           `json:"position"`
	Priority    string        `json:"priority"`
	Assignee    *AssigneeInfo `json:"assignee"`
	DueDate     *time.Time    `json:"due_date"`
	Labels      []LabelInfo   `json:"labels"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type TaskLabelRequest struct {
	LabelID uuid.UUID `json:"label_id"`
}

type CreateCommentRequest struct {
	Content string `json:"content"`
}

type CommentResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
