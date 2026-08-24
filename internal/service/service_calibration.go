package service

import (
	"context"

	"task219-colorreview/internal/model"
	"task219-colorreview/internal/sampling"
)

// CreateCalibration 记录测色仪校准（含偏移计算）。
func (s *Service) CreateCalibration(ctx context.Context, c *model.InstrumentCalibration) (*model.InstrumentCalibration, error) {
	if err := sampling.ValidateCalibration(c); err != nil {
		return nil, err
	}
	created, err := s.store.CreateCalibration(ctx, c)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// LatestCalibration 查询某仪器最近校准。
func (s *Service) LatestCalibration(ctx context.Context, instrumentID string) (*model.InstrumentCalibration, error) {
	return s.store.LatestCalibration(ctx, instrumentID)
}

// ListCalibrations 列出全部校准记录。
func (s *Service) ListCalibrations(ctx context.Context) ([]*model.InstrumentCalibration, error) {
	return s.store.ListCalibrations(ctx)
}
