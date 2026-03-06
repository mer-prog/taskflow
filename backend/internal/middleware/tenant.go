package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/mer-prog/taskflow/internal/model"
)

const TenantRoleKey = "tenant_role"

const (
	RoleViewer = "viewer"
	RoleMember = "member"
	RoleAdmin  = "admin"
	RoleOwner  = "owner"
)

var roleLevel = map[string]int{
	RoleViewer: 0,
	RoleMember: 1,
	RoleAdmin:  2,
	RoleOwner:  3,
}

func HasMinRole(userRole, required string) bool {
	return roleLevel[userRole] >= roleLevel[required]
}

func GetTenantRole(c echo.Context) string {
	if r, ok := c.Get(TenantRoleKey).(string); ok {
		return r
	}
	return ""
}

type MembershipChecker interface {
	CheckMembership(ctx context.Context, tenantID, userID uuid.UUID) (string, error)
}

func TenantScope(checker MembershipChecker) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenantIDStr := c.Request().Header.Get("X-Tenant-ID")
			if tenantIDStr == "" {
				return c.JSON(http.StatusBadRequest, model.ErrorResponse{
					Code:    "MISSING_TENANT",
					Message: "X-Tenant-ID header is required",
				})
			}

			tenantID, err := uuid.Parse(tenantIDStr)
			if err != nil {
				return c.JSON(http.StatusBadRequest, model.ErrorResponse{
					Code:    "INVALID_TENANT_ID",
					Message: "invalid tenant ID format",
				})
			}

			userID := GetUserID(c)

			role, err := checker.CheckMembership(c.Request().Context(), tenantID, userID)
			if err != nil {
				return c.JSON(http.StatusForbidden, model.ErrorResponse{
					Code:    "NOT_TENANT_MEMBER",
					Message: "you are not a member of this tenant",
				})
			}

			c.Set(TenantIDKey, tenantID)
			c.Set(TenantRoleKey, role)

			return next(c)
		}
	}
}
