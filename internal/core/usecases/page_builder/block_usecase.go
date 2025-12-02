package page_builder

import (
	"context"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/pagination"
)

type BlockUseCase struct {
	blockRepo repositories.BlockRepository
}

func NewBlockUseCase(blockRepo repositories.BlockRepository) *BlockUseCase {
	return &BlockUseCase{
		blockRepo: blockRepo,
	}
}

func (uc *BlockUseCase) CreateBlock(ctx context.Context, block *domain.Block) error {
	return uc.blockRepo.Create(ctx, block)
}

func (uc *BlockUseCase) GetBlock(ctx context.Context, id uuid.UUID) (*domain.Block, error) {
	return uc.blockRepo.GetByID(ctx, id)
}

func (uc *BlockUseCase) UpdateBlock(ctx context.Context, block *domain.Block) error {
	existingBlock, err := uc.blockRepo.GetByID(ctx, block.ID)
	if err != nil {
		return err
	}

	existingBlock.Name = block.Name
	existingBlock.Type = block.Type
	existingBlock.Category = block.Category
	existingBlock.Schema = block.Schema
	existingBlock.PreviewTemplate = block.PreviewTemplate
	existingBlock.IsGlobal = block.IsGlobal

	return uc.blockRepo.Update(ctx, existingBlock)
}

func (uc *BlockUseCase) DeleteBlock(ctx context.Context, id uuid.UUID) error {
	return uc.blockRepo.Delete(ctx, id)
}

func (uc *BlockUseCase) ListBlocks(ctx context.Context, filter repositories.BlockFilter, page *pagination.OffsetPagination) ([]*domain.Block, int64, error) {
	return uc.blockRepo.List(ctx, filter, page)
}
