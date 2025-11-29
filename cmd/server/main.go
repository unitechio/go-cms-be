package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/owner/go-cms/internal/adapters/external/email"
	"github.com/owner/go-cms/internal/adapters/repositories"
	"github.com/owner/go-cms/internal/adapters/repositories/postgres"
	"github.com/owner/go-cms/internal/config"
	"github.com/owner/go-cms/internal/core/usecases"
	"github.com/owner/go-cms/internal/core/usecases/audit"
	"github.com/owner/go-cms/internal/core/usecases/auth"
	"github.com/owner/go-cms/internal/core/usecases/authorization"
	"github.com/owner/go-cms/internal/core/usecases/user"
	"github.com/owner/go-cms/internal/http/handlers"
	authHandlers "github.com/owner/go-cms/internal/http/handlers/authorization"
	"github.com/owner/go-cms/internal/http/middleware"
	"github.com/owner/go-cms/internal/http/router"
	"github.com/owner/go-cms/internal/infrastructure/cache"
	"github.com/owner/go-cms/internal/infrastructure/database"
	"github.com/owner/go-cms/internal/infrastructure/storage"
	"github.com/owner/go-cms/internal/infrastructure/websocket"
	"github.com/owner/go-cms/pkg/logger"
	"go.uber.org/zap"
)

// @title GO CMS API
// @version 1.0
// @description Enterprise CRM System API with comprehensive features
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	env := os.Getenv("APP_ENV")
	configFile := ".env.development"
	if env == "production" {
		configFile = ".env.production"
	}
	log.Printf("Loading configuration from %s", configFile)

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(cfg.Logging.Level, cfg.Logging.Format, cfg.Logging.Output); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting GO CMS application...")

	if err := database.Init(&cfg.Database); err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer database.Close()

	// Run migrations
	if err := database.AutoMigrate(database.GetDB()); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Create indexes
	if err := database.CreateIndexes(); err != nil {
		logger.Warn("Failed to create some indexes", zap.Error(err))
	}

	// Seed initial data
	if err := database.SeedData(database.GetDB()); err != nil {
		logger.Warn("Failed to seed data", zap.Error(err))
	}

	// Initialize Redis
	if err := cache.Init(&cfg.Redis); err != nil {
		logger.Fatal("Failed to initialize Redis", zap.Error(err))
	}
	defer cache.Close()

	// Initialize MinIO
	if err := storage.Init(&cfg.MinIO); err != nil {
		logger.Fatal("Failed to initialize MinIO", zap.Error(err))
	}

	// Initialize email service
	emailService := email.NewService(&cfg.SMTP)

	// Initialize repositories
	userRepo := postgres.NewUserRepository(database.GetDB())
	otpRepo := postgres.NewOTPRepository(database.GetDB())
	refreshTokenRepo := postgres.NewRefreshTokenRepository(database.GetDB())

	// Initialize authorization repositories
	moduleRepo := postgres.NewModuleRepository(database.GetDB())
	departmentRepo := postgres.NewDepartmentRepository(database.GetDB())
	serviceRepo := postgres.NewServiceRepository(database.GetDB())
	scopeRepo := postgres.NewScopeRepository(database.GetDB())
	roleRepo := postgres.NewRoleRepository(database.GetDB())
	permissionRepo := postgres.NewPermissionRepository(database.GetDB())

	// Initialize notification repository
	notificationRepo := repositories.NewNotificationRepository(database.GetDB())

	// Initialize audit log repository
	auditLogRepo := postgres.NewAuditLogRepository(database.GetDB())

	// Initialize use cases
	authUseCase := auth.NewUseCase(userRepo, otpRepo, refreshTokenRepo, cfg, emailService)
	userUseCase := user.NewUserUseCase(userRepo, roleRepo)

	// Initialize authorization use cases
	moduleUseCase := authorization.NewModuleUseCase(moduleRepo)
	departmentUseCase := authorization.NewDepartmentUseCase(departmentRepo, moduleRepo)
	serviceUseCase := authorization.NewServiceUseCase(serviceRepo, departmentRepo)
	scopeUseCase := authorization.NewScopeUseCase(scopeRepo)
	roleUseCase := authorization.NewRoleUseCase(roleRepo, permissionRepo)
	permissionUseCase := authorization.NewPermissionUseCase(permissionRepo)

	// Initialize WebSocket Hub
	wsHub := websocket.NewHub(logger.GetLogger())
	// Start WebSocket Hub in background
	go wsHub.(*websocket.Hub).Run()

	// Initialize notification use case
	notificationUseCase := usecases.NewNotificationUseCase(notificationRepo, wsHub, logger.GetLogger())

	// Initialize audit log use case
	auditLogUseCase := audit.NewUseCase(auditLogRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authUseCase)
	userHandler := handlers.NewUserHandler(userUseCase)

	// Initialize authorization handlers
	moduleHandler := authHandlers.NewModuleHandler(moduleUseCase)
	departmentHandler := authHandlers.NewDepartmentHandler(departmentUseCase)
	serviceHandler := authHandlers.NewServiceHandler(serviceUseCase)
	scopeHandler := authHandlers.NewScopeHandler(scopeUseCase)
	roleHandler := handlers.NewRoleHandler(roleUseCase)
	permissionHandler := handlers.NewPermissionHandler(permissionUseCase)

	// Initialize notification and WebSocket handlers
	notificationHandler := handlers.NewNotificationHandler(notificationUseCase, logger.GetLogger())
	websocketHandler := handlers.NewWebSocketHandler(wsHub, logger.GetLogger())

	// Initialize audit log handler
	auditLogHandler := handlers.NewAuditLogHandler(auditLogUseCase)

	// Initialize permission checker
	permissionChecker := middleware.NewPermissionChecker(database.GetDB())

	// Setup router
	r := router.NewRouter(
		cfg,
		permissionChecker,
		authHandler,
		userHandler,
		moduleHandler,
		departmentHandler,
		serviceHandler,
		scopeHandler,
		roleHandler,
		permissionHandler,
		notificationHandler,
		websocketHandler,
		auditLogHandler,
		auditLogUseCase,
	)
	engine := r.Setup()

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.Server.GetServerAddr(),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server starting",
			zap.String("address", cfg.Server.GetServerAddr()),
			zap.String("environment", os.Getenv("ENV")),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
