package handler

import (
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/mer-prog/taskflow/internal/middleware"
	"github.com/mer-prog/taskflow/internal/model"
	"github.com/mer-prog/taskflow/internal/service"
)

var slugRegexp = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type TenantHandler struct {
	svc *service.TenantService
}

func NewTenantHandler(svc *service.TenantService) *TenantHandler {
	return &TenantHandler{svc: svc}
}

func (h *TenantHandler) Register(g *echo.Group) {
	g.POST("", h.create)
	g.GET("", h.list)
	g.GET("/:id", h.get)
	g.PATCH("/:id", h.update)
	g.GET("/:id/members", h.getMembers)
	g.POST("/:id/members", h.addMember)
	g.PATCH("/:id/members/:uid", h.updateMemberRole)
	g.DELETE("/:id/members/:uid", h.removeMember)
}

func (h *TenantHandler) create(c echo.Context) error {
	var req model.CreateTenantRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "BAD_REQUEST", Message: "invalid request body"})
	}
	if req.Name == "" || req.Slug == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "VALIDATION_ERROR", Message: "name and slug are required"})
	}
	if !slugRegexp.MatchString(req.Slug) {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "VALIDATION_ERROR", Message: "slug must be lowercase alphanumeric with hyphens only"})
	}

	userID := middleware.GetUserID(c)
	resp, err := h.svc.Create(c.Request().Context(), req, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "failed to create tenant"})
	}
	return c.JSON(http.StatusCreated, resp)
}

func (h *TenantHandler) list(c echo.Context) error {
	userID := middleware.GetUserID(c)
	tenants, err := h.svc.ListByUser(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "failed to list tenants"})
	}
	return c.JSON(http.StatusOK, tenants)
}

func (h *TenantHandler) get(c echo.Context) error {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "INVALID_ID", Message: "invalid tenant ID"})
	}

	userID := middleware.GetUserID(c)
	if _, err := h.svc.CheckMembership(c.Request().Context(), tenantID, userID); err != nil {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{Code: "FORBIDDEN", Message: "not a member of this tenant"})
	}

	resp, err := h.svc.Get(c.Request().Context(), tenantID)
	if err != nil {
		return c.JSON(http.StatusNotFound, model.ErrorResponse{Code: "NOT_FOUND", Message: "tenant not found"})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *TenantHandler) update(c echo.Context) error {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "INVALID_ID", Message: "invalid tenant ID"})
	}

	userID := middleware.GetUserID(c)
	role, err := h.svc.CheckMembership(c.Request().Context(), tenantID, userID)
	if err != nil {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{Code: "FORBIDDEN", Message: "not a member of this tenant"})
	}
	if !middleware.HasMinRole(role, middleware.RoleAdmin) {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{Code: "INSUFFICIENT_ROLE", Message: "admin or owner role required"})
	}

	var req model.UpdateTenantRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "BAD_REQUEST", Message: "invalid request body"})
	}
	if req.Slug != nil && !slugRegexp.MatchString(*req.Slug) {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "VALIDATION_ERROR", Message: "slug must be lowercase alphanumeric with hyphens only"})
	}

	resp, err := h.svc.Update(c.Request().Context(), tenantID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "failed to update tenant"})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *TenantHandler) getMembers(c echo.Context) error {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "INVALID_ID", Message: "invalid tenant ID"})
	}

	userID := middleware.GetUserID(c)
	if _, err := h.svc.CheckMembership(c.Request().Context(), tenantID, userID); err != nil {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{Code: "FORBIDDEN", Message: "not a member of this tenant"})
	}

	members, err := h.svc.GetMembers(c.Request().Context(), tenantID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "failed to get members"})
	}
	return c.JSON(http.StatusOK, members)
}

func (h *TenantHandler) addMember(c echo.Context) error {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "INVALID_ID", Message: "invalid tenant ID"})
	}

	userID := middleware.GetUserID(c)
	role, err := h.svc.CheckMembership(c.Request().Context(), tenantID, userID)
	if err != nil {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{Code: "FORBIDDEN", Message: "not a member of this tenant"})
	}
	if !middleware.HasMinRole(role, middleware.RoleAdmin) {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{Code: "INSUFFICIENT_ROLE", Message: "admin or owner role required"})
	}

	var req model.AddTenantMemberRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "BAD_REQUEST", Message: "invalid request body"})
	}
	if req.UserID == uuid.Nil || req.Role == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "VALIDATION_ERROR", Message: "user_id and role are required"})
	}

	if err := h.svc.AddMember(c.Request().Context(), tenantID, req.UserID, req.Role); err != nil {
		return c.JSON(http.StatusConflict, model.ErrorResponse{Code: "MEMBER_EXISTS", Message: "user is already a member"})
	}
	return c.NoContent(http.StatusCreated)
}

func (h *TenantHandler) updateMemberRole(c echo.Context) error {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "INVALID_ID", Message: "invalid tenant ID"})
	}
	targetUID, err := uuid.Parse(c.Param("uid"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "INVALID_ID", Message: "invalid user ID"})
	}

	userID := middleware.GetUserID(c)
	role, err := h.svc.CheckMembership(c.Request().Context(), tenantID, userID)
	if err != nil {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{Code: "FORBIDDEN", Message: "not a member of this tenant"})
	}
	if !middleware.HasMinRole(role, middleware.RoleOwner) {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{Code: "INSUFFICIENT_ROLE", Message: "owner role required"})
	}

	var req model.UpdateTenantMemberRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "BAD_REQUEST", Message: "invalid request body"})
	}
	if req.Role == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "VALIDATION_ERROR", Message: "role is required"})
	}

	if err := h.svc.UpdateMemberRole(c.Request().Context(), tenantID, targetUID, req.Role); err != nil {
		return handleTenantError(c, err)
	}
	return c.NoContent(http.StatusOK)
}

func (h *TenantHandler) removeMember(c echo.Context) error {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "INVALID_ID", Message: "invalid tenant ID"})
	}
	targetUID, err := uuid.Parse(c.Param("uid"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "INVALID_ID", Message: "invalid user ID"})
	}

	userID := middleware.GetUserID(c)
	role, err := h.svc.CheckMembership(c.Request().Context(), tenantID, userID)
	if err != nil {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{Code: "FORBIDDEN", Message: "not a member of this tenant"})
	}
	if !middleware.HasMinRole(role, middleware.RoleAdmin) {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{Code: "INSUFFICIENT_ROLE", Message: "admin or owner role required"})
	}

	if err := h.svc.RemoveMember(c.Request().Context(), tenantID, targetUID); err != nil {
		return handleTenantError(c, err)
	}
	return c.NoContent(http.StatusOK)
}

func handleTenantError(c echo.Context, err error) error {
	switch err {
	case service.ErrCannotRemoveOwner:
		return c.JSON(http.StatusForbidden, model.ErrorResponse{Code: "CANNOT_REMOVE_OWNER", Message: "cannot remove or demote the tenant owner"})
	case service.ErrNotTenantMember:
		return c.JSON(http.StatusNotFound, model.ErrorResponse{Code: "NOT_FOUND", Message: "member not found"})
	default:
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "an unexpected error occurred"})
	}
}
