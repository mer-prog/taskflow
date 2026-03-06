package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mer-prog/taskflow/internal/middleware"
	"github.com/mer-prog/taskflow/internal/model"
	"github.com/mer-prog/taskflow/internal/service"
)

type TaskHandler struct {
	svc *service.TaskService
}

func NewTaskHandler(svc *service.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
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
			Code:    "BAD_REQUEST",
			Message: "invalid request body",
		})
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "title is required",
		})
	}

	if req.ColumnID == uuid.Nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "column_id is required",
		})
	}

	tenantID := middleware.GetTenantID(c)
	ctx := c.Request().Context()

	resp, err := h.svc.Create(ctx, tenantID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *TaskHandler) get(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid task id",
		})
	}

	tenantID := middleware.GetTenantID(c)
	ctx := c.Request().Context()

	resp, err := h.svc.Get(ctx, taskID, tenantID)
	if err != nil {
		return c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code:    "NOT_FOUND",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *TaskHandler) update(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid task id",
		})
	}

	tenantID := middleware.GetTenantID(c)

	var req model.UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid request body",
		})
	}

	ctx := c.Request().Context()
	resp, err := h.svc.Update(ctx, taskID, tenantID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *TaskHandler) delete(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid task id",
		})
	}

	tenantID := middleware.GetTenantID(c)
	ctx := c.Request().Context()

	if err := h.svc.Delete(ctx, taskID, tenantID); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *TaskHandler) move(c echo.Context) error {
	var req model.MoveTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid request body",
		})
	}

	if req.TaskID == uuid.Nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "task_id is required",
		})
	}

	if req.ToColumnID == uuid.Nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "to_column_id is required",
		})
	}

	tenantID := middleware.GetTenantID(c)
	ctx := c.Request().Context()

	if err := h.svc.Move(ctx, tenantID, req); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
	}

	return c.NoContent(http.StatusOK)
}

func (h *TaskHandler) addLabel(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid task id",
		})
	}

	var req model.TaskLabelRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid request body",
		})
	}

	if req.LabelID == uuid.Nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "label_id is required",
		})
	}

	ctx := c.Request().Context()

	if err := h.svc.AddLabel(ctx, taskID, req.LabelID); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
	}

	return c.NoContent(http.StatusCreated)
}

func (h *TaskHandler) removeLabel(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid task id",
		})
	}

	labelID, err := uuid.Parse(c.Param("lid"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid label id",
		})
	}

	ctx := c.Request().Context()

	if err := h.svc.RemoveLabel(ctx, taskID, labelID); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
	}

	return c.NoContent(http.StatusOK)
}

func (h *TaskHandler) addComment(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid task id",
		})
	}

	var req model.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid request body",
		})
	}

	if req.Content == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "content is required",
		})
	}

	tenantID := middleware.GetTenantID(c)
	userID := middleware.GetUserID(c)
	ctx := c.Request().Context()

	resp, err := h.svc.CreateComment(ctx, tenantID, taskID, userID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *TaskHandler) getComments(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid task id",
		})
	}

	tenantID := middleware.GetTenantID(c)
	ctx := c.Request().Context()

	resp, err := h.svc.GetComments(ctx, taskID, tenantID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}
