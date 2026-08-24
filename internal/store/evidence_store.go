package store

import (
	"context"
	"fmt"
	"time"

	"task219-colorreview/internal/model"
)

const evidenceCols = "id, batch_id, kind, description, status, attached_at, created_at"

// CreateEvidence 写入工艺证据。
func (s *Store) CreateEvidence(ctx context.Context, e *model.ProcessEvidence) (*model.ProcessEvidence, error) {
	ts := now()
	if e.Status == "" {
		e.Status = model.EvidencePending
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO evidences (batch_id, kind, description, status, attached_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		e.BatchID, string(e.Kind), e.Description, string(e.Status),
		e.AttachedAt.UTC().Format(time.RFC3339Nano), ts)
	if err != nil {
		return nil, fmt.Errorf("insert evidence: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.GetEvidence(ctx, id)
}

// GetEvidence 按 ID 查询工艺证据。
func (s *Store) GetEvidence(ctx context.Context, id int64) (*model.ProcessEvidence, error) {
	var e model.ProcessEvidence
	var attachedAt, createdAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT `+evidenceCols+` FROM evidences WHERE id = ?`, id).
		Scan(&e.ID, &e.BatchID, &e.Kind, &e.Description, &e.Status, &attachedAt, &createdAt)
	if err != nil {
		return nil, mapErr(err)
	}
	e.AttachedAt, _ = time.Parse(time.RFC3339Nano, attachedAt)
	e.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	return &e, nil
}

// ListEvidences 列出某批次全部工艺证据。
func (s *Store) ListEvidences(ctx context.Context, batchID int64) ([]*model.ProcessEvidence, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+evidenceCols+` FROM evidences WHERE batch_id = ? ORDER BY id ASC`, batchID)
	if err != nil {
		return nil, fmt.Errorf("list evidences: %w", err)
	}
	defer rows.Close()

	var out []*model.ProcessEvidence
	for rows.Next() {
		var e model.ProcessEvidence
		var attachedAt, createdAt string
		if err := rows.Scan(&e.ID, &e.BatchID, &e.Kind, &e.Description, &e.Status,
			&attachedAt, &createdAt); err != nil {
			return nil, fmt.Errorf("scan evidence: %w", err)
		}
		e.AttachedAt, _ = time.Parse(time.RFC3339Nano, attachedAt)
		e.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		out = append(out, &e)
	}
	return out, rows.Err()
}

// SetEvidenceStatus 更新工艺证据状态。
func (s *Store) SetEvidenceStatus(ctx context.Context, id int64, status model.EvidenceStatus) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE evidences SET status = ? WHERE id = ?`, string(status), id)
	if err != nil {
		return fmt.Errorf("set evidence status: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}
