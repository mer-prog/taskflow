package model

import "github.com/google/uuid"

type CreateLabelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type LabelResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
}
