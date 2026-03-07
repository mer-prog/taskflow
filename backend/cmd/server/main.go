package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/mer-prog/taskflow/internal/adapter"
	"github.com/mer-prog/taskflow/internal/config"
	"github.com/mer-prog/taskflow/internal/handler"
	"github.com/mer-prog/taskflow/internal/middleware"
	"github.com/mer-prog/taskflow/internal/model"
	"github.com/mer-prog/taskflow/internal/repository"
	"github.com/mer-prog/taskflow/internal/service"
	"github.com/mer-prog/taskflow/internal/ws"
)

func main() {
	cfg := config.Load()

	pool, err := pgxpool.New(context.Background(), cfg.DBURL())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("connected to database")

	queries := repository.New(pool)

	// Adapters
	authRepo := adapter.NewAuthRepository(queries)
	tenantRepo := adapter.NewTenantRepository(queries)
	projectRepo := adapter.NewProjectRepository(queries)
	boardRepo := adapter.NewBoardRepository(queries)
	taskRepo := adapter.NewTaskRepository(queries)
	dashboardRepo := adapter.NewDashboardRepository(queries)

	// Services
	authSvc := service.NewAuthService(authRepo, cfg)
	tenantSvc := service.NewTenantService(tenantRepo)
	projectSvc := service.NewProjectService(projectRepo)
	boardSvc := service.NewBoardService(boardRepo)
	taskSvc := service.NewTaskService(taskRepo)
	dashboardSvc := service.NewDashboardService(dashboardRepo)

	// WebSocket
	hubManager := ws.NewHubManager()

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc, cfg.IsProduction())
	tenantHandler := handler.NewTenantHandler(tenantSvc)
	projectHandler := handler.NewProjectHandler(projectSvc)
	boardHandler := handler.NewBoardHandler(boardSvc)
	columnHandler := handler.NewColumnHandler(boardSvc, hubManager)
	taskHandler := handler.NewTaskHandler(taskSvc, boardSvc, hubManager)
	labelHandler := handler.NewLabelHandler(taskSvc)
	dashboardHandler := handler.NewDashboardHandler(dashboardSvc)
	wsHandler := handler.NewWSHandler(hubManager, cfg)

	e := echo.New()
	e.HideBanner = true

	e.Use(echomw.Logger())
	e.Use(echomw.Recover())
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-Tenant-ID"},
		AllowCredentials: true,
	}))

	v1 := e.Group("/api/v1")

	v1.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, model.ErrorResponse{Code: "OK", Message: "healthy"})
	})

	// Auth routes (public)
	authHandler.Register(v1)

	// WebSocket route (auth handled inside handler)
	wsHandler.Register(v1)

	jwtMw := middleware.JWTAuth(cfg)
	tenantScope := middleware.TenantScope(tenantSvc)

	// Tenant routes (JWT only)
	tenants := v1.Group("/tenants", jwtMw)
	tenantHandler.Register(tenants)

	// Tenant-scoped routes (JWT + X-Tenant-ID)
	scoped := v1.Group("", jwtMw, tenantScope)

	projects := scoped.Group("/projects")
	projectHandler.Register(projects)
	boardHandler.RegisterProjectRoutes(projects)

	boards := scoped.Group("/boards")
	boardHandler.Register(boards)

	columns := scoped.Group("/columns")
	columnHandler.Register(columns)

	tasks := scoped.Group("/tasks")
	taskHandler.Register(tasks)

	labelHandler.Register(scoped)

	dashboard := scoped.Group("/dashboard")
	dashboardHandler.Register(dashboard)

	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	log.Println("server stopped")
}
