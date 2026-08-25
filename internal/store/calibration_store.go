package store

import (
	"context"
	"fmt"
	"time"

	"task219-colorreview/internal/model"
)

const calibrationCols = "id, instrument_id, calibrated_at, ref_l, ref_a, ref_b, offset_l, offset_a, offset_b, created_at"

// CreateCalibration 记录一次测色仪校准。
func (s *Store) CreateCalibration(ctx context.Context, c *model.InstrumentCalibration) (*model.InstrumentCalibration, error) {
	ts := now()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO instrument_calibrations (instrument_id, calibrated_at, ref_l, ref_a, ref_b, offset_l, offset_a, offset_b, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.InstrumentID, c.CalibratedAt.UTC().Format(time.RFC3339Nano),
		c.RefL, c.RefA, c.RefB, c.OffsetL, c.OffsetA, c.OffsetB, ts)
	if err != nil {
		return nil, fmt.Errorf("insert calibration: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.GetCalibration(ctx, id)
}

// GetCalibration 按 ID 查询校准记录。
func (s *Store) GetCalibration(ctx context.Context, id int64) (*model.InstrumentCalibration, error) {
	var c model.InstrumentCalibration
	var calibratedAt, createdAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT `+calibrationCols+` FROM instrument_calibrations WHERE id = ?`, id).
		Scan(&c.ID, &c.InstrumentID, &calibratedAt, &c.RefL, &c.RefA, &c.RefB,
			&c.OffsetL, &c.OffsetA, &c.OffsetB, &createdAt)
	if err != nil {
		return nil, mapErr(err)
	}
	c.CalibratedAt, _ = time.Parse(time.RFC3339Nano, calibratedAt)
	c.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return &c, nil
}

// LatestCalibration 返回某仪器最近一次校准记录。
func (s *Store) LatestCalibration(ctx context.Context, instrumentID string) (*model.InstrumentCalibration, error) {
	var c model.InstrumentCalibration
	var calibratedAt, createdAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT `+calibrationCols+` FROM instrument_calibrations WHERE instrument_id = ? ORDER BY calibrated_at DESC, id DESC LIMIT 1`,
		instrumentID).
		Scan(&c.ID, &c.InstrumentID, &calibratedAt, &c.RefL, &c.RefA, &c.RefB,
			&c.OffsetL, &c.OffsetA, &c.OffsetB, &createdAt)
	if err != nil {
		return nil, mapErr(err)
	}
	c.CalibratedAt, _ = time.Parse(time.RFC3339Nano, calibratedAt)
	c.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return &c, nil
}

// EffectiveCalibration 返回某仪器在指定时刻已经生效的最新校准记录，
// 即校准时间不晚于 at 的最新一条。测量样本应套用此校准偏移，
// 而非更早的偏移或晚于测量时刻的校准。
func (s *Store) EffectiveCalibration(ctx context.Context, instrumentID string, at time.Time) (*model.InstrumentCalibration, error) {
	var c model.InstrumentCalibration
	var calibratedAt, createdAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT `+calibrationCols+` FROM instrument_calibrations
		 WHERE instrument_id = ? AND calibrated_at <= ?
		 ORDER BY calibrated_at DESC, id DESC LIMIT 1`,
		instrumentID, at.UTC().Format(time.RFC3339Nano)).
		Scan(&c.ID, &c.InstrumentID, &calibratedAt, &c.RefL, &c.RefA, &c.RefB,
			&c.OffsetL, &c.OffsetA, &c.OffsetB, &createdAt)
	if err != nil {
		return nil, mapErr(err)
	}
	c.CalibratedAt, _ = time.Parse(time.RFC3339Nano, calibratedAt)
	c.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return &c, nil
}

// ListCalibrationsByInstrument 列出某仪器全部校准记录，按时间升序。
func (s *Store) ListCalibrationsByInstrument(ctx context.Context, instrumentID string) ([]*model.InstrumentCalibration, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+calibrationCols+` FROM instrument_calibrations WHERE instrument_id = ? ORDER BY calibrated_at ASC, id ASC`,
		instrumentID)
	if err != nil {
		return nil, fmt.Errorf("list calibrations by instrument: %w", err)
	}
	defer rows.Close()

	var out []*model.InstrumentCalibration
	for rows.Next() {
		var c model.InstrumentCalibration
		var calibratedAt, createdAt string
		if err := rows.Scan(&c.ID, &c.InstrumentID, &calibratedAt, &c.RefL, &c.RefA, &c.RefB,
			&c.OffsetL, &c.OffsetA, &c.OffsetB, &createdAt); err != nil {
			return nil, fmt.Errorf("scan calibration: %w", err)
		}
		c.CalibratedAt, _ = time.Parse(time.RFC3339Nano, calibratedAt)
		c.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		out = append(out, &c)
	}
	return out, rows.Err()
}

// ListCalibrations 列出全部校准记录，按时间倒序。
func (s *Store) ListCalibrations(ctx context.Context) ([]*model.InstrumentCalibration, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+calibrationCols+` FROM instrument_calibrations ORDER BY calibrated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list calibrations: %w", err)
	}
	defer rows.Close()

	var out []*model.InstrumentCalibration
	for rows.Next() {
		var c model.InstrumentCalibration
		var calibratedAt, createdAt string
		if err := rows.Scan(&c.ID, &c.InstrumentID, &calibratedAt, &c.RefL, &c.RefA, &c.RefB,
			&c.OffsetL, &c.OffsetA, &c.OffsetB, &createdAt); err != nil {
			return nil, fmt.Errorf("scan calibration: %w", err)
		}
		c.CalibratedAt, _ = time.Parse(time.RFC3339Nano, calibratedAt)
		c.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		out = append(out, &c)
	}
	return out, rows.Err()
}
