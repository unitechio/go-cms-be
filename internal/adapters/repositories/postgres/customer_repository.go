package postgres

import (
	"context"

	"strconv"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"github.com/owner/go-cms/pkg/pagination"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type customerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository creates a new customer repository
func NewCustomerRepository(db *gorm.DB) repositories.CustomerRepository {
	return &customerRepository{db: db}
}

// Create creates a new customer
func (r *customerRepository) Create(ctx context.Context, customer *domain.Customer) error {
	if err := r.db.WithContext(ctx).Create(customer).Error; err != nil {
		logger.Error("Failed to create customer", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to create customer", 500)
	}
	return nil
}

// GetByID retrieves a customer by ID
func (r *customerRepository) GetByID(ctx context.Context, id uint) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.WithContext(ctx).
		Preload("AssignedUser").
		First(&customer, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "customer not found", 404)
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get customer", 500)
	}
	return &customer, nil
}

// GetByEmail retrieves a customer by email
func (r *customerRepository) GetByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&customer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "customer not found", 404)
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get customer", 500)
	}
	return &customer, nil
}

// Update updates an existing customer
func (r *customerRepository) Update(ctx context.Context, customer *domain.Customer) error {
	if err := r.db.WithContext(ctx).Save(customer).Error; err != nil {
		logger.Error("Failed to update customer", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update customer", 500)
	}
	return nil
}

// Delete soft deletes a customer
func (r *customerRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Customer{}, id).Error; err != nil {
		logger.Error("Failed to delete customer", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete customer", 500)
	}
	return nil
}

// List retrieves all customers with pagination and filters
func (r *customerRepository) List(ctx context.Context, filter repositories.CustomerFilter, page *pagination.OffsetPagination) ([]*domain.Customer, int64, error) {
	var customers []*domain.Customer
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Customer{})

	// Apply filters
	if filter.Email != "" {
		query = query.Where("email = ?", filter.Email)
	}
	if filter.Phone != "" {
		query = query.Where("phone = ?", filter.Phone)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.AssignedTo != nil {
		query = query.Where("assigned_to = ?", *filter.AssignedTo)
	}
	if filter.Source != "" {
		query = query.Where("source = ?", filter.Source)
	}
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR company ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern)
	}
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to count customers", 500)
	}

	// Apply pagination
	if page != nil {
		query = query.Offset(page.GetOffset()).Limit(page.PerPage)
	}

	// Execute query
	err := query.
		Preload("AssignedUser").
		Order("created_at DESC").
		Find(&customers).Error
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list customers", 500)
	}

	return customers, total, nil
}

// ListWithCursor retrieves customers with cursor-based pagination
func (r *customerRepository) ListWithCursor(ctx context.Context, filter repositories.CustomerFilter, cursor *pagination.Cursor, limit int) ([]*domain.Customer, *pagination.Cursor, error) {
	var customers []*domain.Customer

	query := r.db.WithContext(ctx).Model(&domain.Customer{})

	// Apply filters (same as List)
	if filter.Email != "" {
		query = query.Where("email = ?", filter.Email)
	}
	if filter.Phone != "" {
		query = query.Where("phone = ?", filter.Phone)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.AssignedTo != nil {
		query = query.Where("assigned_to = ?", *filter.AssignedTo)
	}
	if filter.Source != "" {
		query = query.Where("source = ?", filter.Source)
	}
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR company ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// Apply cursor
	if cursor != nil && cursor.After != "" {
		// Convert cursor.After (string) to uint64 for ID comparison
		cursorID, err := strconv.ParseUint(cursor.After, 10, 64)
		if err != nil {
			return nil, nil, errors.Wrap(err, errors.ErrCodeBadRequest, "invalid cursor", 400)
		}
		query = query.Where("id > ?", cursorID)
	}

	// Execute query
	err := query.
		Preload("AssignedUser").
		Order("id ASC").
		Limit(limit + 1). // Fetch one extra to determine if there are more
		Find(&customers).Error
	if err != nil {
		return nil, nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list customers", 500)
	}

	// Build next cursor
	var nextCursor *pagination.Cursor
	if len(customers) > limit {
		customers = customers[:limit]
		lastID := customers[len(customers)-1].ID
		nextCursor = &pagination.Cursor{
			After: strconv.FormatUint(uint64(lastID), 10),
		}
	}

	return customers, nextCursor, nil
}

// AssignToUser assigns a customer to a user
func (r *customerRepository) AssignToUser(ctx context.Context, customerID uint, userID uuid.UUID) error {
	err := r.db.WithContext(ctx).
		Model(&domain.Customer{}).
		Where("id = ?", customerID).
		Update("assigned_to", userID).Error
	if err != nil {
		logger.Error("Failed to assign customer to user", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to assign customer", 500)
	}
	return nil
}

// UnassignFromUser removes user assignment from a customer
func (r *customerRepository) UnassignFromUser(ctx context.Context, customerID uint) error {
	err := r.db.WithContext(ctx).
		Model(&domain.Customer{}).
		Where("id = ?", customerID).
		Update("assigned_to", nil).Error
	if err != nil {
		logger.Error("Failed to unassign customer from user", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to unassign customer", 500)
	}
	return nil
}

// GetByAssignedUser retrieves customers assigned to a specific user
func (r *customerRepository) GetByAssignedUser(ctx context.Context, userID uuid.UUID) ([]*domain.Customer, error) {
	var customers []*domain.Customer
	err := r.db.WithContext(ctx).
		Where("assigned_to = ?", userID).
		Order("created_at DESC").
		Find(&customers).Error
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get customers by assigned user", 500)
	}
	return customers, nil
}

// UpdateStatus updates customer status
func (r *customerRepository) UpdateStatus(ctx context.Context, customerID uint, status domain.UserStatus) error {
	err := r.db.WithContext(ctx).
		Model(&domain.Customer{}).
		Where("id = ?", customerID).
		Update("status", status).Error
	if err != nil {
		logger.Error("Failed to update customer status", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update customer status", 500)
	}
	return nil
}
