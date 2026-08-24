// Package evidence 负责工艺证据的管理与关联判定。
package evidence

import (
	"fmt"

	"task219-colorreview/internal/model"
)

// Validate 校验工艺证据输入。
func Validate(e *model.ProcessEvidence) error {
	if e.BatchID <= 0 {
		return fmt.Errorf("%w: batch_id required", model.ErrInvalidArgument)
	}
	if e.Description == "" {
		return fmt.Errorf("%w: description required", model.ErrInvalidArgument)
	}
	switch e.Kind {
	case model.EvidenceBath, model.EvidenceSampling, model.EvidenceInstrument, model.EvidenceOther:
	default:
		return fmt.Errorf("%w: unknown evidence kind %q", model.ErrInvalidArgument, e.Kind)
	}
	return nil
}

// CanConfirm 判断证据能否被确认：冲突证据需先消解。
func CanConfirm(e *model.ProcessEvidence) bool {
	return e != nil && e.Status != model.EvidenceConflict
}

// Linkable 判断证据是否处于可关联状态。
func Linkable(e *model.ProcessEvidence) bool {
	return e != nil && (e.Status == model.EvidencePending || e.Status == model.EvidenceLinked)
}

// IsConflict 判断证据是否被标记为冲突。
func IsConflict(e *model.ProcessEvidence) bool { return e != nil && e.Status == model.EvidenceConflict }
