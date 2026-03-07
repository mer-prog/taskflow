package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"github.com/mer-prog/taskflow/internal/config"
	"github.com/mer-prog/taskflow/internal/model"
	"github.com/mer-prog/taskflow/internal/ws"
)

type WSBoardChecker interface {
	CheckMembership(ctx context.Context, tenantID, userID uuid.UUID) (string, error)
}

type WSHandler struct {
	hub     *ws.HubManager
	cfg     *config.Config
	checker WSBoardChecker
}

func NewWSHandler(hub *ws.HubManager, cfg *config.Config, checker WSBoardChecker) *WSHandler {
	return &WSHandler{hub: hub, cfg: cfg, checker: checker}
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
	// Token is passed via Sec-WebSocket-Protocol header to avoid URL exposure
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		protocols := websocket.Subprotocols(c.Request())
		for _, p := range protocols {
			if strings.HasPrefix(p, "access_token.") {
				authHeader = "Bearer " + strings.TrimPrefix(p, "access_token.")
				break
			}
		}
	}

	claims, err := h.parseJWT(authHeader)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Code: "UNAUTHORIZED", Message: "invalid or missing token",
		})
	}

	// Verify tenant membership
	if claims.TenantID != uuid.Nil {
		if _, err := h.checker.CheckMembership(c.Request().Context(), claims.TenantID, claims.UserID); err != nil {
			return c.JSON(http.StatusForbidden, model.ErrorResponse{
				Code: "FORBIDDEN", Message: "not a member of this tenant",
			})
		}
	}

	// Determine the response protocol for the handshake
	var responseProtocol []string
	protocols := websocket.Subprotocols(c.Request())
	for _, p := range protocols {
		if strings.HasPrefix(p, "access_token.") {
			responseProtocol = []string{p}
			break
		}
	}

	upgrader := websocket.Upgrader{
		Subprotocols: responseProtocol,
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
	client := ws.NewClient(hub, conn, claims.UserID.String())
	hub.Register(client)

	go client.WritePump()
	go client.ReadPump()

	return nil
}

func (h *WSHandler) parseJWT(authHeader string) (*model.JWTClaims, error) {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return nil, fmt.Errorf("invalid authorization header")
	}

	claims := &model.JWTClaims{}
	token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
