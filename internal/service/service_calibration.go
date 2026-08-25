package service

import (
	"context"
	"time"

	"task219-colorreview/internal/model"
	"task219-colorreview/internal/sampling"
)

// CreateCalibration 记录测色仪校准（含偏移计算）。
func (s *Service) CreateCalibration(ctx context.Context, c *model.InstrumentCalibration) (*model.InstrumentCalibration, error) {
	if err := sampling.ValidateCalibration(c); err != nil {
		return nil, err
	}
	return s.store.CreateCalibration(ctx, c)
}

// LatestCalibration 查询某仪器最近校准。
func (s *Service) LatestCalibration(ctx context.Context, instrumentID string) (*model.InstrumentCalibration, error) {
	return s.store.LatestCalibration(ctx, instrumentID)
}

// EffectiveCalibration 查询某仪器在指定时刻已经生效的最新校准。
// 用于测色点修正：测量样本必须套用测量时点已生效的校准偏移，而非更早或更晚的记录。
func (s *Service) EffectiveCalibration(ctx context.Context, instrumentID string, at time.Time) (*model.InstrumentCalibration, error) {
	history, err := s.store.ListCalibrationsByInstrument(ctx, instrumentID)
	if err != nil {
		return nil, err
	}
	schedule := sampling.NewSchedule(instrumentID, history)
	if cal := schedule.EffectiveAt(at); cal != nil {
		return cal, nil
	}
	return nil, model.ErrCalibrationMissing
}

// ListCalibrations 列出全部校准记录。
func (s *Service) ListCalibrations(ctx context.Context) ([]*model.InstrumentCalibration, error) {
	return s.store.ListCalibrations(ctx)
}
