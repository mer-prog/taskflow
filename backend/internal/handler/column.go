package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mer-prog/taskflow/internal/middleware"
	"github.com/mer-prog/taskflow/internal/model"
	"github.com/mer-prog/taskflow/internal/service"
	"github.com/mer-prog/taskflow/internal/ws"
)

type ColumnHandler struct {
	svc *service.BoardService
	hub *ws.HubManager
}

func NewColumnHandler(svc *service.BoardService, hub *ws.HubManager) *ColumnHandler {
	return &ColumnHandler{svc: svc, hub: hub}
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
			Code: "BAD_REQUEST", Message: "invalid request body",
		})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "name is required",
		})
	}

	if req.BoardID == uuid.Nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "board_id is required",
		})
	}

	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)
	ctx := c.Request().Context()

	col, err := h.svc.CreateColumn(ctx, tenantID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code: "INTERNAL_ERROR", Message: "internal server error",
		})
	}

	h.broadcast(req.BoardID.String(), "column:created", userID.String(), col)

	return c.JSON(http.StatusCreated, col)
}

func (h *ColumnHandler) update(c echo.Context) error {
	colID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "invalid column id",
		})
	}

	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)

	var req model.UpdateColumnRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "invalid request body",
		})
	}

	ctx := c.Request().Context()
	col, err := h.svc.UpdateColumn(ctx, colID, tenantID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code: "INTERNAL_ERROR", Message: "internal server error",
		})
	}

	h.broadcast(col.BoardID.String(), "column:updated", userID.String(), col)

	return c.JSON(http.StatusOK, col)
}

func (h *ColumnHandler) delete(c echo.Context) error {
	colID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "invalid column id",
		})
	}

	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)
	ctx := c.Request().Context()

	// Look up board_id before deleting
	boardID, _ := h.svc.GetBoardIDByColumnID(ctx, colID, tenantID)

	if err := h.svc.DeleteColumn(ctx, colID, tenantID); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code: "INTERNAL_ERROR", Message: "internal server error",
		})
	}

	if boardID != uuid.Nil {
		h.broadcast(boardID.String(), "column:deleted", userID.String(), map[string]string{"id": colID.String()})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ColumnHandler) reorder(c echo.Context) error {
	var req model.ReorderColumnsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "invalid request body",
		})
	}

	if len(req.ColumnIDs) == 0 {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "column_ids is required",
		})
	}

	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)
	ctx := c.Request().Context()

	if err := h.svc.ReorderColumns(ctx, tenantID, req); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code: "INTERNAL_ERROR", Message: "internal server error",
		})
	}

	// Get board_id from first column
	if boardID, err := h.svc.GetBoardIDByColumnID(ctx, req.ColumnIDs[0], tenantID); err == nil {
		h.broadcast(boardID.String(), "column:reordered", userID.String(), req)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ColumnHandler) broadcast(boardID, eventType, userID string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	h.hub.Broadcast(boardID, ws.WSMessage{
		Type:    eventType,
		Payload: data,
		UserID:  userID,
	})
}
