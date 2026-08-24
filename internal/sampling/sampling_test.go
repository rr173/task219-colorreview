package sampling

import (
	"errors"
	"testing"
	"time"

	"task219-colorreview/internal/model"
)

func TestCalibrationAndTimelineRules(t *testing.T) {
	cal := &model.InstrumentCalibration{
		InstrumentID: "spectro-7",
		CalibratedAt: time.Unix(100, 0).UTC(),
		RefL:         50,
		RefA:         0,
		RefB:         0,
		OffsetL:      5,
		OffsetA:      2,
		OffsetB:      -1,
	}
	if err := ValidateCalibration(cal); err != nil {
		t.Fatalf("valid calibration rejected: %v", err)
	}
	if l, a, b := ApplyCalibration(55, 2, -1, cal); l != 50 || a != 0 || b != 0 {
		t.Fatalf("ApplyCalibration = (%v, %v, %v), want (50, 0, 0)", l, a, b)
	}
	if l, a, b := ComputeOffsets(50, 0, 0, 55, 2, -1); l != 5 || a != 2 || b != -1 {
		t.Fatalf("ComputeOffsets = (%v, %v, %v), want (5, 2, -1)", l, a, b)
	}

	validPoint := &model.MeasurePoint{
		BatchID:    9,
		SampleNo:   1,
		ColorSpace: "lab",
		L:          50,
		MeasuredAt: time.Unix(101, 0).UTC(),
	}
	if err := ValidatePoint(validPoint); err != nil {
		t.Fatalf("valid measure point rejected: %v", err)
	}
	missingSpace := *validPoint
	missingSpace.ColorSpace = ""
	if !errors.Is(ValidatePoint(&missingSpace), model.ErrColorSpaceMissing) {
		t.Fatal("missing color space should return ErrColorSpaceMissing")
	}

	times := []time.Time{
		time.Unix(1, 0),
		time.Unix(2, 0),
		time.Unix(3, 0),
	}
	if err := EnsureMonotonicTime(times); err != nil {
		t.Fatalf("increasing timeline rejected: %v", err)
	}
	times[2] = times[1]
	if !errors.Is(EnsureMonotonicTime(times), model.ErrTimeInverted) {
		t.Fatal("repeated timestamp should return ErrTimeInverted")
	}
}

func TestCalibrationScheduleUsesLatestTimestamp(t *testing.T) {
	old := &model.InstrumentCalibration{CalibratedAt: time.Unix(100, 0)}
	latest := &model.InstrumentCalibration{CalibratedAt: time.Unix(200, 0)}
	schedule := NewSchedule("spectro-7", []*model.InstrumentCalibration{latest, old})
	if got := schedule.Latest(); got != latest {
		t.Fatalf("Latest returned %p, want newest calibration %p", got, latest)
	}
	if !schedule.ValidAt(time.Unix(250, 0), time.Minute) {
		t.Fatal("calibration should be valid within its max age")
	}
	if schedule.ValidAt(time.Unix(400, 0), time.Minute) {
		t.Fatal("calibration should expire after its max age")
	}
}
