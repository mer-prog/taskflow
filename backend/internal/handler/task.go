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

type TaskHandler struct {
	svc      *service.TaskService
	boardSvc *service.BoardService
	hub      *ws.HubManager
}

func NewTaskHandler(svc *service.TaskService, boardSvc *service.BoardService, hub *ws.HubManager) *TaskHandler {
	return &TaskHandler{svc: svc, boardSvc: boardSvc, hub: hub}
}

func (h *TaskHandler) Register(g *echo.Group) {
	g.POST("", h.create)
	g.PATCH("/move", h.move)
	g.GET("/:id", h.get)
	g.PATCH("/:id", h.update)
	g.DELETE("/:id", h.delete)
	g.POST("/:id/labels", h.addLabel)
	g.DELETE("/:id/labels/:lid", h.removeLabel)
	g.POST("/:id/comments", h.addComment)
	g.GET("/:id/comments", h.getComments)
}

func (h *TaskHandler) create(c echo.Context) error {
	var req model.CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "invalid request body",
		})
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "title is required",
		})
	}

	if req.ColumnID == uuid.Nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "column_id is required",
		})
	}

	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)
	ctx := c.Request().Context()

	resp, err := h.svc.Create(ctx, tenantID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code: "INTERNAL_ERROR", Message: "internal server error",
		})
	}

	if boardID, err := h.boardSvc.GetBoardIDByColumnID(ctx, req.ColumnID, tenantID); err == nil {
		h.broadcast(boardID.String(), "task:created", userID.String(), resp)
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *TaskHandler) get(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "invalid task id",
		})
	}

	tenantID := middleware.GetTenantID(c)
	ctx := c.Request().Context()

	resp, err := h.svc.Get(ctx, taskID, tenantID)
	if err != nil {
		return c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code: "NOT_FOUND", Message: "task not found",
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *TaskHandler) update(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "invalid task id",
		})
	}

	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)

	var req model.UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "invalid request body",
		})
	}

	ctx := c.Request().Context()
	resp, err := h.svc.Update(ctx, taskID, tenantID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code: "INTERNAL_ERROR", Message: "internal server error",
		})
	}

	if boardID, err := h.boardSvc.GetBoardIDByColumnID(ctx, resp.ColumnID, tenantID); err == nil {
		h.broadcast(boardID.String(), "task:updated", userID.String(), resp)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *TaskHandler) delete(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "invalid task id",
		})
	}

	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)
	ctx := c.Request().Context()

	// Look up board_id before deleting
	task, err := h.svc.Get(ctx, taskID, tenantID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code: "INTERNAL_ERROR", Message: "internal server error",
		})
	}
	boardID, _ := h.boardSvc.GetBoardIDByColumnID(ctx, task.ColumnID, tenantID)

	if err := h.svc.Delete(ctx, taskID, tenantID); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code: "INTERNAL_ERROR", Message: "internal server error",
		})
	}

	if boardID != uuid.Nil {
		h.broadcast(boardID.String(), "task:deleted", userID.String(), map[string]string{"id": taskID.String()})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *TaskHandler) move(c echo.Context) error {
	var req model.MoveTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "invalid request body",
		})
	}

	if req.TaskID == uuid.Nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "task_id is required",
		})
	}

	if req.ToColumnID == uuid.Nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "to_column_id is required",
		})
	}

	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)
	ctx := c.Request().Context()

	if err := h.svc.Move(ctx, tenantID, req); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code: "INTERNAL_ERROR", Message: "internal server error",
		})
	}

	if boardID, err := h.boardSvc.GetBoardIDByColumnID(ctx, req.ToColumnID, tenantID); err == nil {
		h.broadcast(boardID.String(), "task:moved", userID.String(), req)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *TaskHandler) addLabel(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "invalid task id",
		})
	}

	var req model.TaskLabelRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "invalid request body",
		})
	}

	if req.LabelID == uuid.Nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "label_id is required",
		})
	}

	ctx := c.Request().Context()

	if err := h.svc.AddLabel(ctx, taskID, req.LabelID); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code: "INTERNAL_ERROR", Message: "internal server error",
		})
	}

	return c.NoContent(http.StatusCreated)
}

func (h *TaskHandler) removeLabel(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "invalid task id",
		})
	}

	labelID, err := uuid.Parse(c.Param("lid"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "invalid label id",
		})
	}

	ctx := c.Request().Context()

	if err := h.svc.RemoveLabel(ctx, taskID, labelID); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code: "INTERNAL_ERROR", Message: "internal server error",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *TaskHandler) addComment(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "invalid task id",
		})
	}

	var req model.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "invalid request body",
		})
	}

	if req.Content == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "content is required",
		})
	}

	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)
	ctx := c.Request().Context()

	resp, err := h.svc.CreateComment(ctx, tenantID, taskID, userID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code: "INTERNAL_ERROR", Message: "internal server error",
		})
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *TaskHandler) getComments(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "invalid task id",
		})
	}

	tenantID := middleware.GetTenantID(c)
	ctx := c.Request().Context()

	resp, err := h.svc.GetComments(ctx, taskID, tenantID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code: "INTERNAL_ERROR", Message: "internal server error",
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *TaskHandler) broadcast(boardID, eventType, userID string, payload interface{}) {
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
