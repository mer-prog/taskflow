package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mer-prog/taskflow/internal/middleware"
	"github.com/mer-prog/taskflow/internal/model"
	"github.com/mer-prog/taskflow/internal/service"
)

type BoardHandler struct {
	svc *service.BoardService
}

func NewBoardHandler(svc *service.BoardService) *BoardHandler {
	return &BoardHandler{svc: svc}
}

func (h *BoardHandler) Register(g *echo.Group) {
	g.POST("", h.create)
	g.GET("/:id", h.get)
	g.PATCH("/:id", h.update)
	g.DELETE("/:id", h.delete)
}

func (h *BoardHandler) RegisterProjectRoutes(g *echo.Group) {
	g.GET("/:id/boards", h.listByProject)
}

func toBoardResponse(b service.BoardData) model.BoardResponse {
	return model.BoardResponse{
		ID:        b.ID,
		ProjectID: b.ProjectID,
		Name:      b.Name,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}

func (h *BoardHandler) listByProject(c echo.Context) error {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid project id",
		})
	}

	tenantID := middleware.GetTenantID(c)
	ctx := c.Request().Context()

	boards, err := h.svc.ListByProject(ctx, projectID, tenantID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
		})
	}

	result := make([]model.BoardResponse, len(boards))
	for i, b := range boards {
		result[i] = toBoardResponse(b)
	}

	return c.JSON(http.StatusOK, result)
}

func (h *BoardHandler) create(c echo.Context) error {
	var req model.CreateBoardRequest
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

	tenantID := middleware.GetTenantID(c)
	ctx := c.Request().Context()

	board, err := h.svc.Create(ctx, tenantID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
		})
	}

	return c.JSON(http.StatusCreated, toBoardResponse(*board))
}

func (h *BoardHandler) get(c echo.Context) error {
	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid board id",
		})
	}

	tenantID := middleware.GetTenantID(c)
	ctx := c.Request().Context()

	resp, err := h.svc.Get(ctx, boardID, tenantID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *BoardHandler) update(c echo.Context) error {
	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid board id",
		})
	}

	tenantID := middleware.GetTenantID(c)

	var req model.UpdateBoardRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid request body",
		})
	}

	ctx := c.Request().Context()
	board, err := h.svc.Update(ctx, boardID, tenantID, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
		})
	}

	return c.JSON(http.StatusOK, toBoardResponse(*board))
}

func (h *BoardHandler) delete(c echo.Context) error {
	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid board id",
		})
	}

	tenantID := middleware.GetTenantID(c)
	ctx := c.Request().Context()

	if err := h.svc.Delete(ctx, boardID, tenantID); err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
		})
	}

	return c.NoContent(http.StatusNoContent)
}
