package service_test

import (
	"context"
	"testing"

	"task219-colorreview/internal/model"
	"task219-colorreview/internal/service"
	"task219-colorreview/internal/store"
)

func TestSupersedeCreatesNextConclusionVersion(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := service.New(st)
	b, err := svc.CreateBatch(ctx, &model.DyeBatch{Code: "VER-001", Name: "版本批次", Recipe: "R1", ColorSpace: "lab"})
	if err != nil { t.Fatal(err) }
	c, err := svc.CreateConclusion(ctx, &model.ReviewConclusion{BatchID: b.ID, Verdict: model.VerdictBath, Summary: "初版"})
	if err != nil { t.Fatal(err) }
	if _, err = svc.PublishConclusion(ctx, c.ID); err != nil { t.Fatal(err) }
	next, err := svc.SupersedeConclusion(ctx, c.ID, model.VerdictSampling, "替代版")
	if err != nil { t.Fatal(err) }
	if next.Version != 2 { t.Fatalf("version=%d, want 2", next.Version) }
	versions, err := svc.ListConclusionVersions(ctx, b.ID)
	if err != nil { t.Fatal(err) }
	if len(versions) != 1 || versions[0]["version"] != 1 { t.Fatalf("snapshots=%v", versions) }
}
