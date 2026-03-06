package model

import (
	"time"

	"github.com/google/uuid"
)

type CreateTenantRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type UpdateTenantRequest struct {
	Name *string `json:"name"`
	Slug *string `json:"slug"`
}

type TenantResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AddTenantMemberRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
}

type UpdateTenantMemberRoleRequest struct {
	Role string `json:"role"`
}

type MemberResponse struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}
