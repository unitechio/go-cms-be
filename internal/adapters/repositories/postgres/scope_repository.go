package postgres

import (
	"context"

	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"github.com/owner/go-cms/pkg/pagination"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type scopeRepository struct {
	db *gorm.DB
}

// NewScopeRepository creates a new scope repository
func NewScopeRepository(db *gorm.DB) repositories.ScopeRepository {
	return &scopeRepository{db: db}
}

func (r *scopeRepository) Create(ctx context.Context, scope *domain.Scope) error {
	if err := r.db.WithContext(ctx).Create(scope).Error; err != nil {
		logger.Error("Failed to create scope", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to create scope", 500)
	}
	return nil
}

func (r *scopeRepository) GetByID(ctx context.Context, id uint) (*domain.Scope, error) {
	var scope domain.Scope
	if err := r.db.WithContext(ctx).First(&scope, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "scope not found", 404)
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get scope", 500)
	}
	return &scope, nil
}

func (r *scopeRepository) GetByCode(ctx context.Context, code string) (*domain.Scope, error) {
	var scope domain.Scope
	if err := r.db.WithContext(ctx).
		Where("code = ?", code).
		First(&scope).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "scope not found", 404)
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get scope", 500)
	}
	return &scope, nil
}

func (r *scopeRepository) List(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Scope], error) {
	var scopes []domain.Scope
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Scope{})

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to count scopes", 500)
	}

	// Apply pagination
	offset := (params.Page - 1) * params.Limit
	if err := query.Order("priority DESC").
		Offset(offset).
		Limit(params.Limit).
		Find(&scopes).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list scopes", 500)
	}

	return &pagination.Result[domain.Scope]{
		Data:  scopes,
		Total: total,
		Page:  params.Page,
		Limit: params.Limit,
	}, nil
}

func (r *scopeRepository) Update(ctx context.Context, scope *domain.Scope) error {
	if err := r.db.WithContext(ctx).Save(scope).Error; err != nil {
		logger.Error("Failed to update scope", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update scope", 500)
	}
	return nil
}

func (r *scopeRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Scope{}, id).Error; err != nil {
		logger.Error("Failed to delete scope", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete scope", 500)
	}
	return nil
}

func (r *scopeRepository) ListAll(ctx context.Context) ([]domain.Scope, error) {
	var scopes []domain.Scope
	if err := r.db.WithContext(ctx).
		Order("priority DESC").
		Find(&scopes).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list all scopes", 500)
	}
	return scopes, nil
}
