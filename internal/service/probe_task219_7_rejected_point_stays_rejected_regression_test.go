package service_test

import (
	"context"
	"testing"
	"time"

	"task219-colorreview/internal/model"
	"task219-colorreview/internal/service"
	"task219-colorreview/internal/store"
)

func TestRejectedPointIsExcludedFromLaterDiffRuns(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := service.New(st)
	b, err := svc.CreateBatch(ctx, &model.DyeBatch{Code: "DIFF-REJ", Name: "剔除批次", Recipe: "R1", ColorSpace: "lab"})
	if err != nil { t.Fatal(err) }
	p, err := svc.AddMeasurePoint(ctx, &model.MeasurePoint{BatchID: b.ID, SampleNo: 1, Position: "edge", ColorSpace: "lab", L: 70, MeasuredAt: time.Now()}, "")
	if err != nil { t.Fatal(err) }
	if _, err = svc.ComputeColorDiff(ctx, service.DiffRequest{BatchID: b.ID, TargetL: 50, TargetA: 0, TargetB: 0, Method: "cie76", Tolerance: 3}); err != nil { t.Fatal(err) }
	if _, err = svc.RejectMeasurePoint(ctx, b.ID, p.ID, "布面污点"); err != nil { t.Fatal(err) }
	if _, err = svc.ComputeColorDiff(ctx, service.DiffRequest{BatchID: b.ID, TargetL: 50, TargetA: 0, TargetB: 0, Method: "cie76", Tolerance: 3}); err != nil { t.Fatal(err) }
	restored, err := st.GetMeasurePoint(ctx, p.ID)
	if err != nil { t.Fatal(err) }
	if restored.Status != model.MeasureRejected || restored.RejectReason != "布面污点" { t.Fatalf("point=%+v, want rejected with reason", restored) }
}
