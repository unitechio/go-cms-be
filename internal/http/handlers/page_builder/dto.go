package page_builder

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
)

// Page DTOs
type CreatePageRequest struct {
	Title          string `json:"title" binding:"required"`
	Slug           string `json:"slug" binding:"required"`
	Template       string `json:"template"`
	Status         string `json:"status"`
	SeoTitle       string `json:"seo_title"`
	SeoDescription string `json:"seo_description"`
	OgImage        string `json:"og_image"`
}

type UpdatePageRequest struct {
	Title          string `json:"title"`
	Slug           string `json:"slug"`
	Template       string `json:"template"`
	Status         string `json:"status"`
	SeoTitle       string `json:"seo_title"`
	SeoDescription string `json:"seo_description"`
	OgImage        string `json:"og_image"`
}

type PageResponse struct {
	ID             uuid.UUID            `json:"id"`
	Title          string               `json:"title"`
	Slug           string               `json:"slug"`
	Template       string               `json:"template"`
	Status         string               `json:"status"`
	SeoTitle       string               `json:"seo_title"`
	SeoDescription string               `json:"seo_description"`
	OgImage        string               `json:"og_image"`
	Blocks         []*PageBlockResponse `json:"blocks,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
	PublishedAt    *time.Time           `json:"published_at,omitempty"`
}

// Block DTOs
type CreateBlockRequest struct {
	Name            string          `json:"name" binding:"required"`
	Type            string          `json:"type" binding:"required"`
	Category        string          `json:"category"`
	Schema          json.RawMessage `json:"schema"`
	PreviewTemplate string          `json:"preview_template"`
	IsGlobal        bool            `json:"is_global"`
}

type UpdateBlockRequest struct {
	Name            string          `json:"name"`
	Type            string          `json:"type"`
	Category        string          `json:"category"`
	Schema          json.RawMessage `json:"schema"`
	PreviewTemplate string          `json:"preview_template"`
	IsGlobal        bool            `json:"is_global"`
}

type BlockResponse struct {
	ID              uuid.UUID       `json:"id"`
	Name            string          `json:"name"`
	Type            string          `json:"type"`
	Category        string          `json:"category"`
	Schema          json.RawMessage `json:"schema"`
	PreviewTemplate string          `json:"preview_template"`
	IsGlobal        bool            `json:"is_global"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// PageBlock DTOs
type AddPageBlockRequest struct {
	BlockID       uuid.UUID       `json:"block_id" binding:"required"`
	ParentBlockID *uuid.UUID      `json:"parent_block_id"`
	Order         int             `json:"order"`
	Config        json.RawMessage `json:"config"`
	Language      string          `json:"language"`
}

type UpdatePageBlockRequest struct {
	Config json.RawMessage `json:"config"`
	Order  int             `json:"order"`
}

type ReorderBlocksRequest struct {
	Blocks []BlockOrder `json:"blocks" binding:"required"`
}

type BlockOrder struct {
	ID    uuid.UUID `json:"id" binding:"required"`
	Order int       `json:"order"`
}

type PageBlockResponse struct {
	ID            uuid.UUID            `json:"id"`
	PageID        uuid.UUID            `json:"page_id"`
	BlockID       uuid.UUID            `json:"block_id"`
	ParentBlockID *uuid.UUID           `json:"parent_block_id,omitempty"`
	Order         int                  `json:"order"`
	Config        json.RawMessage      `json:"config"`
	Language      string               `json:"language"`
	Block         *BlockResponse       `json:"block,omitempty"`
	Children      []*PageBlockResponse `json:"children,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
}

// PageVersion DTOs
type PageVersionResponse struct {
	ID                uuid.UUID       `json:"id"`
	PageID            uuid.UUID       `json:"page_id"`
	VersionNumber     int             `json:"version_number"`
	Title             string          `json:"title"`
	Slug              string          `json:"slug"`
	BlocksSnapshot    json.RawMessage `json:"blocks_snapshot"`
	CreatedBy         uuid.UUID       `json:"created_by"`
	ChangeDescription string          `json:"change_description"`
	CreatedAt         time.Time       `json:"created_at"`
}

// ThemeSetting DTOs
type UpdateThemeRequest struct {
	Config json.RawMessage `json:"config"`
	Name   string          `json:"name"`
}

type ThemeSettingResponse struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Config    json.RawMessage `json:"config"`
	IsActive  bool            `json:"is_active"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Helper functions to map domain to DTO
func ToPageResponse(page *domain.Page) *PageResponse {
	resp := &PageResponse{
		ID:             page.ID,
		Title:          page.Title,
		Slug:           page.Slug,
		Template:       page.Template,
		Status:         string(page.Status),
		SeoTitle:       page.SeoTitle,
		SeoDescription: page.SeoDescription,
		OgImage:        page.OgImage,
		CreatedAt:      page.CreatedAt,
		UpdatedAt:      page.UpdatedAt,
		PublishedAt:    page.PublishedAt,
	}

	if len(page.Blocks) > 0 {
		resp.Blocks = make([]*PageBlockResponse, len(page.Blocks))
		for i, block := range page.Blocks {
			resp.Blocks[i] = ToPageBlockResponse(&block)
		}
	}

	return resp
}

func ToBlockResponse(block *domain.Block) *BlockResponse {
	return &BlockResponse{
		ID:              block.ID,
		Name:            block.Name,
		Type:            block.Type,
		Category:        block.Category,
		Schema:          block.Schema,
		PreviewTemplate: block.PreviewTemplate,
		IsGlobal:        block.IsGlobal,
		CreatedAt:       block.CreatedAt,
		UpdatedAt:       block.UpdatedAt,
	}
}

func ToPageBlockResponse(pb *domain.PageBlock) *PageBlockResponse {
	resp := &PageBlockResponse{
		ID:            pb.ID,
		PageID:        pb.PageID,
		BlockID:       pb.BlockID,
		ParentBlockID: pb.ParentBlockID,
		Order:         pb.Order,
		Config:        pb.Config,
		Language:      pb.Language,
		CreatedAt:     pb.CreatedAt,
	}

	if pb.Block.ID != uuid.Nil {
		resp.Block = ToBlockResponse(&pb.Block)
	}

	if len(pb.Children) > 0 {
		resp.Children = make([]*PageBlockResponse, len(pb.Children))
		for i, child := range pb.Children {
			resp.Children[i] = ToPageBlockResponse(child)
		}
	}

	return resp
}

func ToPageVersionResponse(v *domain.PageVersion) *PageVersionResponse {
	return &PageVersionResponse{
		ID:                v.ID,
		PageID:            v.PageID,
		VersionNumber:     v.VersionNumber,
		Title:             v.Title,
		Slug:              v.Slug,
		BlocksSnapshot:    v.BlocksSnapshot,
		CreatedBy:         v.CreatedBy,
		ChangeDescription: v.ChangeDescription,
		CreatedAt:         v.CreatedAt,
	}
}

func ToThemeSettingResponse(t *domain.ThemeSetting) *ThemeSettingResponse {
	return &ThemeSettingResponse{
		ID:        t.ID,
		Name:      t.Name,
		Config:    t.Config,
		IsActive:  t.IsActive,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
