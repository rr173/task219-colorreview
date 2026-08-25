package service

import (
	"context"
	"time"

	"task219-colorreview/internal/batch"
	"task219-colorreview/internal/evidence"
	"task219-colorreview/internal/model"
)

// AddEvidence 提交工艺证据。
func (s *Service) AddEvidence(ctx context.Context, e *model.ProcessEvidence) (*model.ProcessEvidence, error) {
	if err := evidence.Validate(e); err != nil {
		return nil, err
	}
	b, err := s.store.GetBatch(ctx, e.BatchID)
	if err != nil {
		return nil, err
	}
	if batch.IsSealed(b) {
		return nil, model.ErrBatchSealed
	}
	if e.AttachedAt.IsZero() {
		e.AttachedAt = time.Now()
	}
	e.Status = model.EvidencePending
	return s.store.CreateEvidence(ctx, e)
}

// ListEvidences 列出批次证据。
func (s *Service) ListEvidences(ctx context.Context, batchID int64) ([]*model.ProcessEvidence, error) {
	return s.store.ListEvidences(ctx, batchID)
}

// ConfirmEvidence 确认证据，并检测与同批次其它已确认证据是否冲突。
func (s *Service) ConfirmEvidence(ctx context.Context, batchID, id int64) (*model.ProcessEvidence, error) {
	b, err := s.store.GetBatch(ctx, batchID)
	if err != nil {
		return nil, err
	}
	if batch.IsSealed(b) {
		return nil, model.ErrBatchSealed
	}
	e, err := s.store.GetEvidence(ctx, id)
	if err != nil {
		return nil, err
	}
	if e.BatchID != batchID {
		return nil, model.ErrNotFound
	}
	// 已标记为冲突的证据不可再次确认：否则会重新进入已确认集合，
	// 被后一次确认操作污染最终归因。需先消解冲突后再确认。
	if !evidence.CanConfirm(e) {
		return nil, model.ErrInvalidTransition
	}
	if err := s.store.SetEvidenceStatus(ctx, id, model.EvidenceConfirmed); err != nil {
		return nil, err
	}
	// 冲突检测：与新确认证据指向不同原因的已确认证据标记为 conflict。
	all, _ := s.store.ListEvidences(ctx, batchID)
	for _, other := range all {
		if other.ID == id || other.Status != model.EvidenceConfirmed {
			continue
		}
		if evidence.DetectConflict(e, other) {
			_ = s.store.SetEvidenceStatus(ctx, other.ID, model.EvidenceConflict)
		}
	}
	return s.store.GetEvidence(ctx, id)
}
