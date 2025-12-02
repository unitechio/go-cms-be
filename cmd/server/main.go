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
	"github.com/owner/go-cms/internal/core/usecases/page_builder"
	"github.com/owner/go-cms/internal/core/usecases/user"
	"github.com/owner/go-cms/internal/http/handlers"
	authHandlers "github.com/owner/go-cms/internal/http/handlers/authorization"
	pageBuilderHandlers "github.com/owner/go-cms/internal/http/handlers/page_builder"
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

	log := logger.GetLogger()

	logger.Info("Starting GO CMS application...")

	if err := database.Init(&cfg.Database); err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer database.Close()

	db := database.GetDB()

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Create indexes
	if err := database.CreateIndexes(); err != nil {
		logger.Warn("Failed to create some indexes", zap.Error(err))
	}

	// Seed initial data
	if err := database.SeedData(db); err != nil {
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
	userRepo := postgres.NewUserRepository(db)
	otpRepo := postgres.NewOTPRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)

	// Initialize authorization repositories
	moduleRepo := postgres.NewModuleRepository(db)
	departmentRepo := postgres.NewDepartmentRepository(db)
	serviceRepo := postgres.NewServiceRepository(db)
	scopeRepo := postgres.NewScopeRepository(db)
	roleRepo := postgres.NewRoleRepository(db)
	permissionRepo := postgres.NewPermissionRepository(db)

	// Initialize notification repository
	notificationRepo := repositories.NewNotificationRepository(db)

	// Initialize audit log repository
	auditLogRepo := postgres.NewAuditLogRepository(db)

	// Initialize Page Builder repositories
	pageRepo := postgres.NewPageRepository(db)
	blockRepo := postgres.NewBlockRepository(db)
	pageBlockRepo := postgres.NewPageBlockRepository(db)
	pageVersionRepo := postgres.NewPageVersionRepository(db)
	themeSettingRepo := postgres.NewThemeSettingRepository(db)

	// Initialize use cases
	authUseCase := auth.NewUseCase(userRepo, otpRepo, refreshTokenRepo, cfg, emailService)
	userUseCase := user.NewUserUseCase(userRepo, roleRepo, departmentRepo)

	// Initialize authorization use cases
	moduleUseCase := authorization.NewModuleUseCase(moduleRepo)
	departmentUseCase := authorization.NewDepartmentUseCase(departmentRepo, moduleRepo)
	serviceUseCase := authorization.NewServiceUseCase(serviceRepo, departmentRepo)
	scopeUseCase := authorization.NewScopeUseCase(scopeRepo)
	roleUseCase := authorization.NewRoleUseCase(roleRepo, permissionRepo)
	permissionUseCase := authorization.NewPermissionUseCase(permissionRepo)

	// Initialize WebSocket Hub
	wsHub := websocket.NewHub(log)
	// Start WebSocket Hub in background
	go wsHub.(*websocket.Hub).Run()

	// Initialize notification use case
	notificationUseCase := usecases.NewNotificationUseCase(notificationRepo, wsHub, log)

	// Initialize audit log use case
	auditLogUseCase := audit.NewUseCase(auditLogRepo)

	// Initialize Page Builder use cases
	pageUseCase := page_builder.NewPageUseCase(pageRepo, pageVersionRepo)
	blockUseCase := page_builder.NewBlockUseCase(blockRepo)
	pageBlockUseCase := page_builder.NewPageBlockUseCase(pageBlockRepo, blockRepo)
	pageVersionUseCase := page_builder.NewPageVersionUseCase(pageVersionRepo, pageRepo, pageBlockRepo)
	themeSettingUseCase := page_builder.NewThemeSettingUseCase(themeSettingRepo)

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
	notificationHandler := handlers.NewNotificationHandler(notificationUseCase, log)
	websocketHandler := handlers.NewWebSocketHandler(wsHub, log)

	// Initialize audit log handler
	auditLogHandler := handlers.NewAuditLogHandler(auditLogUseCase)

	// Initialize Page Builder handlers
	pageHandler := pageBuilderHandlers.NewPageHandler(pageUseCase)
	blockHandler := pageBuilderHandlers.NewBlockHandler(blockUseCase)
	pageBlockHandler := pageBuilderHandlers.NewPageBlockHandler(pageBlockUseCase)
	pageVersionHandler := pageBuilderHandlers.NewPageVersionHandler(pageVersionUseCase)
	themeSettingHandler := pageBuilderHandlers.NewThemeSettingHandler(themeSettingUseCase)

	// Initialize permission checker
	permissionChecker := middleware.NewPermissionChecker(db)

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
		// Page Builder handlers
		pageHandler,
		blockHandler,
		pageBlockHandler,
		pageVersionHandler,
		themeSettingHandler,
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
