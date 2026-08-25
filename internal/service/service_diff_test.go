package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"task219-colorreview/internal/model"
	"task219-colorreview/internal/store"
)

// rejectedPointsTeardown 封装「剔除点重算后状态与原因不可逆」的断言。
// frozenDeltaE 是该点被剔除时刻的色差值：重算后必须保持不变，证明剔除点
// 的色差也未被回填覆盖。
func rejectedPointsTeardown(t *testing.T, svc *Service, batchID int64, expectSample int, expectReason string, frozenDeltaE float64) {
	t.Helper()
	points, err := svc.ListMeasurePoints(context.Background(), batchID)
	if err != nil {
		t.Fatalf("list points: %v", err)
	}
	for _, p := range points {
		if p.SampleNo != expectSample {
			continue
		}
		if p.Status != model.MeasureRejected {
			t.Fatalf("rejected point %d status = %s, want %s (剔除不可被重算复活)", expectSample, p.Status, model.MeasureRejected)
		}
		if p.RejectReason != expectReason {
			t.Fatalf("rejected point %d reason = %q, want %q (剔除原因不可被重算清空)", expectSample, p.RejectReason, expectReason)
		}
		if p.DeltaE != frozenDeltaE {
			t.Fatalf("rejected point %d delta_e = %v, want %v (剔除点色差不得被重算覆盖)", expectSample, p.DeltaE, frozenDeltaE)
		}
		return
	}
	t.Fatalf("sample %d not found among %d points", expectSample, len(points))
}

// TestRejectedMeasurePointSurvivesColorDiffRecompute 验证剔除的不可逆语义：
// 先色差计算 → 剔除某点 → 再次色差计算，剔除点必须保持 rejected 且
// reject_reason 不被清空，也不被重新标记为 valid/anomaly。
func TestRejectedMeasurePointSurvivesColorDiffRecompute(t *testing.T) {
	ctx := context.Background()
	s, err := store.Open(filepath.Join(t.TempDir(), "diff-recompute.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	svc := New(s)

	batch, err := svc.CreateBatch(ctx, &model.DyeBatch{
		Code: "BATCH-RECOMPUTE-001", Name: "recompute", Recipe: "R",
		ColorSpace: "lab",
	})
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}
	if _, err = svc.AdvanceBatch(ctx, batch.ID); err != nil {
		t.Fatalf("advance: %v", err)
	}
	if _, err = svc.AdvanceBatch(ctx, batch.ID); err != nil {
		t.Fatalf("advance: %v", err)
	}

	base := time.Now().UTC().Add(-2 * time.Hour)
	seeds := []struct {
		sample    int
		pos       string
		l, a, b   float64
	}{
		{1, "center", 45, -8, -22},
		{2, "edge", 46, 4, -18}, // 偏红异常点
	}
	for _, seed := range seeds {
		if _, err = svc.AddMeasurePoint(ctx, &model.MeasurePoint{
			BatchID: batch.ID, SampleNo: seed.sample, Position: seed.pos,
			ColorSpace: "lab", L: seed.l, A: seed.a, B: seed.b,
			MeasuredAt: base.Add(time.Duration(seed.sample) * time.Minute),
		}, ""); err != nil {
			t.Fatalf("add point %d: %v", seed.sample, err)
		}
	}

	// 目标色与点 1 接近，点 2 偏红将超容差。
	diffReq := DiffRequest{
		BatchID: batch.ID, TargetL: 45, TargetA: -8, TargetB: -22,
		Method: "cie2000", Tolerance: 3.0,
	}
	if _, err = svc.ComputeColorDiff(ctx, diffReq); err != nil {
		t.Fatalf("first diff: %v", err)
	}

	// 剔除偏红点。
	const reason = "边缘污点，取样位置污染"
	points, err := svc.ListMeasurePoints(ctx, batch.ID)
	if err != nil {
		t.Fatalf("list points: %v", err)
	}
	var frozenDeltaE float64
	for _, p := range points {
		if p.SampleNo == 2 {
			frozenDeltaE = p.DeltaE
			if _, err = svc.RejectMeasurePoint(ctx, batch.ID, p.ID, reason); err != nil {
				t.Fatalf("reject point: %v", err)
			}
		}
	}

	// 第二次色差计算：剔除点不得被重新标记或清原因，色差也不得被覆盖。
	// 关键判据：换一个不同的目标色重算——若剔除点未被过滤，其 delta_e 会随
	// 新目标色变化，从而暴露「重算覆盖剔除点色差」的缺陷。
	recomputeReq := diffReq
	recomputeReq.TargetL = 70
	recomputeReq.TargetA = 20
	recomputeReq.TargetB = 30
	if _, err = svc.ComputeColorDiff(ctx, recomputeReq); err != nil {
		t.Fatalf("second diff: %v", err)
	}
	rejectedPointsTeardown(t, svc, batch.ID, 2, reason, frozenDeltaE)

	// 第三次再用更窄容差重算，进一步压测状态与原因的不可逆语义。
	recomputeReq.Tolerance = 0.01
	if _, err = svc.ComputeColorDiff(ctx, recomputeReq); err != nil {
		t.Fatalf("third diff: %v", err)
	}
	rejectedPointsTeardown(t, svc, batch.ID, 2, reason, frozenDeltaE)
}
