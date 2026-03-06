package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/mer-prog/taskflow/internal/middleware"
	"github.com/mer-prog/taskflow/internal/model"
	"github.com/mer-prog/taskflow/internal/service"
)

type LabelHandler struct {
	svc *service.TaskService
}

func NewLabelHandler(svc *service.TaskService) *LabelHandler {
	return &LabelHandler{svc: svc}
}

func (h *LabelHandler) Register(scoped *echo.Group) {
	scoped.GET("/projects/:id/labels", h.list)
	scoped.POST("/projects/:id/labels", h.create)
	scoped.DELETE("/labels/:id", h.delete)
}

func (h *LabelHandler) list(c echo.Context) error {
	tenantID := middleware.GetTenantID(c)
	labels, err := h.svc.ListLabels(c.Request().Context(), tenantID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "failed to list labels"})
	}
	return c.JSON(http.StatusOK, labels)
}

func (h *LabelHandler) create(c echo.Context) error {
	var req model.CreateLabelRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "BAD_REQUEST", Message: "invalid request body"})
	}
	if req.Name == "" || req.Color == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "VALIDATION_ERROR", Message: "name and color are required"})
	}

	tenantID := middleware.GetTenantID(c)
	resp, err := h.svc.CreateLabel(c.Request().Context(), tenantID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "failed to create label"})
	}
	return c.JSON(http.StatusCreated, resp)
}

func (h *LabelHandler) delete(c echo.Context) error {
	labelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Code: "INVALID_ID", Message: "invalid label ID"})
	}

	tenantID := middleware.GetTenantID(c)
	if err := h.svc.DeleteLabel(c.Request().Context(), labelID, tenantID); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "failed to delete label"})
	}
	return c.NoContent(http.StatusNoContent)
}
