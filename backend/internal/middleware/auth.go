package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/mer-prog/taskflow/internal/config"
	"github.com/mer-prog/taskflow/internal/model"
)

const (
	UserIDKey   = "user_id"
	TenantIDKey = "tenant_id"
)

func JWTAuth(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
					Code:    "UNAUTHORIZED",
					Message: "missing authorization header",
				})
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
					Code:    "UNAUTHORIZED",
					Message: "invalid authorization header format",
				})
			}

			claims := &model.JWTClaims{}
			token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(cfg.JWTSecret), nil
			})
			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
					Code:    "UNAUTHORIZED",
					Message: "invalid or expired token",
				})
			}

			c.Set(UserIDKey, claims.UserID)
			c.Set(TenantIDKey, claims.TenantID)

			return next(c)
		}
	}
}

func GetUserID(c echo.Context) uuid.UUID {
	return c.Get(UserIDKey).(uuid.UUID)
}

func GetTenantID(c echo.Context) uuid.UUID {
	return c.Get(TenantIDKey).(uuid.UUID)
}
