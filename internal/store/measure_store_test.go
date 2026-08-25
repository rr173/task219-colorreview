package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"task219-colorreview/internal/model"
)

// TestSetMeasureStatusPreservesRejected 验证 SetMeasureStatus 是剔除状态的
// 写入级护栏：对已 rejected 的行调用，状态与 reject_reason 都不得被改动。
func TestSetMeasureStatusPreservesRejected(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "reject-guard.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	b, err := s.CreateBatch(ctx, &model.DyeBatch{Code: "GUARD-1", Name: "g", Recipe: "r"})
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}
	pt, err := s.CreateMeasurePoint(ctx, &model.MeasurePoint{
		BatchID: b.ID, SampleNo: 1, Position: "edge", ColorSpace: "lab",
		L: 45, A: -8, B: -22, MeasuredAt: time.Unix(100, 0).UTC(),
	})
	if err != nil {
		t.Fatalf("create point: %v", err)
	}
	const reason = "污点剔除"
	if err := s.RejectMeasurePoint(ctx, pt.ID, reason); err != nil {
		t.Fatalf("reject point: %v", err)
	}

	// 直接对 rejected 行改状态：无论传入 valid 还是 anomaly 都应无效。
	if err := s.SetMeasureStatus(ctx, pt.ID, model.MeasureValid); err != nil {
		t.Fatalf("set status valid: %v", err)
	}
	if err := s.SetMeasureStatus(ctx, pt.ID, model.MeasureAnomaly); err != nil {
		t.Fatalf("set status anomaly: %v", err)
	}

	got, err := s.GetMeasurePoint(ctx, pt.ID)
	if err != nil {
		t.Fatalf("get point: %v", err)
	}
	if got.Status != model.MeasureRejected {
		t.Fatalf("status = %s, want rejected (SetMeasureStatus 不得复活剔除点)", got.Status)
	}
	if got.RejectReason != reason {
		t.Fatalf("reject_reason = %q, want %q (剔除原因不得被清空)", got.RejectReason, reason)
	}

	// 对照：正常 pending 点可以正常改状态。
	pt2, err := s.CreateMeasurePoint(ctx, &model.MeasurePoint{
		BatchID: b.ID, SampleNo: 2, Position: "center", ColorSpace: "lab",
		L: 45, A: -8, B: -22, MeasuredAt: time.Unix(101, 0).UTC(),
	})
	if err != nil {
		t.Fatalf("create point 2: %v", err)
	}
	if err := s.SetMeasureStatus(ctx, pt2.ID, model.MeasureAnomaly); err != nil {
		t.Fatalf("set status anomaly on pending point: %v", err)
	}
	got2, _ := s.GetMeasurePoint(ctx, pt2.ID)
	if got2.Status != model.MeasureAnomaly {
		t.Fatalf("pending point status = %s, want anomaly", got2.Status)
	}
}
