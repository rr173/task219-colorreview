package service

import (
	"context"
	"fmt"

	"task219-colorreview/internal/batch"
	"task219-colorreview/internal/model"
	"task219-colorreview/internal/sampling"
)

// AddMeasurePoint 上传测色点（带仪器校准修正）。
func (s *Service) AddMeasurePoint(ctx context.Context, m *model.MeasurePoint, instrumentID string) (*model.MeasurePoint, error) {
	if err := sampling.ValidatePoint(m); err != nil {
		return nil, err
	}
	b, err := s.store.GetBatch(ctx, m.BatchID)
	if err != nil {
		return nil, err
	}
	if batch.IsSealed(b) {
		return nil, model.ErrBatchSealed
	}
	// 若指定仪器，用最近校准偏移修正 Lab 值。
	if instrumentID != "" {
		cal, err := s.store.LatestCalibration(ctx, instrumentID)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", model.ErrCalibrationMissing, err)
		}
		m.L, m.A, m.B = sampling.ApplyCalibration(m.L, m.A, m.B, cal)
	}
	m.Status = model.MeasurePending
	return s.store.CreateMeasurePoint(ctx, m)
}

// ListMeasurePoints 列出批次测色点。
func (s *Service) ListMeasurePoints(ctx context.Context, batchID int64) ([]*model.MeasurePoint, error) {
	return s.store.ListMeasurePoints(ctx, batchID)
}

// RejectMeasurePoint 剔除测色点（记录原因）。
func (s *Service) RejectMeasurePoint(ctx context.Context, batchID, id int64, reason string) (*model.MeasurePoint, error) {
	b, err := s.store.GetBatch(ctx, batchID)
	if err != nil {
		return nil, err
	}
	if batch.IsSealed(b) {
		return nil, model.ErrBatchSealed
	}
	pt, err := s.store.GetMeasurePoint(ctx, id)
	if err != nil {
		return nil, err
	}
	if pt.BatchID != batchID {
		return nil, model.ErrNotFound
	}
	if err := s.store.RejectMeasurePoint(ctx, id, reason); err != nil {
		return nil, err
	}
	return s.store.GetMeasurePoint(ctx, id)
}
