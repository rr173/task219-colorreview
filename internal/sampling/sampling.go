// Package sampling 负责布样测色点的采集校验与测色仪校准。
package sampling

import (
	"fmt"
	"time"

	"task219-colorreview/internal/colorimetry"
	"task219-colorreview/internal/model"
)

// ValidatePoint 校验测色点输入。
func ValidatePoint(m *model.MeasurePoint) error {
	if m.BatchID <= 0 {
		return fmt.Errorf("%w: batch_id required", model.ErrInvalidArgument)
	}
	if m.SampleNo <= 0 {
		return fmt.Errorf("%w: sample_no must be positive", model.ErrInvalidArgument)
	}
	if m.ColorSpace == "" {
		return model.ErrColorSpaceMissing
	}
	if m.ColorSpace != "lab" {
		return fmt.Errorf("%w: unsupported color space %q", model.ErrInvalidArgument, m.ColorSpace)
	}
	if _, err := colorimetry.ParseLab(m.L, m.A, m.B); err != nil {
		return fmt.Errorf("%w: %v", model.ErrInvalidArgument, err)
	}
	if m.MeasuredAt.IsZero() {
		return fmt.Errorf("%w: measured_at required", model.ErrInvalidArgument)
	}
	return nil
}

// ApplyCalibration 用仪器校准偏移修正测色点的 Lab 值。
// 业务规则：实测值减去仪器偏移，得到接近标准色板的校正值。
func ApplyCalibration(l, a, b float64, cal *model.InstrumentCalibration) (float64, float64, float64) {
	if cal == nil {
		return l, a, b
	}
	return l - cal.OffsetL, a - cal.OffsetA, b - cal.OffsetB
}

// ValidateCalibration 校验校准记录：参考色板 Lab 与偏移量需合法。
func ValidateCalibration(c *model.InstrumentCalibration) error {
	if c.InstrumentID == "" {
		return fmt.Errorf("%w: instrument_id required", model.ErrInvalidArgument)
	}
	if _, err := colorimetry.ParseLab(c.RefL, c.RefA, c.RefB); err != nil {
		return fmt.Errorf("%w: %v", model.ErrInvalidArgument, err)
	}
	if c.CalibratedAt.IsZero() {
		return fmt.Errorf("%w: calibrated_at required", model.ErrInvalidArgument)
	}
	return nil
}

// ComputeOffsets 根据标准色板参考值与实测值计算仪器偏移。
func ComputeOffsets(refL, refA, refB, measuredL, measuredA, measuredB float64) (float64, float64, float64) {
	return measuredL - refL, measuredA - refA, measuredB - refB
}

// EnsureMonotonicTime 校验一组测量时间是否严格递增，防止时间轴倒置。
func EnsureMonotonicTime(times []time.Time) error {
	for i := 1; i < len(times); i++ {
		if !times[i].After(times[i-1]) {
			return model.ErrTimeInverted
		}
	}
	return nil
}
