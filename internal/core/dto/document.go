package dto

import "github.com/google/uuid"

// DocumentFilter represents search and pagination parameters for documents
type DocumentFilter struct {
	SearchTerm   string `form:"search_term" json:"search_term"`
	EntityType   string `form:"entity_type" json:"entity_type"`
	EntityID     *uint  `form:"entity_id" json:"entity_id"`
	DocumentType string `form:"document_type" json:"document_type"`
	UploadedBy   *uint  `form:"uploaded_by" json:"uploaded_by"`
	SortBy       string `form:"sort_by" json:"sort_by"`
	SortDir      string `form:"sort_dir" json:"sort_dir"`
	Page         int    `form:"page" json:"page" binding:"min=1"`
	PageSize     int    `form:"page_size" json:"page_size" binding:"min=1,max=100"`
}

// DocumentUploadRequest contains info for document upload
type DocumentUploadRequest struct {
	EntityType   string `form:"entity_type" binding:"required"`
	EntityID     uint   `form:"entity_id" binding:"required"`
	DocumentName string `form:"document_name" binding:"required"`
}

// DocumentUpdateRequest contains fields that can be updated
type DocumentUpdateRequest struct {
	DocumentName string `json:"document_name" binding:"required"`
}

// DocumentPermissionRequest for assigning permissions
type DocumentPermissionRequest struct {
	DocumentID      uint       `json:"document_id" binding:"required"`
	UserID          *uuid.UUID `json:"user_id"`
	JobTitle        string     `json:"job_title"`
	PermissionLevel string     `json:"permission_level" binding:"required,oneof=view comment edit owner"`
}

// DocumentCommentRequest for adding comments
type DocumentCommentRequest struct {
	DocumentID uint   `json:"document_id" binding:"required"`
	Comment    string `json:"comment" binding:"required"`
}

// DocumentResponse standard response with permissions
type DocumentResponse struct {
	ID             uint      `json:"id"`
	DocumentCode   string    `json:"document_code"`
	EntityType     string    `json:"entity_type"`
	EntityID       uint      `json:"entity_id"`
	DocumentName   string    `json:"document_name"`
	DocumentType   string    `json:"document_type"`
	FileSize       int64     `json:"file_size"`
	UploadedBy     uuid.UUID `json:"uploaded_by"`
	UploaderName   string    `json:"uploader_name"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      string    `json:"updated_at"`
	UserPermission string    `json:"user_permission"` // Current user's permission level
}

// PaginatedDocumentsResponse for list responses
type PaginatedDocumentsResponse struct {
	Data       []DocumentResponse `json:"data"`
	TotalCount int                `json:"total_count"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}
