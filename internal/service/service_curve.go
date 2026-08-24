package service

import (
	"context"
	"fmt"

	"task219-colorreview/internal/batch"
	"task219-colorreview/internal/model"
)

// SaveBathCurve 保存浴液曲线（校验时间轴单调）。
func (s *Service) SaveBathCurve(ctx context.Context, c *model.BathCurve) (*model.BathCurve, error) {
	b, err := s.store.GetBatch(ctx, c.BatchID)
	if err != nil {
		return nil, err
	}
	if batch.IsSealed(b) {
		return nil, model.ErrBatchSealed
	}
	if c.Channel == "" {
		return nil, fmt.Errorf("%w: channel required", model.ErrInvalidArgument)
	}
	// 校验采样点时间轴严格单调，防止时间倒置。
	for i := 1; i < len(c.Points); i++ {
		if !c.Points[i].Timestamp.After(c.Points[i-1].Timestamp) {
			return nil, model.ErrTimeInverted
		}
	}
	return s.store.SaveBathCurve(ctx, c)
}

// GetBathCurve 查询某通道浴液曲线。
func (s *Service) GetBathCurve(ctx context.Context, batchID int64, channel string) (*model.BathCurve, error) {
	return s.store.GetBathCurve(ctx, batchID, channel)
}

// ListBathCurves 列出批次全部浴液曲线。
func (s *Service) ListBathCurves(ctx context.Context, batchID int64) ([]*model.BathCurve, error) {
	return s.store.ListBathCurves(ctx, batchID)
}
