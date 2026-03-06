package model

import (
	"time"

	"github.com/google/uuid"
)

type DashboardSummary struct {
	TotalTasks   int           `json:"total_tasks"`
	OverdueTasks int           `json:"overdue_tasks"`
	ByColumn     []ColumnCount `json:"by_column"`
}

type ColumnCount struct {
	ColumnID   uuid.UUID `json:"column_id"`
	ColumnName string    `json:"column_name"`
	BoardName  string    `json:"board_name"`
	TaskCount  int       `json:"task_count"`
}

type DashboardTask struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	Priority     string     `json:"priority"`
	DueDate      *time.Time `json:"due_date"`
	ColumnName   string     `json:"column_name"`
	BoardName    string     `json:"board_name,omitempty"`
	AssigneeName *string    `json:"assignee_name,omitempty"`
}
