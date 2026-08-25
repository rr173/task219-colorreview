package store

import (
	"context"
	"fmt"
	"time"

	"task219-colorreview/internal/model"
)

const measureCols = "id, batch_id, sample_no, position, color_space, l, a, b, measured_at, status, reject_reason, delta_e, created_at"

// CreateMeasurePoint 写入测色点，样本序号在批次内幂等。
func (s *Store) CreateMeasurePoint(ctx context.Context, m *model.MeasurePoint) (*model.MeasurePoint, error) {
	ts := now()
	if m.Status == "" {
		m.Status = model.MeasurePending
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO measure_points (batch_id, sample_no, position, color_space, l, a, b, measured_at, status, reject_reason, delta_e, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.BatchID, m.SampleNo, m.Position, m.ColorSpace, m.L, m.A, m.B,
		m.MeasuredAt.UTC().Format(time.RFC3339Nano), string(m.Status), m.RejectReason, m.DeltaE, ts)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, model.ErrDuplicateSample
		}
		return nil, fmt.Errorf("insert measure point: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.GetMeasurePoint(ctx, id)
}

// GetMeasurePoint 按 ID 查询测色点。
func (s *Store) GetMeasurePoint(ctx context.Context, id int64) (*model.MeasurePoint, error) {
	var m model.MeasurePoint
	var measuredAt, createdAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT `+measureCols+` FROM measure_points WHERE id = ?`, id).
		Scan(&m.ID, &m.BatchID, &m.SampleNo, &m.Position, &m.ColorSpace,
			&m.L, &m.A, &m.B, &measuredAt, &m.Status, &m.RejectReason, &m.DeltaE, &createdAt)
	if err != nil {
		return nil, mapErr(err)
	}
	m.MeasuredAt, _ = time.Parse(time.RFC3339Nano, measuredAt)
	m.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return &m, nil
}

// ListMeasurePoints 列出某批次全部测色点，按样本序号升序。
func (s *Store) ListMeasurePoints(ctx context.Context, batchID int64) ([]*model.MeasurePoint, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+measureCols+` FROM measure_points WHERE batch_id = ? ORDER BY sample_no ASC`, batchID)
	if err != nil {
		return nil, fmt.Errorf("list measure points: %w", err)
	}
	defer rows.Close()

	var out []*model.MeasurePoint
	for rows.Next() {
		var m model.MeasurePoint
		var measuredAt, createdAt string
		if err := rows.Scan(&m.ID, &m.BatchID, &m.SampleNo, &m.Position, &m.ColorSpace,
			&m.L, &m.A, &m.B, &measuredAt, &m.Status, &m.RejectReason, &m.DeltaE, &createdAt); err != nil {
			return nil, fmt.Errorf("scan measure point: %w", err)
		}
		m.MeasuredAt, _ = time.Parse(time.RFC3339Nano, measuredAt)
		m.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		out = append(out, &m)
	}
	return out, rows.Err()
}

// SetMeasureStatus 更新测色点状态。
// 剔除（rejected）是不可逆终局状态：被剔除的点无论重算多少次都不会被
// 重新标记为 valid/anomaly，其 reject_reason 也不会被清空。写入条件显式
// 排除 rejected 行，因此即便上层误调用，剔除语义也不会被破坏。
func (s *Store) SetMeasureStatus(ctx context.Context, id int64, status model.MeasureStatus) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE measure_points SET status = ?, reject_reason = '' WHERE id = ? AND status != ?`,
		string(status), id, string(model.MeasureRejected))
	if err != nil {
		return fmt.Errorf("set measure status: %w", err)
	}
	return nil
}

// RejectMeasurePoint 剔除测色点：必须携带原因，且封存批次不可剔除。
func (s *Store) RejectMeasurePoint(ctx context.Context, id int64, reason string) error {
	if reason == "" {
		return model.ErrRejectReasonMissing
	}
	res, err := s.db.ExecContext(ctx,
		`UPDATE measure_points SET status = ?, reject_reason = ? WHERE id = ?`,
		string(model.MeasureRejected), reason, id)
	if err != nil {
		return fmt.Errorf("reject measure point: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// SetMeasureDeltaE 回填测色点与目标色的色差。
func (s *Store) SetMeasureDeltaE(ctx context.Context, id int64, deltaE float64) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE measure_points SET delta_e = ? WHERE id = ?`, deltaE, id)
	if err != nil {
		return fmt.Errorf("set measure deltaE: %w", err)
	}
	return nil
}

// GetMeasureBySample 按样本序号查询某批次测色点。
func (s *Store) GetMeasureBySample(ctx context.Context, batchID int64, sampleNo int) (*model.MeasurePoint, error) {
	var m model.MeasurePoint
	var measuredAt, createdAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT `+measureCols+` FROM measure_points WHERE batch_id = ? AND sample_no = ?`,
		batchID, sampleNo).
		Scan(&m.ID, &m.BatchID, &m.SampleNo, &m.Position, &m.ColorSpace,
			&m.L, &m.A, &m.B, &measuredAt, &m.Status, &m.RejectReason, &m.DeltaE, &createdAt)
	if err != nil {
		return nil, mapErr(err)
	}
	m.MeasuredAt, _ = time.Parse(time.RFC3339Nano, measuredAt)
	m.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return &m, nil
}
