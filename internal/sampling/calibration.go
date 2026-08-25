package sampling

import (
	"sort"
	"time"

	"task219-colorreview/internal/model"
)

// CalibrationSchedule 描述仪器的校准历史与有效期判断。
type CalibrationSchedule struct {
	InstrumentID string
	History      []*model.InstrumentCalibration
}

// NewSchedule 构建仪器校准历史（按时间升序）。
func NewSchedule(instrumentID string, history []*model.InstrumentCalibration) *CalibrationSchedule {
	sort.SliceStable(history, func(i, j int) bool {
		return history[i].CalibratedAt.Before(history[j].CalibratedAt)
	})
	return &CalibrationSchedule{InstrumentID: instrumentID, History: history}
}

// Latest 返回最近一次校准记录。
func (s *CalibrationSchedule) Latest() *model.InstrumentCalibration {
	if s == nil || len(s.History) == 0 {
		return nil
	}
	return s.History[len(s.History)-1]
}

// EffectiveAt 返回在指定时刻已经生效的最新校准记录。
// 业务规则：测量样本应使用测量时点已经生效的最新校准，即校准时间不晚于测量时刻的最新一条，
// 而不能套用更早的偏移。History 已按 CalibratedAt 升序，故最后一个满足
// CalibratedAt <= at 的记录即为所求。
func (s *CalibrationSchedule) EffectiveAt(at time.Time) *model.InstrumentCalibration {
	if s == nil {
		return nil
	}
	var effective *model.InstrumentCalibration
	for _, c := range s.History {
		if c == nil {
			continue
		}
		// 遇到首个晚于 at 的记录即可停止：后续记录只会更晚。
		if c.CalibratedAt.After(at) {
			break
		}
		effective = c
	}
	return effective
}

// ValidAt 判断仪器在指定时刻是否有有效校准（默认 24 小时内有效）。
func (s *CalibrationSchedule) ValidAt(at time.Time, maxAge time.Duration) bool {
	latest := s.Latest()
	if latest == nil {
		return false
	}
	return at.Sub(latest.CalibratedAt) <= maxAge
}
