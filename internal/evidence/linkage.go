package evidence

import (
	"task219-colorreview/internal/model"
)

// Linkage 描述证据与导致色差的候选原因之间的关联强度。
type Linkage struct {
	Evidence   *model.ProcessEvidence
	Candidate  model.Verdict
	Confidence float64
}

// InferCandidate 根据证据类别推断最可能的色差来源。
// 浴液条件证据 -> bath；取样位置证据 -> sampling；仪器校准证据 -> instrument。
func InferCandidate(e *model.ProcessEvidence) model.Verdict {
	switch e.Kind {
	case model.EvidenceBath:
		return model.VerdictBath
	case model.EvidenceSampling:
		return model.VerdictSampling
	case model.EvidenceInstrument:
		return model.VerdictInstrument
	default:
		return model.VerdictInconclusive
	}
}

// DetectConflict 检测两条证据是否指向互相矛盾的原因。
// 只有同属一个批次、且指向不同确定原因时才算冲突。
func DetectConflict(a, b *model.ProcessEvidence) bool {
	if a == nil || b == nil {
		return false
	}
	ca, cb := InferCandidate(a), InferCandidate(b)
	if ca == model.VerdictInconclusive || cb == model.VerdictInconclusive {
		return false
	}
	return ca != cb
}

// ResolveVerdict 对一组已确认证据做归并，输出最终判定。
// 所有已确认证据指向同一原因时返回该原因；否则返回 mixed（混合来源）。
func ResolveVerdict(confirmed []*model.ProcessEvidence) model.Verdict {
	if len(confirmed) == 0 {
		return model.VerdictInconclusive
	}
	first := InferCandidate(confirmed[0])
	for _, e := range confirmed[1:] {
		if InferCandidate(e) != first {
			return model.VerdictMixed
		}
	}
	if first == model.VerdictInconclusive {
		return model.VerdictInconclusive
	}
	return first
}
