package httpapi

import (
	"net/http"
	"time"

	"task219-colorreview/internal/model"
)

// bathCurveRequest 浴液曲线上传请求体。
type bathCurveRequest struct {
	Channel string                `json:"channel"`
	Points  []bathCurvePointInput `json:"points"`
}

type bathCurvePointInput struct {
	Timestamp string  `json:"timestamp"` // RFC3339
	Value     float64 `json:"value"`
}

func (s *Server) handleSaveBathCurve(w http.ResponseWriter, r *http.Request) {
	id, err := requirePathID(w, r, "id")
	if err != nil {
		return
	}
	var req bathCurveRequest
	if err := decodeBody(w, r, &req); err != nil {
		return
	}
	c := &model.BathCurve{BatchID: id, Channel: req.Channel}
	for _, p := range req.Points {
		t, err := time.Parse(time.RFC3339, p.Timestamp)
		if err != nil {
			writeErr(w, model.ErrInvalidArgument)
			return
		}
		c.Points = append(c.Points, model.BathCurvePoint{Timestamp: t, Value: p.Value})
	}
	saved, err := s.svc.SaveBathCurve(r.Context(), c)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, saved)
}

func (s *Server) handleListBathCurves(w http.ResponseWriter, r *http.Request) {
	id, err := requirePathID(w, r, "id")
	if err != nil {
		return
	}
	curves, err := s.svc.ListBathCurves(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"curves": curves})
}
