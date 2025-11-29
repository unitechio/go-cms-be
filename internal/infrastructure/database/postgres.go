package database

import (
	"fmt"
	"time"

	"github.com/owner/go-cms/internal/config"
	"github.com/owner/go-cms/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DB is the global database instance
var DB *gorm.DB

// Init initializes the database connection
func Init(cfg *config.DatabaseConfig) error {
	var err error

	// Configure GORM logger
	gormLog := gormlogger.Default.LogMode(gormlogger.Info)

	// Connect to database
	DB, err = gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger:                                   gormLog,
		DisableForeignKeyConstraintWhenMigrating: false,
		PrepareStmt:                              true,
		SkipDefaultTransaction:                   true, // Better performance
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL database
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.DBName),
	)

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// Transaction executes a function within a database transaction
func Transaction(fn func(*gorm.DB) error) error {
	return DB.Transaction(fn)
}

// WithContext returns a new DB instance with context
func WithContext(db *gorm.DB) *gorm.DB {
	return db.Session(&gorm.Session{})
}

// Paginate is a GORM scope for pagination
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// CursorPaginate is a GORM scope for cursor-based pagination
func CursorPaginate(cursorID uint, cursorTime time.Time, limit int, direction string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if limit <= 0 {
			limit = 20
		}

		// Add one extra to check if there are more results
		db = db.Limit(limit + 1)

		if cursorID > 0 {
			if direction == "prev" {
				// For previous page, reverse the order
				db = db.Where("(created_at > ? OR (created_at = ? AND id > ?))", cursorTime, cursorTime, cursorID)
				db = db.Order("created_at ASC, id ASC")
			} else {
				// For next page (default)
				db = db.Where("(created_at < ? OR (created_at = ? AND id < ?))", cursorTime, cursorTime, cursorID)
				db = db.Order("created_at DESC, id DESC")
			}
		} else {
			// First page
			db = db.Order("created_at DESC, id DESC")
		}

		return db
	}
}

// CreateIndexes creates database indexes for optimization
func CreateIndexes() error {
	// This will be called after migrations
	// Add composite indexes here
	indexes := []string{
		// Users table indexes
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_status ON users(status)",
		"CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_users_email_status ON users(email, status)",

		// Customers table indexes
		"CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email)",
		"CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone)",
		"CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status)",
		"CREATE INDEX IF NOT EXISTS idx_customers_created_at ON customers(created_at DESC)",

		// Posts table indexes
		"CREATE INDEX IF NOT EXISTS idx_posts_status ON posts(status)",
		"CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id)",
		"CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_posts_published_at ON posts(published_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_posts_status_published ON posts(status, published_at DESC)",

		// Media table indexes
		"CREATE INDEX IF NOT EXISTS idx_media_type ON media(type)",
		"CREATE INDEX IF NOT EXISTS idx_media_uploaded_by ON media(uploaded_by)",
		"CREATE INDEX IF NOT EXISTS idx_media_created_at ON media(created_at DESC)",

		// Roles table indexes
		"CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name)",
		"CREATE INDEX IF NOT EXISTS idx_roles_level ON roles(level)",

		// Permissions table indexes
		"CREATE INDEX IF NOT EXISTS idx_permissions_resource ON permissions(resource)",
		"CREATE INDEX IF NOT EXISTS idx_permissions_action ON permissions(action)",

		// User roles table indexes
		"CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id)",

		// Role permissions table indexes
		"CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id)",
		"CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id)",

		// Audit logs table indexes
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC)",
	}

	for _, index := range indexes {
		if err := DB.Exec(index).Error; err != nil {
			logger.Error("Failed to create index", zap.String("index", index), zap.Error(err))
			return err
		}
	}

	logger.Info("Database indexes created successfully")
	return nil
}

// CreatePartitions creates table partitions for large tables
func CreatePartitions() error {
	// Partition audit_logs table by month
	partitionSQL := `
		-- Convert audit_logs to partitioned table if not already
		DO $$
		BEGIN
			-- Check if table is already partitioned
			IF NOT EXISTS (
				SELECT 1 FROM pg_partitioned_table WHERE partrelid = 'audit_logs'::regclass
			) THEN
				-- Create new partitioned table
				CREATE TABLE IF NOT EXISTS audit_logs_partitioned (
					LIKE audit_logs INCLUDING ALL
				) PARTITION BY RANGE (created_at);

				-- Create partitions for current and next 12 months
				FOR i IN 0..12 LOOP
					EXECUTE format(
						'CREATE TABLE IF NOT EXISTS audit_logs_%s PARTITION OF audit_logs_partitioned
						FOR VALUES FROM (%L) TO (%L)',
						to_char(CURRENT_DATE + (i || ' months')::interval, 'YYYY_MM'),
						date_trunc('month', CURRENT_DATE + (i || ' months')::interval),
						date_trunc('month', CURRENT_DATE + ((i + 1) || ' months')::interval)
					);
				END LOOP;
			END IF;
		END $$;
	`

	if err := DB.Exec(partitionSQL).Error; err != nil {
		logger.Error("Failed to create partitions", zap.Error(err))
		return err
	}

	logger.Info("Database partitions created successfully")
	return nil
}
