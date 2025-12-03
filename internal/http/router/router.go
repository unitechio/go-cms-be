package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/internal/config"
	"github.com/owner/go-cms/internal/core/usecases/audit"
	"github.com/owner/go-cms/internal/http/handlers"
	authHandlers "github.com/owner/go-cms/internal/http/handlers/authorization"
	pageBuilderHandlers "github.com/owner/go-cms/internal/http/handlers/page_builder"
	"github.com/owner/go-cms/internal/http/middleware"
	"github.com/owner/go-cms/pkg/response"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	config              *config.Config
	permissionChecker   *middleware.DefaultPermissionChecker
	authHandler         *handlers.AuthHandler
	userHandler         *handlers.UserHandler
	moduleHandler       *authHandlers.ModuleHandler
	departmentHandler   *authHandlers.DepartmentHandler
	serviceHandler      *authHandlers.ServiceHandler
	scopeHandler        *authHandlers.ScopeHandler
	roleHandler         *handlers.RoleHandler
	permissionHandler   *handlers.PermissionHandler
	notificationHandler *handlers.NotificationHandler
	websocketHandler    *handlers.WebSocketHandler
	auditLogHandler     *handlers.AuditLogHandler
	documentHandler     *handlers.DocumentHandler
	auditLogUseCase     *audit.UseCase
	categoryHandler     *handlers.CategoryHandler
	// Page Builder handlers
	pageHandler         *pageBuilderHandlers.PageHandler
	blockHandler        *pageBuilderHandlers.BlockHandler
	pageBlockHandler    *pageBuilderHandlers.PageBlockHandler
	pageVersionHandler  *pageBuilderHandlers.PageVersionHandler
	themeSettingHandler *pageBuilderHandlers.ThemeSettingHandler
}

func NewRouter(
	cfg *config.Config,
	permissionChecker *middleware.DefaultPermissionChecker,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	moduleHandler *authHandlers.ModuleHandler,
	departmentHandler *authHandlers.DepartmentHandler,
	serviceHandler *authHandlers.ServiceHandler,
	scopeHandler *authHandlers.ScopeHandler,
	roleHandler *handlers.RoleHandler,
	permissionHandler *handlers.PermissionHandler,
	notificationHandler *handlers.NotificationHandler,
	websocketHandler *handlers.WebSocketHandler,
	auditLogHandler *handlers.AuditLogHandler,
	documentHandler *handlers.DocumentHandler,
	auditLogUseCase *audit.UseCase,
	categoryHandler *handlers.CategoryHandler,
	// Page Builder handlers
	pageHandler *pageBuilderHandlers.PageHandler,
	blockHandler *pageBuilderHandlers.BlockHandler,
	pageBlockHandler *pageBuilderHandlers.PageBlockHandler,
	pageVersionHandler *pageBuilderHandlers.PageVersionHandler,
	themeSettingHandler *pageBuilderHandlers.ThemeSettingHandler,
) *Router {
	return &Router{
		config:              cfg,
		permissionChecker:   permissionChecker,
		authHandler:         authHandler,
		userHandler:         userHandler,
		moduleHandler:       moduleHandler,
		departmentHandler:   departmentHandler,
		serviceHandler:      serviceHandler,
		scopeHandler:        scopeHandler,
		roleHandler:         roleHandler,
		permissionHandler:   permissionHandler,
		notificationHandler: notificationHandler,
		websocketHandler:    websocketHandler,
		auditLogHandler:     auditLogHandler,
		documentHandler:     documentHandler,
		auditLogUseCase:     auditLogUseCase,
		categoryHandler:     categoryHandler,
		// Page Builder handlers
		pageHandler:         pageHandler,
		blockHandler:        blockHandler,
		pageBlockHandler:    pageBlockHandler,
		pageVersionHandler:  pageVersionHandler,
		themeSettingHandler: themeSettingHandler,
	}
}

