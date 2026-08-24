package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"task219-colorreview/internal/model"
)

// SaveBathCurve 保存（或覆盖）某批次某通道的浴液曲线。
// 曲线采样点序列以 JSON 编码存储，保证重启后完整恢复。
func (s *Store) SaveBathCurve(ctx context.Context, c *model.BathCurve) (*model.BathCurve, error) {
	if len(c.Points) == 0 {
		return nil, model.ErrInvalidArgument
	}
	raw, err := json.Marshal(c.Points)
	if err != nil {
		return nil, fmt.Errorf("marshal curve: %w", err)
	}
	ts := now()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO bath_curves (batch_id, channel, points, created_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(batch_id, channel) DO UPDATE SET points = excluded.points, created_at = excluded.created_at`,
		c.BatchID, c.Channel, string(raw), ts)
	if err != nil {
		return nil, fmt.Errorf("save bath curve: %w", err)
	}
	return s.GetBathCurve(ctx, c.BatchID, c.Channel)
}

// GetBathCurve 查询某批次某通道的浴液曲线。
func (s *Store) GetBathCurve(ctx context.Context, batchID int64, channel string) (*model.BathCurve, error) {
	var id int64
	var raw, createdAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, points, created_at FROM bath_curves WHERE batch_id = ? AND channel = ?`,
		batchID, channel).Scan(&id, &raw, &createdAt)
	if err != nil {
		return nil, mapErr(err)
	}
	var pts []model.BathCurvePoint
	if err := json.Unmarshal([]byte(raw), &pts); err != nil {
		return nil, fmt.Errorf("unmarshal curve: %w", err)
	}
	t, _ := time.Parse(time.RFC3339Nano, createdAt)
	return &model.BathCurve{ID: id, BatchID: batchID, Channel: channel, Points: pts, CreatedAt: t}, nil
}

// ListBathCurves 列出某批次的全部浴液曲线。
func (s *Store) ListBathCurves(ctx context.Context, batchID int64) ([]*model.BathCurve, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, channel, points, created_at FROM bath_curves WHERE batch_id = ? ORDER BY channel`,
		batchID)
	if err != nil {
		return nil, fmt.Errorf("list bath curves: %w", err)
	}
	defer rows.Close()

	var out []*model.BathCurve
	for rows.Next() {
		var id int64
		var channel, raw, createdAt string
		if err := rows.Scan(&id, &channel, &raw, &createdAt); err != nil {
			return nil, fmt.Errorf("scan curve: %w", err)
		}
		var pts []model.BathCurvePoint
		if err := json.Unmarshal([]byte(raw), &pts); err != nil {
			return nil, fmt.Errorf("unmarshal curve: %w", err)
		}
		t, _ := time.Parse(time.RFC3339Nano, createdAt)
		out = append(out, &model.BathCurve{ID: id, BatchID: batchID, Channel: channel, Points: pts, CreatedAt: t})
	}
	return out, rows.Err()
}

// CurveValueAt 在曲线上按时间线性插值取值，用于对齐测色取样时间。
func CurveValueAt(points []model.BathCurvePoint, at time.Time) (float64, bool) {
	if len(points) == 0 {
		return 0, false
	}
	atU := at.UnixNano()
	first := points[0].Timestamp.UnixNano()
	if atU <= first {
		return points[0].Value, true
	}
	last := points[len(points)-1].Timestamp.UnixNano()
	if atU >= last {
		return points[len(points)-1].Value, true
	}
	for i := 1; i < len(points); i++ {
		prev, cur := points[i-1], points[i]
		pn, cn := prev.Timestamp.UnixNano(), cur.Timestamp.UnixNano()
		if atU >= pn && atU <= cn {
			if cn == pn {
				return cur.Value, true
			}
			frac := float64(atU-pn) / float64(cn-pn)
			return prev.Value + frac*(cur.Value-prev.Value), true
		}
	}
	return 0, false
}
