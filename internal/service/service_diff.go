package service

import (
	"context"
	"fmt"

	"task219-colorreview/internal/colorimetry"
	"task219-colorreview/internal/model"
)

// DiffRequest 色差计算请求参数。
type DiffRequest struct {
	BatchID   int64
	TargetL   float64
	TargetA   float64
	TargetB   float64
	Method    string // cie76 / cie94 / cie2000
	Tolerance float64
}

// ComputeColorDiff 对某批次全部有效测色点做色差计算与容差判定。
// 超容差的点会被标记为异常（anomaly），并把色差回填到测色点记录。
func (s *Service) ComputeColorDiff(ctx context.Context, req DiffRequest) (*model.ColorDiffSummary, error) {
	target, err := colorimetry.ParseLab(req.TargetL, req.TargetA, req.TargetB)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", model.ErrInvalidArgument, err)
	}
	if !colorimetry.ValidMethod(req.Method) {
		return nil, fmt.Errorf("%w: unknown method %q", model.ErrInvalidArgument, req.Method)
	}
	if req.Tolerance < 0 {
		return nil, fmt.Errorf("%w: tolerance must be non-negative", model.ErrInvalidArgument)
	}

	points, err := s.store.ListMeasurePoints(ctx, req.BatchID)
	if err != nil {
		return nil, err
	}

	// 剔除是不可逆终局状态：剔除点不参与色差计算，后续重算既不回填
	// 色差、也不改写其状态或剔除原因，避免把已剔除点重新标记为有效或异常。
	var valid []colorimetry.PointSample
	for _, p := range points {
		if p.Status == model.MeasureRejected {
			continue
		}
		valid = append(valid, colorimetry.PointSample{SampleNo: p.SampleNo, Position: p.Position, L: p.L, A: p.A, B: p.B})
	}

	method := colorimetry.DiffMethod(req.Method)
	report := colorimetry.Evaluate(method, target, req.Tolerance, valid)

	// 回填色差并标记异常点。剔除点不在 report 中，此处再次显式跳过，
	// 确保其状态与 reject_reason 绝不被重算覆盖。
	bySample := make(map[int]float64)
	for _, pd := range report.Points {
		bySample[pd.SampleNo] = pd.DeltaE
	}
	for _, p := range points {
		if p.Status == model.MeasureRejected {
			continue
		}
		d, ok := bySample[p.SampleNo]
		if !ok {
			continue
		}
		_ = s.store.SetMeasureDeltaE(ctx, p.ID, d)
		if d > req.Tolerance {
			_ = s.store.SetMeasureStatus(ctx, p.ID, model.MeasureAnomaly)
		} else {
			_ = s.store.SetMeasureStatus(ctx, p.ID, model.MeasureValid)
		}
	}

	summary := &model.ColorDiffSummary{
		BatchID:      req.BatchID,
		Target:       [3]float64{target.L, target.A, target.B},
		Method:       req.Method,
		Tolerance:    req.Tolerance,
		MaxDeltaE:    report.MaxDeltaE,
		MeanDeltaE:   report.MeanDeltaE,
		AnomalyCount: report.Exceeding,
		ValidCount:   report.Within,
	}
	for _, pd := range report.Points {
		summary.Points = append(summary.Points, model.ColorDiffResult{
			SampleNo:        pd.SampleNo,
			Position:        pd.Position,
			DeltaE76:        colorimetry.DeltaE76(target, colorimetry.Lab{L: pd.L, A: pd.A, B: pd.B}),
			DeltaE2000:      colorimetry.DeltaE2000(target, colorimetry.Lab{L: pd.L, A: pd.A, B: pd.B}),
			WithinTolerance: pd.Within,
		})
	}
	return summary, nil
}
