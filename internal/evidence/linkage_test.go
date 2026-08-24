package evidence

import (
	"testing"

	"task219-colorreview/internal/model"
)

func TestInferCandidate(t *testing.T) {
	cases := map[model.EvidenceKind]model.Verdict{
		model.EvidenceBath:       model.VerdictBath,
		model.EvidenceSampling:   model.VerdictSampling,
		model.EvidenceInstrument: model.VerdictInstrument,
		model.EvidenceOther:      model.VerdictInconclusive,
	}
	for kind, want := range cases {
		if got := InferCandidate(&model.ProcessEvidence{Kind: kind}); got != want {
			t.Fatalf("InferCandidate(%s) = %s, want %s", kind, got, want)
		}
	}
}

func TestDetectConflict(t *testing.T) {
	bath := &model.ProcessEvidence{Kind: model.EvidenceBath}
	sampling := &model.ProcessEvidence{Kind: model.EvidenceSampling}
	if !DetectConflict(bath, sampling) {
		t.Fatal("bath vs sampling should conflict")
	}
	if DetectConflict(bath, bath) {
		t.Fatal("same evidence kind should not conflict")
	}
	other := &model.ProcessEvidence{Kind: model.EvidenceOther}
	if DetectConflict(bath, other) {
		t.Fatal("inconclusive evidence should not conflict")
	}
}

func TestResolveVerdict(t *testing.T) {
	if got := ResolveVerdict(nil); got != model.VerdictInconclusive {
		t.Fatalf("empty should be inconclusive, got %s", got)
	}
	uniform := []*model.ProcessEvidence{
		{Kind: model.EvidenceBath},
		{Kind: model.EvidenceBath},
	}
	if got := ResolveVerdict(uniform); got != model.VerdictBath {
		t.Fatalf("uniform should be bath, got %s", got)
	}
	mixed := []*model.ProcessEvidence{
		{Kind: model.EvidenceBath},
		{Kind: model.EvidenceInstrument},
	}
	if got := ResolveVerdict(mixed); got != model.VerdictMixed {
		t.Fatalf("mixed should be mixed, got %s", got)
	}
}
