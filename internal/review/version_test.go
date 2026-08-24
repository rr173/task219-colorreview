package review

import (
	"testing"

	"task219-colorreview/internal/model"
)

func TestNextVersion(t *testing.T) {
	if v := NewVersioner(nil).NextVersion(); v != 1 {
		t.Fatalf("nil current should give version 1, got %d", v)
	}
	cur := &model.ReviewConclusion{Version: 3, Status: model.ConclusionPublished}
	if v := NewVersioner(cur).NextVersion(); v != 4 {
		t.Fatalf("version 3 should give 4, got %d", v)
	}
}

func TestRequireFrozen(t *testing.T) {
	draft := &model.ReviewConclusion{Status: model.ConclusionDraft}
	if err := NewVersioner(draft).RequireFrozen(); err == nil {
		t.Fatal("draft should not allow supersede")
	}
	published := &model.ReviewConclusion{Status: model.ConclusionPublished}
	if err := NewVersioner(published).RequireFrozen(); err != nil {
		t.Fatalf("published should allow supersede: %v", err)
	}
}

func TestSupersede(t *testing.T) {
	published := &model.ReviewConclusion{BatchID: 7, Version: 2, Status: model.ConclusionPublished}
	next, err := NewVersioner(published).Supersede(model.VerdictInstrument, "new summary")
	if err != nil {
		t.Fatalf("supersede: %v", err)
	}
	if next.Version != 3 {
		t.Fatalf("expected version 3, got %d", next.Version)
	}
	if next.BatchID != 7 {
		t.Fatalf("expected batch_id 7, got %d", next.BatchID)
	}
	if next.Status != model.ConclusionDraft {
		t.Fatalf("successor should be draft, got %s", next.Status)
	}
}

func TestCanPublish(t *testing.T) {
	if !CanPublish(&model.ReviewConclusion{Status: model.ConclusionDraft}) {
		t.Fatal("draft should be publishable")
	}
	if CanPublish(&model.ReviewConclusion{Status: model.ConclusionPublished}) {
		t.Fatal("published should not be re-publishable")
	}
}
