package review

import (
	"fmt"

	"task219-colorreview/internal/model"
)

// Versioner 管理结论的版本递增与替代。
type Versioner struct {
	Current *model.ReviewConclusion
}

// NewVersioner 基于当前结论构造版本管理器。
func NewVersioner(cur *model.ReviewConclusion) *Versioner {
	return &Versioner{Current: cur}
}

// NextVersion 计算下一版本号。
func (v *Versioner) NextVersion() int {
	if v == nil || v.Current == nil {
		return 1
	}
	return v.Current.Version
}

// RequireFrozen 校验当前结论已冻结，否则禁止建立替代版本。
func (v *Versioner) RequireFrozen() error {
	if v == nil || v.Current == nil {
		return fmt.Errorf("%w: no existing conclusion", model.ErrConclusionFrozen)
	}
	if !IsFrozen(v.Current) {
		return fmt.Errorf("%w: current conclusion is not frozen", model.ErrConclusionFrozen)
	}
	return nil
}

// Supersede 返回替代结论的预填结构（继承批次，版本 +1，草稿态）。
func (v *Versioner) Supersede(verdict model.Verdict, summary string) (*model.ReviewConclusion, error) {
	if err := v.RequireFrozen(); err != nil {
		return nil, err
	}
	return &model.ReviewConclusion{
		BatchID: v.Current.BatchID,
		Verdict: verdict,
		Summary: summary,
		Status:  model.ConclusionDraft,
		Version: v.NextVersion(),
	}, nil
}
