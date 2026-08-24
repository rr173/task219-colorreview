package httpapi

import (
	"net/http"

	"task219-colorreview/internal/service"
)

// diffRequest 色差计算请求体。
type diffRequest struct {
	TargetL   float64 `json:"target_l"`
	TargetA   float64 `json:"target_a"`
	TargetB   float64 `json:"target_b"`
	Method    string  `json:"method"`
	Tolerance float64 `json:"tolerance"`
}

func (s *Server) handleComputeColorDiff(w http.ResponseWriter, r *http.Request) {
	id, err := requirePathID(w, r, "id")
	if err != nil {
		return
	}
	var req diffRequest
	if err := decodeBody(w, r, &req); err != nil {
		return
	}
	if len(req.Method) == 0 {
		req.Method = "cie2000"
	}
	summary, err := s.svc.ComputeColorDiff(r.Context(), service.DiffRequest{
		BatchID:   id,
		TargetL:   req.TargetL,
		TargetA:   req.TargetA,
		TargetB:   req.TargetB,
		Method:    req.Method,
		Tolerance: req.Tolerance,
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}
