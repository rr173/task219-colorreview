package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"task219-colorreview/internal/model"
)

const batchCols = "id, code, name, recipe, color_space, status, created_at, updated_at"

// CreateBatch 创建染色批次。
func (s *Store) CreateBatch(ctx context.Context, b *model.DyeBatch) (*model.DyeBatch, error) {
	ts := now()
	if b.Status == "" {
		b.Status = model.BatchRecipe
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO batches (code, name, recipe, color_space, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		b.Code, b.Name, b.Recipe, b.ColorSpace, string(b.Status), ts, ts)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, model.ErrAlreadyExists
		}
		return nil, fmt.Errorf("insert batch: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}
	return s.GetBatch(ctx, id)
}

// GetBatch 按 ID 查询批次。
func (s *Store) GetBatch(ctx context.Context, id int64) (*model.DyeBatch, error) {
	var b model.DyeBatch
	var createdAt, updatedAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT `+batchCols+` FROM batches WHERE id = ?`, id).
		Scan(&b.ID, &b.Code, &b.Name, &b.Recipe, &b.ColorSpace, &b.Status,
			&createdAt, &updatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	b.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	b.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return &b, nil
}

// ListBatches 列出全部批次，按创建时间倒序。
func (s *Store) ListBatches(ctx context.Context) ([]*model.DyeBatch, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+batchCols+` FROM batches ORDER BY id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list batches: %w", err)
	}
	defer rows.Close()

	var out []*model.DyeBatch
	for rows.Next() {
		var b model.DyeBatch
		var createdAt, updatedAt string
		if err := rows.Scan(&b.ID, &b.Code, &b.Name, &b.Recipe, &b.ColorSpace,
			&b.Status, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan batch: %w", err)
		}
		b.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		b.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
		out = append(out, &b)
	}
	return out, rows.Err()
}

// SetBatchStatus 原子更新批次状态（带乐观状态前置校验）。
func (s *Store) SetBatchStatus(ctx context.Context, id int64, from, to model.BatchStatus) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE batches SET status = ?, updated_at = ? WHERE id = ? AND status = ?`,
		string(to), now(), id, string(from))
	if err != nil {
		return fmt.Errorf("set batch status: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return model.ErrInvalidTransition
	}
	return nil
}

// SealBatch 封存批次：仅当当前状态允许封存时生效。
func (s *Store) SealBatch(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE batches SET status = ?, updated_at = ? WHERE id = ? AND status IN (?, ?)`,
		string(model.BatchSealed), now(), id,
		string(model.BatchPendingReview), string(model.BatchConfirmed))
	if err != nil {
		return fmt.Errorf("seal batch: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		// 已封存或状态不允许封存。
		var cur model.BatchStatus
		_ = s.db.QueryRowContext(ctx, `SELECT status FROM batches WHERE id = ?`, id).Scan(&cur)
		if cur == model.BatchSealed {
			return model.ErrBatchSealed
		}
		return model.ErrInvalidTransition
	}
	return nil
}

// UpdateBatchColorSpace 更新批次色彩空间声明。
func (s *Store) UpdateBatchColorSpace(ctx context.Context, id int64, colorSpace string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE batches SET color_space = ?, updated_at = ? WHERE id = ?`,
		colorSpace, now(), id)
	if err != nil {
		return fmt.Errorf("update color space: %w", err)
	}
	return nil
}

// BatchExists 判断批次是否存在。
func (s *Store) BatchExists(ctx context.Context, id int64) (bool, error) {
	var one int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM batches WHERE id = ?`, id).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteBatch 删除批次及其关联数据，用于失败导入的补偿回滚。
func (s *Store) DeleteBatch(ctx context.Context, id int64) error {
	return WithTx(ctx, s.db, func(tx *sql.Tx) error {
		for _, table := range []string{
			"conclusion_versions", "conclusions", "evidences", "measure_points",
			"bath_curves", "batches",
		} {
			if _, err := tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE "+map[string]string{
				"conclusion_versions": "batch_id",
				"conclusions":         "batch_id",
				"evidences":           "batch_id",
				"measure_points":      "batch_id",
				"bath_curves":         "batch_id",
				"batches":             "id",
			}[table]+" = ?", id); err != nil {
				return fmt.Errorf("delete %s: %w", table, err)
			}
		}
		return nil
	})
}
