package handlers

import (
	"github.com/owner/go-cms/internal/core/domain"
)

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	FirstName  string `json:"first_name" binding:"required"`
	LastName   string `json:"last_name" binding:"required"`
	Phone      string `json:"phone"`
	Department string `json:"department"` // Department name from frontend
	Position   string `json:"position"`
	Status     string `json:"status"`
	RoleIDs    []uint `json:"role_ids" binding:"required,min=1"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Email      *string `json:"email,omitempty"`
	Password   *string `json:"password,omitempty"`
	FirstName  *string `json:"first_name,omitempty"`
	LastName   *string `json:"last_name,omitempty"`
	Phone      *string `json:"phone,omitempty"`
	Department *string `json:"department,omitempty"` // Department name from frontend
	Position   *string `json:"position,omitempty"`
	Status     *string `json:"status,omitempty"`
	RoleIDs    []uint  `json:"role_ids,omitempty"`
}

// ToUser converts CreateUserRequest to domain.User
// departmentID should be resolved by looking up the department name
func (r *CreateUserRequest) ToUser(departmentID *uint) *domain.User {
	status := domain.UserStatusActive
	if r.Status != "" {
		status = domain.UserStatus(r.Status)
	}

	return &domain.User{
		Email:        r.Email,
		Password:     r.Password,
		FirstName:    r.FirstName,
		LastName:     r.LastName,
		Phone:        r.Phone,
		DepartmentID: departmentID,
		Position:     r.Position,
		Status:       status,
	}
}

// ToUser converts UpdateUserRequest to domain.User
func (r *UpdateUserRequest) ToUser(departmentID *uint) *domain.User {
	user := &domain.User{}
	if r.Email != nil {
		user.Email = *r.Email
	}
	if r.Password != nil {
		user.Password = *r.Password
	}
	if r.FirstName != nil {
		user.FirstName = *r.FirstName
	}
	if r.LastName != nil {
		user.LastName = *r.LastName
	}
	if r.Phone != nil {
		user.Phone = *r.Phone
	}
	if r.Position != nil {
		user.Position = *r.Position
	}
	if r.Status != nil {
		user.Status = domain.UserStatus(*r.Status)
	}
	user.DepartmentID = departmentID

	return user
}
