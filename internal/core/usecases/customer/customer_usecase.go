package customer

import (
	"context"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"github.com/owner/go-cms/pkg/pagination"
	"go.uber.org/zap"
)

// UseCase defines the customer use case interface
type UseCase interface {
	// Customer CRUD
	CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*domain.Customer, error)
	GetCustomer(ctx context.Context, id uint) (*domain.Customer, error)
	UpdateCustomer(ctx context.Context, id uint, req UpdateCustomerRequest) (*domain.Customer, error)
	DeleteCustomer(ctx context.Context, id uint) error
	ListCustomers(ctx context.Context, filter repositories.CustomerFilter, page *pagination.OffsetPagination) ([]*domain.Customer, int64, error)

	// Customer Assignment
	AssignToUser(ctx context.Context, customerID uint, userID uuid.UUID) error
	UnassignFromUser(ctx context.Context, customerID uint) error
	GetCustomersByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Customer, error)

	// Customer Status
	UpdateStatus(ctx context.Context, customerID uint, status domain.UserStatus) error

	// Search
	SearchCustomers(ctx context.Context, query string) ([]*domain.Customer, error)
}

// useCase implements the UseCase interface
type useCase struct {
	customerRepo repositories.CustomerRepository
	userRepo     repositories.UserRepository
}

// NewUseCase creates a new customer use case
func NewUseCase(
	customerRepo repositories.CustomerRepository,
	userRepo repositories.UserRepository,
) UseCase {
	return &useCase{
		customerRepo: customerRepo,
		userRepo:     userRepo,
	}
}

// CreateCustomerRequest represents a create customer request
type CreateCustomerRequest struct {
	Email      string     `json:"email" binding:"required,email"`
	FirstName  string     `json:"first_name" binding:"required"`
	LastName   string     `json:"last_name" binding:"required"`
	Phone      string     `json:"phone"`
	Company    string     `json:"company"`
	Address    string     `json:"address"`
	City       string     `json:"city"`
	State      string     `json:"state"`
	Country    string     `json:"country"`
	PostalCode string     `json:"postal_code"`
	Notes      string     `json:"notes"`
	Tags       string     `json:"tags"`
	Source     string     `json:"source"`
	AssignedTo *uuid.UUID `json:"assigned_to"`
}

// UpdateCustomerRequest represents an update customer request
type UpdateCustomerRequest struct {
	Email      *string `json:"email"`
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
	Phone      *string `json:"phone"`
	Company    *string `json:"company"`
	Address    *string `json:"address"`
	City       *string `json:"city"`
	State      *string `json:"state"`
	Country    *string `json:"country"`
	PostalCode *string `json:"postal_code"`
	Status     *string `json:"status"`
	Notes      *string `json:"notes"`
	Tags       *string `json:"tags"`
	Source     *string `json:"source"`
}

// CreateCustomer creates a new customer
func (uc *useCase) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*domain.Customer, error) {
	// Check if email already exists
	existingCustomer, err := uc.customerRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingCustomer != nil {
		logger.Warn("Customer with email already exists", zap.String("email", req.Email))
		return nil, errors.New(errors.ErrCodeConflict, "customer with this email already exists", 409)
	}

	// Validate assigned user if provided
	if req.AssignedTo != nil {
		_, err := uc.userRepo.GetByID(ctx, *req.AssignedTo)
		if err != nil {
			logger.Error("Assigned user not found", zap.Error(err), zap.String("user_id", req.AssignedTo.String()))
			return nil, errors.Wrap(err, errors.ErrCodeNotFound, "assigned user not found", 404)
		}
	}

	// Create customer
	customer := &domain.Customer{
		Email:      req.Email,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Phone:      req.Phone,
		Company:    req.Company,
		Address:    req.Address,
		City:       req.City,
		State:      req.State,
		Country:    req.Country,
		PostalCode: req.PostalCode,
		Status:     domain.UserStatusActive,
		Notes:      req.Notes,
		Tags:       req.Tags,
		Source:     req.Source,
		AssignedTo: req.AssignedTo,
	}

	if err := uc.customerRepo.Create(ctx, customer); err != nil {
		logger.Error("Failed to create customer", zap.Error(err))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to create customer", 500)
	}

	logger.Info("Customer created successfully",
		zap.Uint("customer_id", customer.ID),
		zap.String("email", customer.Email))

	return customer, nil
}

// GetCustomer gets a customer by ID
func (uc *useCase) GetCustomer(ctx context.Context, id uint) (*domain.Customer, error) {
	customer, err := uc.customerRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Failed to get customer", zap.Error(err), zap.Uint("id", id))
		return nil, err
	}
	return customer, nil
}

