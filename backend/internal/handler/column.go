package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mer-prog/taskflow/internal/middleware"
	"github.com/mer-prog/taskflow/internal/model"
	"github.com/mer-prog/taskflow/internal/service"
)

type ColumnHandler struct {
	svc *service.BoardService
}

func NewColumnHandler(svc *service.BoardService) *ColumnHandler {
	return &ColumnHandler{svc: svc}
}

func (h *ColumnHandler) Register(g *echo.Group) {
	g.POST("", h.create)
	g.PATCH("/reorder", h.reorder)
	g.PATCH("/:id", h.update)
	g.DELETE("/:id", h.delete)
}

func (h *ColumnHandler) create(c echo.Context) error {
	var req model.CreateColumnRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid request body",
		})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "name is required",
		})
	}

	if req.BoardID == uuid.Nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "board_id is required",
		})
	}

	tenantID := middleware.GetTenantID(c)
	ctx := c.Request().Context()

	col, err := h.svc.CreateColumn(ctx, tenantID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, col)
}

func (h *ColumnHandler) update(c echo.Context) error {
	colID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid column id",
		})
	}

	tenantID := middleware.GetTenantID(c)

	var req model.UpdateColumnRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid request body",
		})
	}

	ctx := c.Request().Context()
	col, err := h.svc.UpdateColumn(ctx, colID, tenantID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, col)
}

func (h *ColumnHandler) delete(c echo.Context) error {
	colID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid column id",
		})
	}

	tenantID := middleware.GetTenantID(c)
	ctx := c.Request().Context()

	if err := h.svc.DeleteColumn(ctx, colID, tenantID); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ColumnHandler) reorder(c echo.Context) error {
	var req model.ReorderColumnsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid request body",
		})
	}

	if len(req.ColumnIDs) == 0 {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "column_ids is required",
		})
	}

	tenantID := middleware.GetTenantID(c)
	ctx := c.Request().Context()

	if err := h.svc.ReorderColumns(ctx, tenantID, req); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}
