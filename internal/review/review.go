// Package review 负责复核结论的生成与版本化发布。
package review

import (
	"fmt"

	"task219-colorreview/internal/model"
)

// Validate 校验结论字段。
func Validate(c *model.ReviewConclusion) error {
	if c.BatchID <= 0 {
		return fmt.Errorf("%w: batch_id required", model.ErrInvalidArgument)
	}
	switch c.Verdict {
	case model.VerdictBath, model.VerdictSampling, model.VerdictInstrument,
		model.VerdictMixed, model.VerdictInconclusive:
	default:
		return fmt.Errorf("%w: unknown verdict %q", model.ErrInvalidArgument, c.Verdict)
	}
	if c.Summary == "" {
		return fmt.Errorf("%w: summary required", model.ErrInvalidArgument)
	}
	return nil
}

// CanPublish 判断结论能否发布。
func CanPublish(c *model.ReviewConclusion) bool {
	return c != nil && (c.Status == model.ConclusionDraft || c.Status == model.ConclusionPending)
}

// IsFrozen 判断结论是否已冻结（发布后只能建立替代版本）。
func IsFrozen(c *model.ReviewConclusion) bool {
	return c != nil && (c.Status == model.ConclusionPublished || c.Status == model.ConclusionSuperseded)
}
