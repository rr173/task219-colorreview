package service_test

import (
	"context"
	"testing"
	"time"

	"task219-colorreview/internal/model"
	"task219-colorreview/internal/service"
	"task219-colorreview/internal/store"
)

func TestMeasureUsesMostRecentCalibration(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer st.Close()
	svc := service.New(st)
	b, err := svc.CreateBatch(ctx, &model.DyeBatch{Code: "CAL-001", Name: "校准批次", Recipe: "R1", ColorSpace: "lab"})
	if err != nil { t.Fatal(err) }
	base := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)
	for _, c := range []*model.InstrumentCalibration{
		{InstrumentID: "I-1", CalibratedAt: base, RefL: 50, OffsetL: 1},
		{InstrumentID: "I-1", CalibratedAt: base.Add(time.Hour), RefL: 50, OffsetL: 4},
	} {
		c.RefA, c.RefB = 0, 0
		if _, err := svc.CreateCalibration(ctx, c); err != nil { t.Fatal(err) }
	}
	m, err := svc.AddMeasurePoint(ctx, &model.MeasurePoint{BatchID: b.ID, SampleNo: 1, Position: "center", ColorSpace: "lab", L: 60, MeasuredAt: base.Add(2*time.Hour)}, "I-1")
	if err != nil { t.Fatal(err) }
	if m.L != 56 {
		t.Fatalf("measurement used L=%v, want 56 from the newest +4 offset", m.L)
	}
}
