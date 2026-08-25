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
// 任一步失败时报告失败并补偿回滚已创建的批次、曲线、测色点、证据、结论
// 与独立写入的校准记录，绝不留下一次失败导入产生的残缺批次。
func (s *Server) handleDemoImport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 已创建的资源：成功写入后才赋值，回滚时按相反顺序撤销。
	var (
		batchID       int64
		calibrationID int64
		committed     bool
	)
	// 回滚在独立、未取消的上下文中执行：导入请求的 ctx 可能已超时或被取消，
	// 补偿清理必须仍能落库。补偿错误不覆盖原始导入错误，仅以日志形式保留。
	rollback := func() {
		if calibrationID != 0 {
			rctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = s.svc.DeleteCalibration(rctx, calibrationID)
		}
		if batchID != 0 {
			rctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = s.svc.Store().DeleteBatch(rctx, batchID)
		}
	}
	defer func() {
		if !committed {
			rollback()
		}
	}()

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
	batchID = batch.ID

	// 2. 推进到待复核
	if _, err := s.svc.AdvanceBatch(ctx, batch.ID); err != nil {
		writeErr(w, err)
		return
	}
	if _, err := s.svc.AdvanceBatch(ctx, batch.ID); err != nil {
		writeErr(w, err)
		return
	}

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

	// 4. 记录仪器校准（失败即终止并回滚，不再吞掉错误）
	cal, err := s.svc.CreateCalibration(ctx, &model.InstrumentCalibration{
		InstrumentID: "SPECTRO-01",
		CalibratedAt: base,
		RefL:         50, RefA: 0, RefB: 0,
		OffsetL: 0.2, OffsetA: -0.1, OffsetB: 0.1,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	calibrationID = cal.ID

	// 5. 上传多点测色（含一个偏红异常点）
	type pointSeed struct {
		sample int
		pos    string
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
	if _, err := s.svc.ComputeColorDiff(ctx, serviceDiffRequest(batch.ID, target)); err != nil {
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
	if _, err := s.svc.ConfirmEvidence(ctx, batch.ID, ev.ID); err != nil {
		writeErr(w, err)
		return
	}

	// 8. 剔除异常点并发布结论
	points, err := s.svc.ListMeasurePoints(ctx, batch.ID)
	if err != nil {
		writeErr(w, err)
		return
	}
	for _, p := range points {
		if p.SampleNo == 4 {
			if _, err := s.svc.RejectMeasurePoint(ctx, batch.ID, p.ID, "边缘污点，取样位置污染"); err != nil {
				writeErr(w, err)
				return
			}
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

	committed = true
	writeJSON(w, http.StatusCreated, map[string]any{
		"batch_id":         batch.ID,
		"batch_code":       batch.Code,
		"measure_points":   len(points),
		"conclusion_id":    published.ID,
		"conclusion_state": published.Status,
		"verdict":          published.Verdict,
	})
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
