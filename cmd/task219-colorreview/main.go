// Command task219-colorreview 是纺织染色批次色差证据复核台的入口。
//
// 用法：
//
//	./colorreview --addr :8080 --db colorreview.db   # 启动 HTTP 服务
//	./colorreview --smoke-test                        # 端到端自检（Docker 判据）
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"task219-colorreview/internal/colorimetry"
	"task219-colorreview/internal/httpapi"
	"task219-colorreview/internal/model"
	"task219-colorreview/internal/service"
	"task219-colorreview/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "colorreview.db", "SQLite database path")
	smoke := flag.Bool("smoke-test", false, "run end-to-end self test and exit")
	flag.Parse()

	if *smoke {
		if err := runSmoke(); err != nil {
			log.Fatalf("smoke test failed: %v", err)
		}
		fmt.Println("SMOKE TEST PASSED")
		return
	}

	s, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer s.Close()

	svc := service.New(s)
	server := httpapi.New(svc)

	log.Printf("colorreview listening on %s (db=%s)", *addr, *dbPath)
	if err := http.ListenAndServe(*addr, server.Handler()); err != nil {
		log.Fatalf("http server: %v", err)
	}
}

// runSmoke 完整复现端到端场景：染色批次 → 浴液曲线 → 多点测色 → 色差计算 →
// 异常点剔除 → 证据确认 → 结论发布 → 关闭重开验证持久化恢复。
func runSmoke() error {
	dir, err := os.MkdirTemp("", "colorreview-smoke-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	dbPath := filepath.Join(dir, "smoke.db")

	// 第一段：构建场景。
	ctx := context.Background()
	s, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	svc := service.New(s)

	batch, err := svc.CreateBatch(ctx, &model.DyeBatch{
		Code:       "BATCH-INDIGO-001",
		Name:       "靛蓝棉布染色批次",
		Recipe:     "R-INDIGO-42",
		ColorSpace: "lab",
	})
	if err != nil {
		return err
	}
	// 推进到待复核。
	if _, err = svc.AdvanceBatch(ctx, batch.ID); err != nil {
		return err
	}
	if _, err = svc.AdvanceBatch(ctx, batch.ID); err != nil {
		return err
	}

	base := time.Now().UTC().Add(-2 * time.Hour)
	// 浴液温度曲线。
	tempCurve := &model.BathCurve{BatchID: batch.ID, Channel: "temperature"}
	for i := 0; i < 6; i++ {
		tempCurve.Points = append(tempCurve.Points, model.BathCurvePoint{
			Timestamp: base.Add(time.Duration(i) * 10 * time.Minute),
			Value:     60 + float64(i)*2,
		})
	}
	if _, err = svc.SaveBathCurve(ctx, tempCurve); err != nil {
		return err
	}

	// 仪器校准。
	if _, err = svc.CreateCalibration(ctx, &model.InstrumentCalibration{
		InstrumentID: "SPECTRO-01",
		CalibratedAt: base,
		RefL:         50, RefA: 0, RefB: 0,
		OffsetL: 0.2, OffsetA: -0.1, OffsetB: 0.1,
	}); err != nil {
		return err
	}

	// 多点测色（第 4 点为偏红异常）。
	seeds := []struct {
		sample int
		pos    string
		l, a, b float64
	}{
		{1, "center", 45, -8, -22},
		{2, "left", 44.5, -7.8, -21.5},
		{3, "right", 45.5, -8.2, -22.4},
		{4, "edge", 46, 4, -18},
		{5, "top", 45, -8, -21.8},
	}
	for _, seed := range seeds {
		if _, err = svc.AddMeasurePoint(ctx, &model.MeasurePoint{
			BatchID:    batch.ID,
			SampleNo:   seed.sample,
			Position:   seed.pos,
			ColorSpace: "lab",
			L:          seed.l,
			A:          seed.a,
			B:          seed.b,
			MeasuredAt: base.Add(time.Duration(seed.sample) * time.Minute),
		}, "SPECTRO-01"); err != nil {
			return err
		}
	}

	// 色差计算：目标色为靛蓝标准。
	target := colorimetry.Lab{L: 45, A: -8, B: -22}
	summary, err := svc.ComputeColorDiff(ctx, service.DiffRequest{
		BatchID:   batch.ID,
		TargetL:   target.L,
		TargetA:   target.A,
		TargetB:   target.B,
		Method:    "cie2000",
		Tolerance: 3.0,
	})
	if err != nil {
		return err
	}
	if summary.AnomalyCount != 1 {
		return fmt.Errorf("expected 1 anomaly point, got %d", summary.AnomalyCount)
	}

	// 提交并确认证据。
	ev, err := svc.AddEvidence(ctx, &model.ProcessEvidence{
		BatchID:     batch.ID,
		Kind:        model.EvidenceBath,
		Description: "边缘取样时段浴液 pH 瞬时漂移，导致边缘点偏红",
	})
	if err != nil {
		return err
	}
	if _, err = svc.ConfirmEvidence(ctx, batch.ID, ev.ID); err != nil {
		return err
	}

	// 剔除异常点。
	points, err := svc.ListMeasurePoints(ctx, batch.ID)
	if err != nil {
		return err
	}
	for _, p := range points {
		if p.SampleNo == 4 {
			if _, err = svc.RejectMeasurePoint(ctx, batch.ID, p.ID, "边缘污点，取样位置污染"); err != nil {
				return err
			}
		}
	}

	// 创建并发布结论。
	conclusion, err := svc.CreateConclusion(ctx, &model.ReviewConclusion{
		BatchID: batch.ID,
		Verdict: model.VerdictBath,
		Summary: "边缘测色点偏红由浴液 pH 瞬时漂移引起，剔除污点后其余点色差均在容差内",
	})
	if err != nil {
		return err
	}
	published, err := svc.PublishConclusion(ctx, conclusion.ID)
	if err != nil {
		return err
	}

	// 关闭数据库，验证重启恢复。
	if err = s.Close(); err != nil {
		return err
	}

	s2, err := store.Open(dbPath)
	if err != nil {
		return err
	}
	defer s2.Close()
	svc2 := service.New(s2)

	restoredBatch, err := svc2.GetBatch(ctx, batch.ID)
	if err != nil {
		return err
	}
	if restoredBatch.Status != model.BatchPendingReview {
		return fmt.Errorf("batch status not restored: %s", restoredBatch.Status)
	}

	restoredPoints, err := svc2.ListMeasurePoints(ctx, batch.ID)
	if err != nil {
		return err
	}
	if len(restoredPoints) != 5 {
		return fmt.Errorf("expected 5 measure points after restart, got %d", len(restoredPoints))
	}

	restoredConclusion, err := svc2.GetConclusion(ctx, batch.ID)
	if err != nil {
		return err
	}
	if restoredConclusion.Status != model.ConclusionPublished {
		return fmt.Errorf("conclusion not restored: %s", restoredConclusion.Status)
	}
	if restoredConclusion.Version != published.Version {
		return fmt.Errorf("conclusion version mismatch after restart")
	}

	fmt.Printf("smoke: batch=%d code=%s points=%d anomalies=%d verdict=%s conclusion=v%d status=%s\n",
		batch.ID, batch.Code, len(restoredPoints), summary.AnomalyCount,
		published.Verdict, published.Version, published.Status)
	return nil
}