func (r *Router) Setup() *gin.Engine {
	if r.config.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// Global middleware
	engine.Use(middleware.RecoveryMiddleware())
	engine.Use(middleware.LoggerMiddleware())
	engine.Use(middleware.CORSMiddleware(&r.config.CORS))
	engine.Use(middleware.TimeoutMiddleware(30 * time.Second))
	// Audit logging middleware - logs all requests
	engine.Use(middleware.AuditLogger(r.auditLogUseCase))

	engine.GET("/health", r.healthCheck)
	engine.GET("/ping", r.ping)

	// Swagger documentation
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := engine.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := v1.Group("/public")
		{
			public.GET("/health", r.healthCheck)
		}

		// Auth routes (no authentication required for login/register)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", r.authHandler.Register)
			auth.POST("/login", r.authHandler.Login)
			auth.POST("/refresh", r.authHandler.RefreshToken)
			auth.POST("/forgot-password", r.authHandler.ForgotPassword)
			auth.POST("/reset-password", r.authHandler.ResetPassword)
			auth.POST("/verify-email", r.authHandler.VerifyEmail)
			auth.POST("/resend-otp", r.authHandler.ResendOTP)
		}

		// Protected routes (authentication required)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(&r.config.JWT))
		{
			// Auth protected routes
			authProtected := protected.Group("/auth")
			{
				authProtected.POST("/logout", r.authHandler.Logout)
				authProtected.POST("/change-password", r.authHandler.ChangePassword)
				authProtected.GET("/me", r.authHandler.GetMe)
				authProtected.PUT("/me", r.authHandler.UpdateProfile)

				// 2FA routes
				authProtected.POST("/2fa/enable", r.authHandler.Enable2FA)
				authProtected.POST("/2fa/disable", r.authHandler.Disable2FA)
				authProtected.POST("/2fa/verify", r.authHandler.Verify2FA)
			}

			// User management routes
			users := protected.Group("/users")
			{
				users.GET("", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminUsersRead), r.userHandler.ListUsers)
				users.POST("", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminUsersCreate), r.userHandler.CreateUser)
				users.GET("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminUsersRead), r.userHandler.GetUser)
				users.PUT("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminUsersUpdate), r.userHandler.UpdateUser)
				users.DELETE("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminUsersDelete), r.userHandler.DeleteUser)

				// User roles
				users.GET("/:id/roles", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminUsersRead), r.userHandler.GetUserRoles)
				users.POST("/:id/roles", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminUsersUpdate), r.userHandler.AssignRole)
				users.DELETE("/:id/roles/:roleId", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminUsersUpdate), r.userHandler.RemoveRole)
			}

			// Customer management routes
			customers := protected.Group("/customers")
			{
				customers.GET("", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionCRMCustomersRead), r.placeholder("List Customers"))
				customers.POST("", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionCRMCustomersCreate), r.placeholder("Create Customer"))
				customers.GET("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionCRMCustomersRead), r.placeholder("Get Customer"))
				customers.PUT("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionCRMCustomersUpdate), r.placeholder("Update Customer"))
				customers.DELETE("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionCRMCustomersDelete), r.placeholder("Delete Customer"))
			}

			// Post management routes
			posts := protected.Group("/posts")
			{
				posts.GET("", r.placeholder("List Posts"))
				posts.POST("", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionContentPostsCreate), r.placeholder("Create Post"))
				posts.GET("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionContentPostsRead), r.placeholder("Get Post"))
				posts.PUT("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionContentPostsUpdate), r.placeholder("Update Post"))
				posts.DELETE("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionContentPostsDelete), r.placeholder("Delete Post"))
				posts.POST("/:id/publish", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionContentPostsPublish), r.placeholder("Publish Post"))
				posts.POST("/:id/schedule", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionContentPostsPublish), r.placeholder("Schedule Post"))
			}

			// Media management routes
			media := protected.Group("/media")
			{
				media.GET("", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionContentMediaRead), r.placeholder("List Media"))
				media.POST("/upload", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionContentMediaUpload), r.placeholder("Upload Media"))
				media.GET("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionContentMediaRead), r.placeholder("Get Media"))
				media.DELETE("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionContentMediaDelete), r.placeholder("Delete Media"))
				media.GET("/:id/presigned-url", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionContentMediaRead), r.placeholder("Get Presigned URL"))
			}

			// Category management routes
			categories := protected.Group("/categories")
			{
				categories.GET("/tree", r.categoryHandler.GetCategoryTree)
				categories.GET("/active", r.categoryHandler.GetActiveCategories)
				categories.GET("", r.categoryHandler.ListCategories)
				categories.POST("", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionContentPostsCreate), r.categoryHandler.CreateCategory)
				categories.GET("/:id", r.categoryHandler.GetCategory)
				categories.PUT("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionContentPostsUpdate), r.categoryHandler.UpdateCategory)
				categories.DELETE("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionContentPostsDelete), r.categoryHandler.DeleteCategory)
				categories.PUT("/:id/reorder", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionContentPostsUpdate), r.categoryHandler.ReorderCategory)
			}

			// Role management routes
			roles := protected.Group("/roles")
			{
				roles.GET("", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminRolesRead), r.roleHandler.ListRoles)
				roles.POST("", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminRolesCreate), r.roleHandler.CreateRole)
				roles.GET("/hierarchy", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminRolesRead), r.roleHandler.GetRoleHierarchy)
				roles.GET("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminRolesRead), r.roleHandler.GetRole)
				roles.PUT("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminRolesUpdate), r.roleHandler.UpdateRole)
				roles.DELETE("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminRolesDelete), r.roleHandler.DeleteRole)

				// Role permissions
				roles.GET("/:id/permissions", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminRolesRead), r.roleHandler.GetRolePermissions)
				roles.POST("/:id/permissions", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminPermissionsManage), r.roleHandler.AssignPermission)
				roles.DELETE("/:id/permissions/:permissionId", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminPermissionsManage), r.roleHandler.RemovePermission)
			}

			// Permission management routes
			permissions := protected.Group("/permissions")
			{
				permissions.GET("", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminPermissionsManage), r.permissionHandler.ListPermissions)
				permissions.POST("", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminPermissionsManage), r.permissionHandler.CreatePermission)
				permissions.GET("/module/:module", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminPermissionsManage), r.permissionHandler.GetPermissionsByModule)
				permissions.GET("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminPermissionsManage), r.permissionHandler.GetPermission)
				permissions.PUT("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminPermissionsManage), r.permissionHandler.UpdatePermission)
				permissions.DELETE("/:id", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminPermissionsManage), r.permissionHandler.DeletePermission)
			}

			// Notification routes
			notifications := protected.Group("/notifications")
			{
				// User notification routes
				notifications.GET("/me", r.notificationHandler.GetMyNotifications)
				notifications.GET("/unread-count", r.notificationHandler.GetUnreadCount)
				notifications.GET("/stats", r.notificationHandler.GetStats)
				notifications.POST("/mark-all-read", r.notificationHandler.MarkAllAsRead)
				notifications.DELETE("/me", r.notificationHandler.DeleteAllMyNotifications)

				// Individual notification routes
				notifications.GET("/:id", r.notificationHandler.GetNotification)
				notifications.PUT("/:id", r.notificationHandler.UpdateNotification)
				notifications.DELETE("/:id", r.notificationHandler.DeleteNotification)
				notifications.POST("/:id/read", r.notificationHandler.MarkAsRead)
				notifications.POST("/:id/unread", r.notificationHandler.MarkAsUnread)

				// Admin routes
				notifications.GET("", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminUsersRead), r.notificationHandler.GetAllNotifications)
				notifications.POST("", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminUsersRead), r.notificationHandler.CreateNotification)
				notifications.POST("/broadcast", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminUsersRead), r.notificationHandler.BroadcastNotification)
			}
			document := protected.Group("/documents")
			{
				// Document CRUD
				document.POST("/upload", r.documentHandler.UploadDocument)
				document.GET("/list", r.documentHandler.GetDocuments)
				document.GET("/entity/:type/:id", r.documentHandler.GetDocumentsByEntity)
				document.GET("/:id", r.documentHandler.GetDocumentByID)
				document.GET("/code/:code", r.documentHandler.GetDocumentByCode)
				document.GET("/view/:id", r.documentHandler.ViewDocument)
				document.GET("/view-url/:id", r.documentHandler.GetDocumentViewURL)
				document.PUT("/:id", r.documentHandler.UpdateDocument)
				document.DELETE("/:id", r.documentHandler.DeleteDocument)
				document.GET("/download/:id", r.documentHandler.DownloadDocument)

				// Permissions
				document.POST("/permissions", r.documentHandler.AddDocumentPermission)
				document.GET("/:id/permissions", r.documentHandler.GetDocumentPermissions)
				document.PUT("/permissions/:id", r.documentHandler.UpdateDocumentPermission)
				document.DELETE("/permissions/:id", r.documentHandler.DeleteDocumentPermission)

				// Comments
				document.POST("/comments", r.documentHandler.AddDocumentComment)
				document.GET("/:id/comments", r.documentHandler.GetDocumentComments)
				document.PUT("/comments/:id", r.documentHandler.UpdateDocumentComment)
				document.DELETE("/comments/:id", r.documentHandler.DeleteDocumentComment)

				// Versions
				document.GET("/:id/versions", r.documentHandler.GetDocumentVersions)
			}

			// WebSocket routes
			ws := protected.Group("/ws")
			{
				ws.GET("", r.websocketHandler.HandleWebSocket)
				ws.GET("/online-users", r.websocketHandler.GetOnlineUsers)
				ws.GET("/stats", r.websocketHandler.GetConnectionStats)
				ws.POST("/broadcast", middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminUsersRead), r.websocketHandler.BroadcastMessage)
			}

			// Audit log routes (admin only)
			auditLogs := protected.Group("/audit-logs")
			auditLogs.Use(middleware.AuthorizeMiddleware(r.permissionChecker, middleware.PermissionAdminUsersRead))
			{
				auditLogs.GET("", r.auditLogHandler.ListAuditLogs)
				auditLogs.GET("/:id", r.auditLogHandler.GetAuditLog)
				auditLogs.GET("/user/:user_id", r.auditLogHandler.GetUserAuditLogs)
				auditLogs.GET("/resource", r.auditLogHandler.GetResourceAuditLogs)
				auditLogs.DELETE("/cleanup", r.auditLogHandler.CleanupOldLogs)
			}

			// Module routes
			modules := protected.Group("/modules")
			{
				modules.POST("", r.moduleHandler.CreateModule)
				modules.GET("", r.moduleHandler.ListModules)
				modules.GET("/active", r.moduleHandler.ListActiveModules)
				modules.GET("/:id", r.moduleHandler.GetModule)
				modules.GET("/code/:code", r.moduleHandler.GetModuleByCode)
				modules.PUT("/:id", r.moduleHandler.UpdateModule)
				modules.DELETE("/:id", r.moduleHandler.DeleteModule)
			}

			// Department routes
			departments := protected.Group("/departments")
			{
				departments.POST("", r.departmentHandler.CreateDepartment)
				departments.GET("", r.departmentHandler.ListDepartments)
				departments.GET("/active", r.departmentHandler.ListActiveDepartments)
				departments.GET("/:id", r.departmentHandler.GetDepartment)
				departments.GET("/code/:code", r.departmentHandler.GetDepartmentByCode)
				departments.PUT("/:id", r.departmentHandler.UpdateDepartment)
				departments.DELETE("/:id", r.departmentHandler.DeleteDepartment)
			}

			// Module departments route
			protected.GET("/modules/:id/departments", r.departmentHandler.ListDepartmentsByModule)

			// Service routes
			services := protected.Group("/services")
			{
				services.POST("", r.serviceHandler.CreateService)
				services.GET("", r.serviceHandler.ListServices)
				services.GET("/active", r.serviceHandler.ListActiveServices)
				services.GET("/:id", r.serviceHandler.GetService)
				services.GET("/code/:code", r.serviceHandler.GetServiceByCode)
				services.PUT("/:id", r.serviceHandler.UpdateService)
				services.DELETE("/:id", r.serviceHandler.DeleteService)
			}

			// Department services route
			protected.GET("/departments/:id/services", r.serviceHandler.ListServicesByDepartment)

			// Scope routes
			scopes := protected.Group("/scopes")
			{
				scopes.POST("", r.scopeHandler.CreateScope)
				scopes.GET("", r.scopeHandler.ListScopes)
				scopes.GET("/all", r.scopeHandler.ListAllScopes)
				scopes.GET("/:id", r.scopeHandler.GetScope)
				scopes.GET("/code/:code", r.scopeHandler.GetScopeByCode)
				scopes.PUT("/:id", r.scopeHandler.UpdateScope)
				scopes.DELETE("/:id", r.scopeHandler.DeleteScope)
			}

			// Page Builder routes
			pages := protected.Group("/pages")
			{
				pages.GET("", r.pageHandler.ListPages)
				pages.POST("", r.pageHandler.CreatePage)
				pages.GET("/:id", r.pageHandler.GetPage)
				pages.PUT("/:id", r.pageHandler.UpdatePage)
				pages.DELETE("/:id", r.pageHandler.DeletePage)
				pages.POST("/:id/duplicate", r.pageHandler.DuplicatePage)
				pages.POST("/:id/publish", r.pageHandler.PublishPage)

				// Page blocks
				pages.POST("/:id/blocks", r.pageBlockHandler.AddBlockToPage)
				pages.PUT("/:id/blocks/:blockId", r.pageBlockHandler.UpdatePageBlock)
				pages.DELETE("/:id/blocks/:blockId", r.pageBlockHandler.RemoveBlockFromPage)
				pages.PUT("/:id/blocks/reorder", r.pageBlockHandler.ReorderBlocks)

				// Page versions
				pages.GET("/:id/versions", r.pageVersionHandler.GetVersionHistory)
				pages.POST("/:id/versions/:versionId/revert", r.pageVersionHandler.RevertToVersion)
			}

			// Block routes
			blocks := protected.Group("/blocks")
			{
				blocks.GET("", r.blockHandler.ListBlocks)
				blocks.POST("", r.blockHandler.CreateBlock)
				blocks.GET("/:id", r.blockHandler.GetBlock)
				blocks.PUT("/:id", r.blockHandler.UpdateBlock)
				blocks.DELETE("/:id", r.blockHandler.DeleteBlock)
			}

			// Theme settings routes
			themeSettings := protected.Group("/theme-settings")
			{
				themeSettings.GET("", r.themeSettingHandler.GetAllThemes)
				themeSettings.GET("/active", r.themeSettingHandler.GetActiveTheme)
				themeSettings.PUT("/:id", r.themeSettingHandler.UpdateTheme)
				themeSettings.POST("/:id/activate", r.themeSettingHandler.ActivateTheme)
			}
		}
	}

	return engine
}

// healthCheck handles health check requests
// @Summary Health check
// @Description Check if the service is healthy
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /health [get]
func (r *Router) healthCheck(c *gin.Context) {
	response.Success(c, gin.H{
		"status":  "healthy",
		"service": "go-cms",
		"version": "1.0.0",
	})
}

// ping handles ping requests
// @Summary Ping
// @Description Ping the service
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /ping [get]
func (r *Router) ping(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "pong",
	})
}

// placeholder is a temporary handler for routes that are not yet implemented
func (r *Router) placeholder(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Success(c, gin.H{
			"message": "Endpoint: " + name + " - Coming soon",
			"status":  "not_implemented",
		})
	}
}