// UpdateCustomer updates a customer
func (uc *useCase) UpdateCustomer(ctx context.Context, id uint, req UpdateCustomerRequest) (*domain.Customer, error) {
	// Get existing customer
	customer, err := uc.customerRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Customer not found", zap.Error(err), zap.Uint("id", id))
		return nil, err
	}

	// Check if email is being changed and if it's already taken
	if req.Email != nil && *req.Email != customer.Email {
		existingCustomer, err := uc.customerRepo.GetByEmail(ctx, *req.Email)
		if err == nil && existingCustomer != nil {
			logger.Warn("Customer with email already exists", zap.String("email", *req.Email))
			return nil, errors.New(errors.ErrCodeConflict, "customer with this email already exists", 409)
		}
		customer.Email = *req.Email
	}

	// Update fields if provided
	if req.FirstName != nil {
		customer.FirstName = *req.FirstName
	}

	if req.LastName != nil {
		customer.LastName = *req.LastName
	}

	if req.Phone != nil {
		customer.Phone = *req.Phone
	}

	if req.Company != nil {
		customer.Company = *req.Company
	}

	if req.Address != nil {
		customer.Address = *req.Address
	}

	if req.City != nil {
		customer.City = *req.City
	}

	if req.State != nil {
		customer.State = *req.State
	}

	if req.Country != nil {
		customer.Country = *req.Country
	}

	if req.PostalCode != nil {
		customer.PostalCode = *req.PostalCode
	}

	if req.Status != nil {
		customer.Status = domain.UserStatus(*req.Status)
	}

	if req.Notes != nil {
		customer.Notes = *req.Notes
	}

	if req.Tags != nil {
		customer.Tags = *req.Tags
	}

	if req.Source != nil {
		customer.Source = *req.Source
	}

	// Update customer
	if err := uc.customerRepo.Update(ctx, customer); err != nil {
		logger.Error("Failed to update customer", zap.Error(err), zap.Uint("id", id))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update customer", 500)
	}

	logger.Info("Customer updated successfully", zap.Uint("customer_id", customer.ID))

	return customer, nil
}

// DeleteCustomer deletes a customer
func (uc *useCase) DeleteCustomer(ctx context.Context, id uint) error {
	// Check if customer exists
	_, err := uc.customerRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Customer not found", zap.Error(err), zap.Uint("id", id))
		return err
	}

	if err := uc.customerRepo.Delete(ctx, id); err != nil {
		logger.Error("Failed to delete customer", zap.Error(err), zap.Uint("id", id))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete customer", 500)
	}

	logger.Info("Customer deleted successfully", zap.Uint("customer_id", id))

	return nil
}

// ListCustomers lists customers with filters and pagination
func (uc *useCase) ListCustomers(ctx context.Context, filter repositories.CustomerFilter, page *pagination.OffsetPagination) ([]*domain.Customer, int64, error) {
	customers, total, err := uc.customerRepo.List(ctx, filter, page)
	if err != nil {
		logger.Error("Failed to list customers", zap.Error(err))
		return nil, 0, err
	}
	return customers, total, nil
}

// AssignToUser assigns a customer to a user
func (uc *useCase) AssignToUser(ctx context.Context, customerID uint, userID uuid.UUID) error {
	// Validate customer exists
	_, err := uc.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		logger.Error("Customer not found", zap.Error(err), zap.Uint("customer_id", customerID))
		return err
	}

	// Validate user exists
	_, err = uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("User not found", zap.Error(err), zap.String("user_id", userID.String()))
		return err
	}

	if err := uc.customerRepo.AssignToUser(ctx, customerID, userID); err != nil {
		logger.Error("Failed to assign customer to user", zap.Error(err),
			zap.Uint("customer_id", customerID), zap.String("user_id", userID.String()))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to assign customer to user", 500)
	}

	logger.Info("Customer assigned to user successfully",
		zap.Uint("customer_id", customerID), zap.String("user_id", userID.String()))

	return nil
}

// UnassignFromUser unassigns a customer from a user
func (uc *useCase) UnassignFromUser(ctx context.Context, customerID uint) error {
	// Validate customer exists
	_, err := uc.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		logger.Error("Customer not found", zap.Error(err), zap.Uint("customer_id", customerID))
		return err
	}

	if err := uc.customerRepo.UnassignFromUser(ctx, customerID); err != nil {
		logger.Error("Failed to unassign customer from user", zap.Error(err), zap.Uint("customer_id", customerID))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to unassign customer from user", 500)
	}

	logger.Info("Customer unassigned from user successfully", zap.Uint("customer_id", customerID))

	return nil
}

// GetCustomersByUser gets all customers assigned to a user
func (uc *useCase) GetCustomersByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Customer, error) {
	customers, err := uc.customerRepo.GetByAssignedUser(ctx, userID)
	if err != nil {
		logger.Error("Failed to get customers by user", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, err
	}
	return customers, nil
}

// UpdateStatus updates the customer status
func (uc *useCase) UpdateStatus(ctx context.Context, customerID uint, status domain.UserStatus) error {
	// Validate customer exists
	_, err := uc.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		logger.Error("Customer not found", zap.Error(err), zap.Uint("customer_id", customerID))
		return err
	}

	if err := uc.customerRepo.UpdateStatus(ctx, customerID, status); err != nil {
		logger.Error("Failed to update customer status", zap.Error(err), zap.Uint("customer_id", customerID))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update customer status", 500)
	}

	logger.Info("Customer status updated successfully",
		zap.Uint("customer_id", customerID), zap.String("status", string(status)))

	return nil
}

// SearchCustomers searches customers by query
func (uc *useCase) SearchCustomers(ctx context.Context, query string) ([]*domain.Customer, error) {
	// Use List with search filter instead of non-existent Search method
	filter := repositories.CustomerFilter{
		Search: query,
	}
	customers, _, err := uc.customerRepo.List(ctx, filter, nil)
	if err != nil {
		logger.Error("Failed to search customers", zap.Error(err), zap.String("query", query))
		return nil, err
	}
	return customers, nil
}
