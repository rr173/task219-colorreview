package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"task219-colorreview/internal/model"
)

const conclusionCols = "id, batch_id, verdict, summary, status, version, published_at, created_at, updated_at"

// CreateConclusion 创建复核结论（初版 version=1，草稿态）。
func (s *Store) CreateConclusion(ctx context.Context, c *model.ReviewConclusion) (*model.ReviewConclusion, error) {
	ts := now()
	if c.Status == "" {
		c.Status = model.ConclusionDraft
	}
	if c.Version == 0 {
		c.Version = 1
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO conclusions (batch_id, verdict, summary, status, version, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		c.BatchID, string(c.Verdict), c.Summary, string(c.Status), c.Version, ts, ts)
	if err != nil {
		return nil, fmt.Errorf("insert conclusion: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.GetConclusion(ctx, id)
}

// GetConclusion 按 ID 查询结论。
func (s *Store) GetConclusion(ctx context.Context, id int64) (*model.ReviewConclusion, error) {
	var c model.ReviewConclusion
	var publishedAt sql.NullString
	var createdAt, updatedAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT `+conclusionCols+` FROM conclusions WHERE id = ?`, id).
		Scan(&c.ID, &c.BatchID, &c.Verdict, &c.Summary, &c.Status, &c.Version,
			&publishedAt, &createdAt, &updatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	if publishedAt.Valid {
		t, _ := time.Parse(time.RFC3339Nano, publishedAt.String)
		c.PublishedAt = &t
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	c.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return &c, nil
}

// GetConclusionByBatch 查询某批次当前有效结论（非替代态）。
func (s *Store) GetConclusionByBatch(ctx context.Context, batchID int64) (*model.ReviewConclusion, error) {
	var c model.ReviewConclusion
	var publishedAt sql.NullString
	var createdAt, updatedAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT `+conclusionCols+` FROM conclusions WHERE batch_id = ? AND status != ? ORDER BY version DESC LIMIT 1`,
		batchID, string(model.ConclusionSuperseded)).
		Scan(&c.ID, &c.BatchID, &c.Verdict, &c.Summary, &c.Status, &c.Version,
			&publishedAt, &createdAt, &updatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	if publishedAt.Valid {
		t, _ := time.Parse(time.RFC3339Nano, publishedAt.String)
		c.PublishedAt = &t
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	c.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return &c, nil
}

// UpdateConclusion 更新结论内容与状态。
func (s *Store) UpdateConclusion(ctx context.Context, c *model.ReviewConclusion) error {
	var publishedAt any
	if c.PublishedAt != nil {
		publishedAt = c.PublishedAt.UTC().Format(time.RFC3339Nano)
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE conclusions SET verdict = ?, summary = ?, status = ?, version = ?, published_at = ?, updated_at = ? WHERE id = ?`,
		string(c.Verdict), c.Summary, string(c.Status), c.Version, publishedAt, now(), c.ID)
	if err != nil {
		return fmt.Errorf("update conclusion: %w", err)
	}
	return nil
}

// SupersedeConclusion 把当前结论置为替代态（版本迭代时）。
func (s *Store) SupersedeConclusion(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE conclusions SET status = ?, updated_at = ? WHERE id = ?`,
		string(model.ConclusionSuperseded), now(), id)
	if err != nil {
		return fmt.Errorf("supersede conclusion: %w", err)
	}
	return nil
}

// SnapshotConclusionVersion 写入不可变结论版本快照。
func (s *Store) SnapshotConclusionVersion(ctx context.Context, c *model.ReviewConclusion) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO conclusion_versions (conclusion_id, batch_id, verdict, summary, version, snapshot_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		c.ID, c.BatchID, string(c.Verdict), c.Summary, c.Version, now())
	if err != nil {
		return fmt.Errorf("snapshot conclusion: %w", err)
	}
	return nil
}

// ListConclusionVersions 列出某批次结论的历史版本快照。
func (s *Store) ListConclusionVersions(ctx context.Context, batchID int64) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT version, verdict, summary, snapshot_at FROM conclusion_versions WHERE batch_id = ? ORDER BY version ASC`,
		batchID)
	if err != nil {
		return nil, fmt.Errorf("list conclusion versions: %w", err)
	}
	defer rows.Close()

	var out []map[string]any
	for rows.Next() {
		var version int
		var verdict, summary, snapshotAt string
		if err := rows.Scan(&version, &verdict, &summary, &snapshotAt); err != nil {
			return nil, fmt.Errorf("scan conclusion version: %w", err)
		}
		out = append(out, map[string]any{
			"version":     version,
			"verdict":     verdict,
			"summary":     summary,
			"snapshot_at": snapshotAt,
		})
	}
	return out, rows.Err()
}
