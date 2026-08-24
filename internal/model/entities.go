// Package model 定义纺织染色批次色差证据复核台的领域实体与状态机。
package model

import "time"

// BatchStatus 染色批次生命周期状态。
type BatchStatus string

const (
	BatchRecipe       BatchStatus = "recipe"        // 配方中
	BatchInProduction BatchStatus = "in_production" // 生产中
	BatchPendingReview BatchStatus = "pending_review" // 待复核
	BatchConfirmed    BatchStatus = "confirmed"     // 已确认
	BatchSealed       BatchStatus = "sealed"        // 封存
)

// ValidBatchTransitions 定义批次状态机的合法流转。
var ValidBatchTransitions = map[BatchStatus][]BatchStatus{
	BatchRecipe:        {BatchInProduction},
	BatchInProduction:  {BatchPendingReview},
	BatchPendingReview: {BatchConfirmed, BatchSealed},
	BatchConfirmed:     {BatchSealed},
	BatchSealed:        {},
}

// MeasureStatus 测色点状态。
type MeasureStatus string

const (
	MeasurePending   MeasureStatus = "pending"   // 待校准
	MeasureValid     MeasureStatus = "valid"     // 有效
	MeasureAnomaly   MeasureStatus = "anomaly"   // 异常
	MeasureRejected  MeasureStatus = "rejected"  // 剔除
)

// EvidenceStatus 工艺证据状态。
type EvidenceStatus string

const (
	EvidencePending   EvidenceStatus = "pending"   // 待关联
	EvidenceLinked    EvidenceStatus = "linked"    // 已关联
	EvidenceConflict  EvidenceStatus = "conflict"  // 冲突
	EvidenceConfirmed EvidenceStatus = "confirmed" // 确认
)

// ConclusionStatus 复核结论状态。
type ConclusionStatus string

const (
	ConclusionDraft     ConclusionStatus = "draft"      // 草稿
	ConclusionPending   ConclusionStatus = "pending"    // 待确认
	ConclusionPublished ConclusionStatus = "published"  // 发布
	ConclusionSuperseded ConclusionStatus = "superseded" // 替代
)

// ValidConclusionTransitions 定义结论状态机合法流转。
var ValidConclusionTransitions = map[ConclusionStatus][]ConclusionStatus{
	ConclusionDraft:     {ConclusionPending, ConclusionPublished},
	ConclusionPending:   {ConclusionPublished},
	ConclusionPublished: {ConclusionSuperseded},
	ConclusionSuperseded: {},
}

// DyeBatch 染色批次。
type DyeBatch struct {
	ID        int64       `json:"id"`
	Code      string      `json:"code"`       // 批次号
	Name      string      `json:"name"`       // 批次名称
	Recipe    string      `json:"recipe"`     // 配方标识
	ColorSpace string     `json:"color_space"` // 色彩空间声明
	Status    BatchStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// BathCurvePoint 浴液曲线单个采样点（温度或 pH）。
type BathCurvePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// BathCurve 浴液曲线（温度或 pH 通道）。
type BathCurve struct {
	ID        int64            `json:"id"`
	BatchID   int64            `json:"batch_id"`
	Channel   string           `json:"channel"` // temperature / ph
	Points    []BathCurvePoint `json:"points"`
	CreatedAt time.Time        `json:"created_at"`
}

// MeasurePoint 布样测色点。
type MeasurePoint struct {
	ID          int64         `json:"id"`
	BatchID     int64         `json:"batch_id"`
	SampleNo    int           `json:"sample_no"`    // 样本序号
	Position    string        `json:"position"`     // 布样位置
	ColorSpace  string        `json:"color_space"`  // 色彩空间
	L           float64       `json:"l"`
	A           float64       `json:"a"`
	B           float64       `json:"b"`
	MeasuredAt  time.Time     `json:"measured_at"`  // 测量时间
	Status      MeasureStatus `json:"status"`
	RejectReason string       `json:"reject_reason,omitempty"`
	DeltaE      float64       `json:"delta_e,omitempty"` // 与目标色的色差
	CreatedAt   time.Time     `json:"created_at"`
}

// InstrumentCalibration 测色仪校准记录。
type InstrumentCalibration struct {
	ID           int64     `json:"id"`
	InstrumentID string    `json:"instrument_id"`
	CalibratedAt time.Time `json:"calibrated_at"`
	RefL         float64   `json:"ref_l"` // 标准色板 L
	RefA         float64   `json:"ref_a"`
	RefB         float64   `json:"ref_b"`
	OffsetL      float64   `json:"offset_l"` // 实测与标准偏差
	OffsetA      float64   `json:"offset_a"`
	OffsetB      float64   `json:"offset_b"`
	CreatedAt    time.Time `json:"created_at"`
}

// EvidenceKind 工艺证据类别。
type EvidenceKind string

const (
	EvidenceBath        EvidenceKind = "bath"         // 浴液条件
	EvidenceSampling    EvidenceKind = "sampling"     // 取样位置
	EvidenceInstrument  EvidenceKind = "instrument"   // 仪器校准
	EvidenceOther       EvidenceKind = "other"        // 其他
)

// ProcessEvidence 工艺证据。
type ProcessEvidence struct {
	ID          int64          `json:"id"`
	BatchID     int64          `json:"batch_id"`
	Kind        EvidenceKind   `json:"kind"`
	Description string         `json:"description"`
	Status      EvidenceStatus `json:"status"`
	AttachedAt  time.Time      `json:"attached_at"`
	CreatedAt   time.Time      `json:"created_at"`
}

// Verdict 色差来源判定。
type Verdict string

const (
	VerdictBath       Verdict = "bath"        // 浴液条件
	VerdictSampling   Verdict = "sampling"    // 取样位置
	VerdictInstrument Verdict = "instrument"  // 仪器校准
	VerdictMixed      Verdict = "mixed"       // 混合来源
	VerdictInconclusive Verdict = "inconclusive" // 无法判定
)

// ReviewConclusion 复核结论。
type ReviewConclusion struct {
	ID        int64            `json:"id"`
	BatchID   int64            `json:"batch_id"`
	Verdict   Verdict          `json:"verdict"`
	Summary   string           `json:"summary"`
	Status    ConclusionStatus `json:"status"`
	Version   int              `json:"version"`
	PublishedAt *time.Time     `json:"published_at,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// ColorDiffResult 单个测色点的色差计算结果。
type ColorDiffResult struct {
	SampleNo     int     `json:"sample_no"`
	Position     string  `json:"position"`
	DeltaE76     float64 `json:"delta_e76"`
	DeltaE2000   float64 `json:"delta_e2000"`
	WithinTolerance bool `json:"within_tolerance"`
}

// ColorDiffSummary 批次色差汇总。
type ColorDiffSummary struct {
	BatchID       int64             `json:"batch_id"`
	Target        [3]float64        `json:"target"` // 目标色 Lab
	Method        string            `json:"method"` // 使用的方法
	Tolerance     float64           `json:"tolerance"`
	Points        []ColorDiffResult `json:"points"`
	MaxDeltaE     float64           `json:"max_delta_e"`
	MeanDeltaE    float64           `json:"mean_delta_e"`
	AnomalyCount  int               `json:"anomaly_count"`
	ValidCount    int               `json:"valid_count"`
}
