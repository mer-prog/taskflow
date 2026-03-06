package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/mer-prog/taskflow/internal/middleware"
	"github.com/mer-prog/taskflow/internal/model"
	"github.com/mer-prog/taskflow/internal/service"
)

type DashboardHandler struct {
	svc *service.DashboardService
}

func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) Register(g *echo.Group) {
	g.GET("/summary", h.summary)
	g.GET("/overdue", h.overdue)
	g.GET("/my-tasks", h.myTasks)
}

func (h *DashboardHandler) summary(c echo.Context) error {
	tenantID := middleware.GetTenantID(c)
	resp, err := h.svc.GetSummary(c.Request().Context(), tenantID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "failed to get summary"})
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *DashboardHandler) overdue(c echo.Context) error {
	tenantID := middleware.GetTenantID(c)
	tasks, err := h.svc.GetOverdueTasks(c.Request().Context(), tenantID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "failed to get overdue tasks"})
	}
	return c.JSON(http.StatusOK, tasks)
}

func (h *DashboardHandler) myTasks(c echo.Context) error {
	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)
	tasks, err := h.svc.GetMyTasks(c.Request().Context(), userID, tenantID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Code: "INTERNAL_ERROR", Message: "failed to get my tasks"})
	}
	return c.JSON(http.StatusOK, tasks)
}
