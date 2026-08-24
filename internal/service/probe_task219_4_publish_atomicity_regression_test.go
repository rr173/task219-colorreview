package service_test

import (
	"context"
	"testing"

	"task219-colorreview/internal/model"
	"task219-colorreview/internal/service"
	"task219-colorreview/internal/store"
)

func TestPublishRollsBackWhenSnapshotFails(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := service.New(st)
	b, err := svc.CreateBatch(ctx, &model.DyeBatch{Code: "PUB-001", Name: "发布批次", Recipe: "R1", ColorSpace: "lab"})
	if err != nil { t.Fatal(err) }
	c, err := svc.CreateConclusion(ctx, &model.ReviewConclusion{BatchID: b.ID, Verdict: model.VerdictBath, Summary: "浴液证据"})
	if err != nil { t.Fatal(err) }
	if _, err := st.DB().ExecContext(ctx, `CREATE TRIGGER fail_snapshot BEFORE INSERT ON conclusion_versions BEGIN SELECT RAISE(ABORT, 'snapshot blocked'); END`); err != nil { t.Fatal(err) }
	if _, err = svc.PublishConclusion(ctx, c.ID); err == nil { t.Fatal("publish unexpectedly succeeded") }
	restored, err := svc.GetConclusion(ctx, b.ID)
	if err != nil { t.Fatal(err) }
	if restored.Status != model.ConclusionDraft { t.Fatalf("status=%s, want draft after rollback", restored.Status) }
	versions, err := svc.ListConclusionVersions(ctx, b.ID)
	if err != nil { t.Fatal(err) }
	if len(versions) != 0 { t.Fatalf("versions=%d, want 0", len(versions)) }
}
