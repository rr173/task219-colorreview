package httpapi

import (
	"net/http"
	"time"

	"task219-colorreview/internal/model"
)

// calibrationRequest 校准记录请求体。
type calibrationRequest struct {
	InstrumentID string  `json:"instrument_id"`
	CalibratedAt string  `json:"calibrated_at"`
	RefL         float64 `json:"ref_l"`
	RefA         float64 `json:"ref_a"`
	RefB         float64 `json:"ref_b"`
	OffsetL      float64 `json:"offset_l"`
	OffsetA      float64 `json:"offset_a"`
	OffsetB      float64 `json:"offset_b"`
}

func (s *Server) handleCreateCalibration(w http.ResponseWriter, r *http.Request) {
	var req calibrationRequest
	if err := decodeBody(w, r, &req); err != nil {
		return
	}
	calibratedAt, err := time.Parse(time.RFC3339, req.CalibratedAt)
	if err != nil {
		writeErr(w, model.ErrInvalidArgument)
		return
	}
	c := &model.InstrumentCalibration{
		InstrumentID: req.InstrumentID,
		CalibratedAt: calibratedAt,
		RefL:         req.RefL,
		RefA:         req.RefA,
		RefB:         req.RefB,
		OffsetL:      req.OffsetL,
		OffsetA:      req.OffsetA,
		OffsetB:      req.OffsetB,
	}
	created, err := s.svc.CreateCalibration(r.Context(), c)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleListCalibrations(w http.ResponseWriter, r *http.Request) {
	cals, err := s.svc.ListCalibrations(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"calibrations": cals})
}
