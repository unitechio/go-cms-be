package audit

import (
	"context"
	"net/http"

	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/pagination"
)

// UseCase handles audit log business logic
type UseCase struct {
	repo repositories.AuditLogRepository
}

// NewUseCase creates a new audit log use case
func NewUseCase(repo repositories.AuditLogRepository) *UseCase {
	return &UseCase{
		repo: repo,
	}
}

// Create creates a new audit log entry
func (uc *UseCase) Create(ctx context.Context, log *domain.AuditLog) error {
	if log == nil {
		return errors.New(errors.ErrCodeValidation, "audit log cannot be nil", http.StatusBadRequest)
	}

	return uc.repo.Create(ctx, log)
}

// List retrieves audit logs with filters and pagination
func (uc *UseCase) List(ctx context.Context, filter repositories.AuditLogFilter, page *pagination.OffsetPagination) ([]*domain.AuditLog, int64, error) {
	return uc.repo.List(ctx, filter, page)
}

// ListWithCursor retrieves audit logs with cursor-based pagination
func (uc *UseCase) ListWithCursor(ctx context.Context, filter repositories.AuditLogFilter, cursor *pagination.Cursor, limit int) ([]*domain.AuditLog, *pagination.Cursor, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	return uc.repo.ListWithCursor(ctx, filter, cursor, limit)
}

// GetByID retrieves an audit log by ID
func (uc *UseCase) GetByID(ctx context.Context, id uint) (*domain.AuditLog, error) {
	if id == 0 {
		return nil, errors.New(errors.ErrCodeValidation, "invalid audit log ID", http.StatusBadRequest)
	}

	log, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeNotFound, "audit log not found", http.StatusNotFound)
	}

	return log, nil
}

// GetByUserID retrieves audit logs for a specific user
func (uc *UseCase) GetByUserID(ctx context.Context, userID uint, limit int) ([]*domain.AuditLog, error) {
	if userID == 0 {
		return nil, errors.New(errors.ErrCodeValidation, "invalid user ID", http.StatusBadRequest)
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	return uc.repo.GetByUserID(ctx, userID, limit)
}

// GetByResource retrieves audit logs for a specific resource
func (uc *UseCase) GetByResource(ctx context.Context, resource string, resourceID uint) ([]*domain.AuditLog, error) {
	if resource == "" {
		return nil, errors.New(errors.ErrCodeValidation, "resource cannot be empty", http.StatusBadRequest)
	}

	if resourceID == 0 {
		return nil, errors.New(errors.ErrCodeValidation, "invalid resource ID", http.StatusBadRequest)
	}

	return uc.repo.GetByResource(ctx, resource, resourceID)
}

// DeleteOlderThan deletes audit logs older than specified days
func (uc *UseCase) DeleteOlderThan(ctx context.Context, days int) error {
	if days <= 0 {
		return errors.New(errors.ErrCodeValidation, "days must be greater than 0", http.StatusBadRequest)
	}

	return uc.repo.DeleteOlderThan(ctx, days)
}
