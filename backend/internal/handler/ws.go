package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"github.com/mer-prog/taskflow/internal/config"
	"github.com/mer-prog/taskflow/internal/model"
	"github.com/mer-prog/taskflow/internal/ws"
)

type WSHandler struct {
	hub *ws.HubManager
	cfg *config.Config
}

func NewWSHandler(hub *ws.HubManager, cfg *config.Config) *WSHandler {
	return &WSHandler{hub: hub, cfg: cfg}
}

func (h *WSHandler) Register(g *echo.Group) {
	g.GET("/ws", h.connect)
}

func (h *WSHandler) connect(c echo.Context) error {
	boardID := c.QueryParam("board_id")
	if boardID == "" {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Code: "BAD_REQUEST", Message: "board_id is required",
		})
	}

	// Manual JWT verification (WebSocket doesn't go through middleware)
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		// Try query param fallback for browser WebSocket API
		authHeader = "Bearer " + c.QueryParam("token")
	}

	userID, err := h.parseJWT(authHeader)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Code: "UNAUTHORIZED", Message: "invalid or missing token",
		})
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if !h.cfg.IsProduction() {
				return true
			}
			origin := r.Header.Get("Origin")
			for _, allowed := range h.cfg.CORSOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
		},
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	hub := h.hub.GetOrCreateHub(boardID)
	client := ws.NewClient(hub, conn, userID)
	hub.Register(client)

	go client.WritePump()
	go client.ReadPump()

	return nil
}

func (h *WSHandler) parseJWT(authHeader string) (string, error) {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return "", fmt.Errorf("invalid authorization header")
	}

	claims := &model.JWTClaims{}
	token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	return claims.UserID.String(), nil
}
