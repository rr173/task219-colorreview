package httpapi

import (
	"net/http"
	"time"

	"task219-colorreview/internal/model"
)

// evidenceRequest 证据提交请求体。
type evidenceRequest struct {
	Kind        string `json:"kind"`
	Description string `json:"description"`
	AttachedAt  string `json:"attached_at,omitempty"`
}

func (s *Server) handleAddEvidence(w http.ResponseWriter, r *http.Request) {
	id, err := requirePathID(w, r, "id")
	if err != nil {
		return
	}
	var req evidenceRequest
	if err := decodeBody(w, r, &req); err != nil {
		return
	}
	e := &model.ProcessEvidence{
		BatchID:     id,
		Kind:        model.EvidenceKind(req.Kind),
		Description: req.Description,
	}
	if req.AttachedAt != "" {
		t, err := time.Parse(time.RFC3339, req.AttachedAt)
		if err != nil {
			writeErr(w, model.ErrInvalidArgument)
			return
		}
		e.AttachedAt = t
	}
	created, err := s.svc.AddEvidence(r.Context(), e)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleListEvidences(w http.ResponseWriter, r *http.Request) {
	id, err := requirePathID(w, r, "id")
	if err != nil {
		return
	}
	evs, err := s.svc.ListEvidences(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"evidences": evs})
}

func (s *Server) handleConfirmEvidence(w http.ResponseWriter, r *http.Request) {
	id, err := requirePathID(w, r, "id")
	if err != nil {
		return
	}
	eid, err := requirePathID(w, r, "eid")
	if err != nil {
		return
	}
	updated, err := s.svc.ConfirmEvidence(r.Context(), id, eid)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
