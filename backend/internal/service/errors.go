package service

import "errors"

var (
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")

	ErrTenantNotFound     = errors.New("tenant not found")
	ErrProjectNotFound    = errors.New("project not found")
	ErrNotTenantMember    = errors.New("not a member of this tenant")
	ErrInsufficientRole   = errors.New("insufficient role")
	ErrSlugAlreadyExists  = errors.New("slug already exists")
	ErrMemberAlreadyExists = errors.New("member already exists")
	ErrCannotRemoveOwner  = errors.New("cannot remove tenant owner")
)
