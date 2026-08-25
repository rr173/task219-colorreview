package service_test

import (
	"context"
	"testing"
	"time"

	"task219-colorreview/internal/model"
	"task219-colorreview/internal/service"
	"task219-colorreview/internal/store"
)

func TestAllRejectedPointsReturnEmptyDiff(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := service.New(st)
	b, err := svc.CreateBatch(ctx, &model.DyeBatch{Code: "DIFF-ALL-REJ", Name: "全剔除批次", Recipe: "R1", ColorSpace: "lab"})
	if err != nil { t.Fatal(err) }
	p, err := svc.AddMeasurePoint(ctx, &model.MeasurePoint{BatchID: b.ID, SampleNo: 1, Position: "edge", ColorSpace: "lab", L: 50, MeasuredAt: time.Now()}, "")
	if err != nil { t.Fatal(err) }
	if _, err = svc.RejectMeasurePoint(ctx, b.ID, p.ID, "污点"); err != nil { t.Fatal(err) }
	summary, err := svc.ComputeColorDiff(ctx, service.DiffRequest{BatchID: b.ID, TargetL: 50, TargetA: 0, TargetB: 0, Method: "cie76", Tolerance: 3})
	if err != nil { t.Fatal(err) }
	if summary.AnomalyCount != 0 || len(summary.Points) != 0 { t.Fatalf("unexpected summary: %+v", summary) }
}
