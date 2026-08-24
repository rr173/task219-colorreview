package httpapi

import (
	"net/http"
	"time"

	"task219-colorreview/internal/model"
)

// measurePointRequest 测色点上传请求体。
type measurePointRequest struct {
	SampleNo   int     `json:"sample_no"`
	Position   string  `json:"position"`
	ColorSpace string  `json:"color_space"`
	L          float64 `json:"l"`
	A          float64 `json:"a"`
	B          float64 `json:"b"`
	MeasuredAt string  `json:"measured_at"`
	Instrument string  `json:"instrument_id,omitempty"`
}

func (s *Server) handleAddMeasurePoint(w http.ResponseWriter, r *http.Request) {
	id, err := requirePathID(w, r, "id")
	if err != nil {
		return
	}
	var req measurePointRequest
	if err := decodeBody(w, r, &req); err != nil {
		return
	}
	measuredAt, err := time.Parse(time.RFC3339, req.MeasuredAt)
	if err != nil {
		writeErr(w, model.ErrInvalidArgument)
		return
	}
	m := &model.MeasurePoint{
		BatchID:    id,
		SampleNo:   req.SampleNo,
		Position:   req.Position,
		ColorSpace: req.ColorSpace,
		L:          req.L,
		A:          req.A,
		B:          req.B,
		MeasuredAt: measuredAt,
	}
	created, err := s.svc.AddMeasurePoint(r.Context(), m, req.Instrument)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleListMeasurePoints(w http.ResponseWriter, r *http.Request) {
	id, err := requirePathID(w, r, "id")
	if err != nil {
		return
	}
	points, err := s.svc.ListMeasurePoints(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"points": points})
}

func (s *Server) handleRejectMeasurePoint(w http.ResponseWriter, r *http.Request) {
	id, err := requirePathID(w, r, "id")
	if err != nil {
		return
	}
	mid, err := requirePathID(w, r, "mid")
	if err != nil {
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	if err := decodeBody(w, r, &body); err != nil {
		return
	}
	updated, err := s.svc.RejectMeasurePoint(r.Context(), id, mid, body.Reason)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
