package service_test

import (
	"context"
	"testing"

	"task219-colorreview/internal/model"
	"task219-colorreview/internal/service"
	"task219-colorreview/internal/store"
)

func TestConflictedEvidenceCannotBeConfirmedAgain(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := service.New(st)
	b, err := svc.CreateBatch(ctx, &model.DyeBatch{Code: "EVD-001", Name: "证据批次", Recipe: "R1", ColorSpace: "lab"})
	if err != nil { t.Fatal(err) }
	a, err := svc.AddEvidence(ctx, &model.ProcessEvidence{BatchID: b.ID, Kind: model.EvidenceBath, Description: "浴液漂移"})
	if err != nil { t.Fatal(err) }
	b2, err := svc.AddEvidence(ctx, &model.ProcessEvidence{BatchID: b.ID, Kind: model.EvidenceSampling, Description: "边缘污染"})
	if err != nil { t.Fatal(err) }
	if _, err = svc.ConfirmEvidence(ctx, b.ID, a.ID); err != nil { t.Fatal(err) }
	if _, err = svc.ConfirmEvidence(ctx, b.ID, b2.ID); err != nil { t.Fatal(err) }
	if _, err = svc.ConfirmEvidence(ctx, b.ID, a.ID); err == nil {
		t.Fatal("conflicted evidence was confirmed again")
	}
	e, err := st.GetEvidence(ctx, a.ID)
	if err != nil { t.Fatal(err) }
	if e.Status != model.EvidenceConflict { t.Fatalf("status=%s, want conflict", e.Status) }
}
