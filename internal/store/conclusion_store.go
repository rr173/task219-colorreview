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

// execer 抽象可执行 SQL 的目标：既能作用于底层连接，也能作用于事务，
// 便于把多步写入纳入同一事务原子提交/回滚。
type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// UpdateConclusion 更新结论内容与状态。
func (s *Store) UpdateConclusion(ctx context.Context, c *model.ReviewConclusion) error {
	return updateConclusion(ctx, s.db, c)
}

// updateConclusion 在给定执行器上更新结论，供事务内复用。
func updateConclusion(ctx context.Context, ex execer, c *model.ReviewConclusion) error {
	var publishedAt any
	if c.PublishedAt != nil {
		publishedAt = c.PublishedAt.UTC().Format(time.RFC3339Nano)
	}
	_, err := ex.ExecContext(ctx,
		`UPDATE conclusions SET verdict = ?, summary = ?, status = ?, version = ?, published_at = ?, updated_at = ? WHERE id = ?`,
		string(c.Verdict), c.Summary, string(c.Status), c.Version, publishedAt, now(), c.ID)
	if err != nil {
		return fmt.Errorf("update conclusion: %w", err)
	}
	return nil
}

// snapshotConclusionVersion 在给定执行器上写入不可变版本快照，供事务内复用。
func snapshotConclusionVersion(ctx context.Context, ex execer, c *model.ReviewConclusion) error {
	_, err := ex.ExecContext(ctx,
		`INSERT OR IGNORE INTO conclusion_versions (conclusion_id, batch_id, verdict, summary, version, snapshot_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		c.ID, c.BatchID, string(c.Verdict), c.Summary, c.Version, now())
	if err != nil {
		return fmt.Errorf("snapshot conclusion: %w", err)
	}
	return nil
}

// SnapshotConclusionVersion 写入不可变结论版本快照。
func (s *Store) SnapshotConclusionVersion(ctx context.Context, c *model.ReviewConclusion) error {
	return snapshotConclusionVersion(ctx, s.db, c)
}

// PublishConclusionAtomically 原子地发布结论：在同一事务内先置为发布态、
// 再写入不可变版本快照。任一步失败整体回滚，结论不会被单独留在"已发布但无快照"
// 的半成品状态。
func (s *Store) PublishConclusionAtomically(ctx context.Context, c *model.ReviewConclusion) error {
	return WithTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := updateConclusion(ctx, tx, c); err != nil {
			return err
		}
		return snapshotConclusionVersion(ctx, tx, c)
	})
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
