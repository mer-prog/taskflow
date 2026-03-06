package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/mer-prog/taskflow/internal/model"
	"github.com/mer-prog/taskflow/internal/service"
)

type AuthHandler struct {
	svc    *service.AuthService
	secure bool
}

func NewAuthHandler(svc *service.AuthService, secure bool) *AuthHandler {
	return &AuthHandler{svc: svc, secure: secure}
}

func (h *AuthHandler) Register(routes *echo.Group) {
	routes.POST("/auth/register", h.register)
	routes.POST("/auth/login", h.login)
	routes.POST("/auth/refresh", h.refresh)
	routes.POST("/auth/logout", h.logout)
}

func (h *AuthHandler) register(c echo.Context) error {
	var req model.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid request body",
		})
	}

	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "email, password, and display_name are required",
		})
	}

	if len(req.Password) < 8 {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "password must be at least 8 characters",
		})
	}

	resp, refreshToken, err := h.svc.Register(c.Request().Context(), req)
	if err != nil {
		return h.handleAuthError(c, err)
	}

	h.setRefreshCookie(c, refreshToken)
	return c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) login(c echo.Context) error {
	var req model.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "invalid request body",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "email and password are required",
		})
	}

	resp, refreshToken, err := h.svc.Login(c.Request().Context(), req)
	if err != nil {
		return h.handleAuthError(c, err)
	}

	h.setRefreshCookie(c, refreshToken)
	return c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) refresh(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "refresh token not found",
		})
	}

	resp, newRefreshToken, err := h.svc.RefreshAccessToken(c.Request().Context(), cookie.Value)
	if err != nil {
		return h.handleAuthError(c, err)
	}

	h.setRefreshCookie(c, newRefreshToken)
	return c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) logout(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return c.NoContent(http.StatusOK)
	}

	_ = h.svc.Logout(c.Request().Context(), cookie.Value)

	h.clearRefreshCookie(c)
	return c.NoContent(http.StatusOK)
}

func (h *AuthHandler) sameSiteMode() http.SameSite {
	if h.secure {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func (h *AuthHandler) setRefreshCookie(c echo.Context, token string) {
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: h.sameSiteMode(),
		MaxAge:   int(7 * 24 * time.Hour / time.Second),
	})
}

func (h *AuthHandler) clearRefreshCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: h.sameSiteMode(),
		MaxAge:   -1,
	})
}

func (h *AuthHandler) handleAuthError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, service.ErrEmailAlreadyExists):
		return c.JSON(http.StatusConflict, model.ErrorResponse{
			Code:    "EMAIL_EXISTS",
			Message: "an account with this email already exists",
		})
	case errors.Is(err, service.ErrInvalidCredentials):
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Code:    "INVALID_CREDENTIALS",
			Message: "invalid email or password",
		})
	case errors.Is(err, service.ErrInvalidRefreshToken):
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Code:    "INVALID_REFRESH_TOKEN",
			Message: "invalid or expired refresh token",
		})
	default:
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "an unexpected error occurred",
		})
	}
}
