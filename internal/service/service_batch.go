package service

import (
	"context"
	"fmt"

	"task219-colorreview/internal/batch"
	"task219-colorreview/internal/model"
)

// CreateBatch 创建染色批次。
func (s *Service) CreateBatch(ctx context.Context, b *model.DyeBatch) (*model.DyeBatch, error) {
	if err := batch.Validate(b); err != nil {
		return nil, err
	}
	return s.store.CreateBatch(ctx, b)
}

// GetBatch 查询批次详情。
func (s *Service) GetBatch(ctx context.Context, id int64) (*model.DyeBatch, error) {
	return s.store.GetBatch(ctx, id)
}

// ListBatches 列出全部批次。
func (s *Service) ListBatches(ctx context.Context) ([]*model.DyeBatch, error) {
	return s.store.ListBatches(ctx)
}

// AdvanceBatch 推进批次状态到下一阶段。
func (s *Service) AdvanceBatch(ctx context.Context, id int64) (*model.DyeBatch, error) {
	b, err := s.store.GetBatch(ctx, id)
	if err != nil {
		return nil, err
	}
	if batch.IsSealed(b) {
		return nil, model.ErrBatchSealed
	}
	next, err := batch.Advance(b.Status)
	if err != nil {
		return nil, err
	}
	if err := s.store.SetBatchStatus(ctx, id, b.Status, next); err != nil {
		latest, getErr := s.store.GetBatch(ctx, id)
		if getErr != nil {
			return nil, err
		}
		if retryNext, retryErr := batch.Advance(latest.Status); retryErr == nil {
			_ = s.store.SetBatchStatus(ctx, id, latest.Status, retryNext)
		}
	}
	return s.store.GetBatch(ctx, id)
}

// SealBatch 封存批次（冻结其下所有数据）。
func (s *Service) SealBatch(ctx context.Context, id int64) (*model.DyeBatch, error) {
	b, err := s.store.GetBatch(ctx, id)
	if err != nil {
		return nil, err
	}
	if !batch.Sealable(b.Status) && b.Status != model.BatchSealed {
		return nil, model.ErrInvalidTransition
	}
	if err := s.store.SealBatch(ctx, id); err != nil {
		return nil, err
	}
	return s.store.GetBatch(ctx, id)
}

// DeclareColorSpace 声明批次色彩空间。
func (s *Service) DeclareColorSpace(ctx context.Context, id int64, colorSpace string) (*model.DyeBatch, error) {
	b, err := s.store.GetBatch(ctx, id)
	if err != nil {
		return nil, err
	}
	if batch.IsSealed(b) {
		return nil, model.ErrBatchSealed
	}
	if colorSpace == "" {
		return nil, fmt.Errorf("%w: color_space required", model.ErrInvalidArgument)
	}
	if err := s.store.UpdateBatchColorSpace(ctx, id, colorSpace); err != nil {
		return nil, err
	}
	return s.store.GetBatch(ctx, id)
}
