package service

import (
	"context"
	"fmt"
	"time"

	"task219-colorreview/internal/batch"
	"task219-colorreview/internal/evidence"
	"task219-colorreview/internal/model"
	"task219-colorreview/internal/review"
)

// CreateConclusion 创建复核结论（自动归并已确认证据推断判定）。
func (s *Service) CreateConclusion(ctx context.Context, c *model.ReviewConclusion) (*model.ReviewConclusion, error) {
	if err := review.Validate(c); err != nil {
		return nil, err
	}
	b, err := s.store.GetBatch(ctx, c.BatchID)
	if err != nil {
		return nil, err
	}
	if batch.IsSealed(b) {
		return nil, model.ErrBatchSealed
	}
	c.Status = model.ConclusionDraft
	c.Version = 1
	return s.store.CreateConclusion(ctx, c)
}

// GetConclusion 查询批次当前结论。
func (s *Service) GetConclusion(ctx context.Context, batchID int64) (*model.ReviewConclusion, error) {
	return s.store.GetConclusionByBatch(ctx, batchID)
}

// PublishConclusion 发布结论：写入不可变版本快照并冻结。
func (s *Service) PublishConclusion(ctx context.Context, id int64) (*model.ReviewConclusion, error) {
	c, err := s.store.GetConclusion(ctx, id)
	if err != nil {
		return nil, err
	}
	if !review.CanPublish(c) {
		return nil, model.ErrInvalidTransition
	}
	now := time.Now().UTC()
	c.Status = model.ConclusionPublished
	c.PublishedAt = &now
	if err := s.store.UpdateConclusion(ctx, c); err != nil {
		return nil, err
	}
	// 发布即写入不可变快照。
	_ = s.store.SnapshotConclusionVersion(ctx, c)
	return s.store.GetConclusion(ctx, id)
}

// SupersedeConclusion 对已发布结论建立替代版本（版本 +1）。
func (s *Service) SupersedeConclusion(ctx context.Context, id int64, verdict model.Verdict, summary string) (*model.ReviewConclusion, error) {
	cur, err := s.store.GetConclusion(ctx, id)
	if err != nil {
		return nil, err
	}
	v := review.NewVersioner(cur)
	next, err := v.Supersede(verdict, summary)
	if err != nil {
		return nil, err
	}
	// 旧结论置为替代态。
	if err := s.store.SupersedeConclusion(ctx, cur.ID); err != nil {
		return nil, err
	}
	return s.store.CreateConclusion(ctx, next)
}

// ListConclusionVersions 列出结论版本历史。
func (s *Service) ListConclusionVersions(ctx context.Context, batchID int64) ([]map[string]any, error) {
	return s.store.ListConclusionVersions(ctx, batchID)
}

// InferVerdict 根据批次已确认证据推断判定（辅助创建结论）。
func (s *Service) InferVerdict(ctx context.Context, batchID int64) (model.Verdict, error) {
	evs, err := s.store.ListEvidences(ctx, batchID)
	if err != nil {
		return "", err
	}
	var confirmed []*model.ProcessEvidence
	for _, e := range evs {
		if e.Status == model.EvidenceConfirmed {
			confirmed = append(confirmed, e)
		}
	}
	if len(confirmed) == 0 {
		return model.VerdictInconclusive, fmt.Errorf("%w: no confirmed evidence", model.ErrConflictUnresolved)
	}
	return evidence.ResolveVerdict(confirmed), nil
}
