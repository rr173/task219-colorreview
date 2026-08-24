package colorimetry

import (
	"sort"
)

// DiffMethod 色差度量方法。
type DiffMethod string

const (
	Method76   DiffMethod = "cie76"
	Method94   DiffMethod = "cie94"
	Method2000 DiffMethod = "cie2000"
)

// ValidMethod 校验色差方法是否受支持。
func ValidMethod(m string) bool {
	switch DiffMethod(m) {
	case Method76, Method94, Method2000:
		return true
	}
	return false
}

// DeltaE 按指定方法计算两 Lab 色差。
func DeltaE(method DiffMethod, a, b Lab) float64 {
	switch method {
	case Method94:
		return DeltaE94(a, b)
	case Method2000:
		return DeltaE2000(a, b)
	default:
		return DeltaE76(a, b)
	}
}

// PointDiff 单个测色点与目标色的色差。
type PointDiff struct {
	SampleNo    int
	Position    string
	L, A, B     float64
	DeltaE      float64
	Within      bool
}

// ToleranceReport 一批测色点的容差判定汇总。
type ToleranceReport struct {
	Target     Lab
	Method     DiffMethod
	Tolerance  float64
	Points     []PointDiff
	MaxDeltaE  float64
	MeanDeltaE float64
	Within     int
	Exceeding  int
}

// PointSample 参与容差判定的单个测色点输入。
type PointSample struct {
	SampleNo int
	Position string
	L, A, B  float64
}

// Evaluate 对一组测色点按目标色和容差做判定。
func Evaluate(method DiffMethod, target Lab, tolerance float64, points []PointSample) ToleranceReport {
	rep := ToleranceReport{Target: target, Method: method, Tolerance: tolerance}
	if len(points) == 0 {
		return rep
	}
	var sum float64
	for _, p := range points {
		d := DeltaE(method, target, Lab{L: p.L, A: p.A, B: p.B})
		pd := PointDiff{SampleNo: p.SampleNo, Position: p.Position, L: p.L, A: p.A, B: p.B, DeltaE: d, Within: d <= tolerance}
		if d > rep.MaxDeltaE {
			rep.MaxDeltaE = d
		}
		sum += d
		if pd.Within {
			rep.Within++
		} else {
			rep.Exceeding++
		}
		rep.Points = append(rep.Points, pd)
	}
	rep.MeanDeltaE = sum / float64(len(points))
	sort.SliceStable(rep.Points, func(i, j int) bool { return rep.Points[i].DeltaE > rep.Points[j].DeltaE })
	return rep
}
