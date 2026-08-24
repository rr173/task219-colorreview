package httpapi

import (
	"context"
	"net/http"
	"time"

	"task219-colorreview/internal/colorimetry"
	"task219-colorreview/internal/model"
	"task219-colorreview/internal/service"
)

// handleSelfCheck 自检：数据库连通 + 数据量统计。
func (s *Server) handleSelfCheck(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	batches, _ := s.svc.ListBatches(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"batch_count":   len(batches),
		"component":     "sqlite",
		"go_version":    "1.26.3",
		"color_methods": []string{"cie76", "cie94", "cie2000"},
	})
}

// handleDemoImport 端到端示例导入：覆盖完整业务闭环。
func (s *Server) handleDemoImport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. 创建染色批次
	batch, err := s.svc.CreateBatch(ctx, &model.DyeBatch{
		Code:       "BATCH-DEMO-001",
		Name:       "靛蓝棉布染色批次",
		Recipe:     "R-INDIGO-42",
		ColorSpace: "lab",
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	completed := false
	defer func() {
		if !completed {
			_ = s.svc.Store().DeleteBatch(context.Background(), batch.ID)
		}
	}()

	// 2. 推进到生产中
	_, _ = s.svc.AdvanceBatch(ctx, batch.ID)
	_, _ = s.svc.AdvanceBatch(ctx, batch.ID)

	// 3. 上传浴液温度曲线
	base := time.Now().UTC().Add(-2 * time.Hour)
	tempCurve := &model.BathCurve{BatchID: batch.ID, Channel: "temperature"}
	for i := 0; i < 6; i++ {
		tempCurve.Points = append(tempCurve.Points, model.BathCurvePoint{
			Timestamp: base.Add(time.Duration(i) * 10 * time.Minute),
			Value:     60 + float64(i)*2, // 温度逐步上升
		})
	}
	if _, err := s.svc.SaveBathCurve(ctx, tempCurve); err != nil {
		writeErr(w, err)
		return
	}

	// 4. 记录仪器校准
	_, err = s.svc.CreateCalibration(ctx, &model.InstrumentCalibration{
		InstrumentID: "SPECTRO-01",
		CalibratedAt: base,
		RefL:         50, RefA: 0, RefB: 0,
		OffsetL: 0.2, OffsetA: -0.1, OffsetB: 0.1,
	})
	if err != nil {
		writeErr(w, err)
		return
	}

	// 5. 上传多点测色（含一个偏红异常点）
	type pointSeed struct {
		sample  int
		pos     string
		l, a, b float64
	}
	seeds := []pointSeed{
		{1, "center", 45, -8, -22},
		{2, "left", 44.5, -7.8, -21.5},
		{3, "right", 45.5, -8.2, -22.4},
		{4, "edge", 46, 4, -18}, // 偏红（a 大幅正向偏移）
		{5, "top", 45, -8, -21.8},
	}
	for _, seed := range seeds {
		_, err := s.svc.AddMeasurePoint(ctx, &model.MeasurePoint{
			BatchID:    batch.ID,
			SampleNo:   seed.sample,
			Position:   seed.pos,
			ColorSpace: "lab",
			L:          seed.l,
			A:          seed.a,
			B:          seed.b,
			MeasuredAt: base.Add(time.Duration(seed.sample) * time.Minute),
		}, "SPECTRO-01")
		if err != nil {
			writeErr(w, err)
			return
		}
	}

	// 6. 色差计算（目标色为靛蓝标准 Lab）
	target := colorimetry.Lab{L: 45, A: -8, B: -22}
	_, err = s.svc.ComputeColorDiff(ctx, serviceDiffRequest(batch.ID, target))
	if err != nil {
		writeErr(w, err)
		return
	}

	// 7. 提交工艺证据并确认
	ev, err := s.svc.AddEvidence(ctx, &model.ProcessEvidence{
		BatchID:     batch.ID,
		Kind:        model.EvidenceBath,
		Description: "边缘取样时段浴液 pH 瞬时漂移，导致边缘点偏红",
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	_, _ = s.svc.ConfirmEvidence(ctx, batch.ID, ev.ID)

	// 8. 剔除异常点并发布结论
	points, _ := s.svc.ListMeasurePoints(ctx, batch.ID)
	for _, p := range points {
		if p.SampleNo == 4 {
			_, _ = s.svc.RejectMeasurePoint(ctx, batch.ID, p.ID, "边缘污点，取样位置污染")
		}
	}
	conclusion, err := s.svc.CreateConclusion(ctx, &model.ReviewConclusion{
		BatchID: batch.ID,
		Verdict: model.VerdictBath,
		Summary: "边缘测色点偏红由浴液 pH 瞬时漂移引起，剔除污点后其余点色差均在容差内",
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	published, err := s.svc.PublishConclusion(ctx, conclusion.ID)
	if err != nil {
		writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"batch_id":         batch.ID,
		"batch_code":       batch.Code,
		"measure_points":   len(points),
		"conclusion_id":    published.ID,
		"conclusion_state": published.Status,
		"verdict":          published.Verdict,
	})
	completed = true
}

// serviceDiffRequest 构造色差计算请求。
func serviceDiffRequest(batchID int64, target colorimetry.Lab) service.DiffRequest {
	return service.DiffRequest{
		BatchID:   batchID,
		TargetL:   target.L,
		TargetA:   target.A,
		TargetB:   target.B,
		Method:    "cie2000",
		Tolerance: 3.0,
	}
}
