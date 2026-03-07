package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/mer-prog/taskflow/internal/middleware"
	"github.com/mer-prog/taskflow/internal/model"
	"github.com/mer-prog/taskflow/internal/service"
)

type ProjectHandler struct {
	svc *service.ProjectService
}

func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

func (h *ProjectHandler) Register(g *echo.Group) {
	g.GET("", h.list)
	g.POST("", h.create)
	g.GET("/:id", h.get)
	g.PATCH("/:id", h.update)
	g.DELETE("/:id", h.archive)
	g.GET("/:id/members", h.getMembers)
	g.POST("/:id/members", h.addMember)
	g.DELETE("/:id/members/:uid", h.removeMember)
}


func (h *ProjectHandler) list(c echo.Context) error {
	tenantID := middleware.GetTenantID(c)
	projects, err := h.svc.List(c.Request().Context(), tenantID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "failed to list projects"})
	}
	return c.JSON(http.StatusOK, projects)
}

func (h *ProjectHandler) create(c echo.Context) error {
	role := middleware.GetTenantRole(c)
	if !middleware.HasMinRole(role, middleware.RoleMember) {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{Code: "INSUFFICIENT_ROLE", Message: "member role or above required"})
	}

	var req model.CreateProjectRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "BAD_REQUEST", Message: "invalid request body"})
	}
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "VALIDATION_ERROR", Message: "name is required"})
	}

	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)

	resp, err := h.svc.Create(c.Request().Context(), tenantID, userID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "failed to create project"})
	}
	return c.JSON(http.StatusCreated, resp)
}

func (h *ProjectHandler) get(c echo.Context) error {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "INVALID_ID", Message: "invalid project ID"})
	}

	tenantID := middleware.GetTenantID(c)
	resp, err := h.svc.Get(c.Request().Context(), projectID, tenantID)
	if err != nil {
		return c.JSON(http.StatusNotFound, model.ErrorResponse{Code: "NOT_FOUND", Message: "project not found"})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *ProjectHandler) update(c echo.Context) error {
	role := middleware.GetTenantRole(c)
	if !middleware.HasMinRole(role, middleware.RoleMember) {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{Code: "INSUFFICIENT_ROLE", Message: "member role or above required"})
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "INVALID_ID", Message: "invalid project ID"})
	}

	var req model.UpdateProjectRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "BAD_REQUEST", Message: "invalid request body"})
	}

	tenantID := middleware.GetTenantID(c)
	resp, err := h.svc.Update(c.Request().Context(), projectID, tenantID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "failed to update project"})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *ProjectHandler) archive(c echo.Context) error {
	role := middleware.GetTenantRole(c)
	if !middleware.HasMinRole(role, middleware.RoleAdmin) {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{Code: "INSUFFICIENT_ROLE", Message: "admin role or above required"})
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "INVALID_ID", Message: "invalid project ID"})
	}

	tenantID := middleware.GetTenantID(c)
	if err := h.svc.Archive(c.Request().Context(), projectID, tenantID); err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{Code: "NOT_FOUND", Message: "project not found"})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "failed to archive project"})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *ProjectHandler) getMembers(c echo.Context) error {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "INVALID_ID", Message: "invalid project ID"})
	}

	members, err := h.svc.GetMembers(c.Request().Context(), projectID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "failed to get members"})
	}
	return c.JSON(http.StatusOK, members)
}

func (h *ProjectHandler) addMember(c echo.Context) error {
	role := middleware.GetTenantRole(c)
	if !middleware.HasMinRole(role, middleware.RoleAdmin) {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{Code: "INSUFFICIENT_ROLE", Message: "admin role or above required"})
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "INVALID_ID", Message: "invalid project ID"})
	}

	var req model.AddProjectMemberRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "BAD_REQUEST", Message: "invalid request body"})
	}
	if req.UserID == uuid.Nil || req.Role == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "VALIDATION_ERROR", Message: "user_id and role are required"})
	}

	if err := h.svc.AddMember(c.Request().Context(), projectID, req.UserID, req.Role); err != nil {
		return c.JSON(http.StatusConflict, model.ErrorResponse{Code: "MEMBER_EXISTS", Message: "user is already a project member"})
	}
	return c.NoContent(http.StatusCreated)
}

func (h *ProjectHandler) removeMember(c echo.Context) error {
	role := middleware.GetTenantRole(c)
	if !middleware.HasMinRole(role, middleware.RoleAdmin) {
		return c.JSON(http.StatusForbidden, model.ErrorResponse{Code: "INSUFFICIENT_ROLE", Message: "admin role or above required"})
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "INVALID_ID", Message: "invalid project ID"})
	}
	targetUID, err := uuid.Parse(c.Param("uid"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "INVALID_ID", Message: "invalid user ID"})
	}

	if err := h.svc.RemoveMember(c.Request().Context(), projectID, targetUID); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "failed to remove member"})
	}
	return c.NoContent(http.StatusNoContent)
}
