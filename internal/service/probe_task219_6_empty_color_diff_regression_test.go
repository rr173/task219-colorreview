package service_test

import (
	"context"
	"testing"

	"task219-colorreview/internal/model"
	"task219-colorreview/internal/service"
	"task219-colorreview/internal/store"
)

func TestEmptyColorDiffReturnsEmptySummary(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := service.New(st)
	b, err := svc.CreateBatch(ctx, &model.DyeBatch{Code: "DIFF-EMPTY", Name: "空测色批次", Recipe: "R1", ColorSpace: "lab"})
	if err != nil { t.Fatal(err) }
	summary, err := svc.ComputeColorDiff(ctx, service.DiffRequest{BatchID: b.ID, TargetL: 50, TargetA: 0, TargetB: 0, Method: "cie2000", Tolerance: 3})
	if err != nil { t.Fatal(err) }
	if summary.AnomalyCount != 0 || len(summary.Points) != 0 { t.Fatalf("unexpected summary: %+v", summary) }
}
