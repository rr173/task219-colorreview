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

// ValidAt 判断仪器在指定时刻是否有有效校准（默认 24 小时内有效）。
func (s *CalibrationSchedule) ValidAt(at time.Time, maxAge time.Duration) bool {
	latest := s.Latest()
	if latest == nil {
		return false
	}
	return at.Sub(latest.CalibratedAt) <= maxAge
}
