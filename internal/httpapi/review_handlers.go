package httpapi

import (
	"net/http"

	"task219-colorreview/internal/model"
)

// conclusionRequest 结论创建请求体。
type conclusionRequest struct {
	Verdict string `json:"verdict"`
	Summary string `json:"summary"`
}

func (s *Server) handleCreateConclusion(w http.ResponseWriter, r *http.Request) {
	id, err := requirePathID(w, r, "id")
	if err != nil {
		return
	}
	var req conclusionRequest
	if err := decodeBody(w, r, &req); err != nil {
		return
	}
	c := &model.ReviewConclusion{
		BatchID: id,
		Verdict: model.Verdict(req.Verdict),
		Summary: req.Summary,
	}
	created, err := s.svc.CreateConclusion(r.Context(), c)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleGetConclusion(w http.ResponseWriter, r *http.Request) {
	id, err := requirePathID(w, r, "id")
	if err != nil {
		return
	}
	c, err := s.svc.GetConclusion(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) handlePublishConclusion(w http.ResponseWriter, r *http.Request) {
	cid, err := requirePathID(w, r, "cid")
	if err != nil {
		return
	}
	c, err := s.svc.PublishConclusion(r.Context(), cid)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) handleSupersedeConclusion(w http.ResponseWriter, r *http.Request) {
	cid, err := requirePathID(w, r, "cid")
	if err != nil {
		return
	}
	var req conclusionRequest
	if err := decodeBody(w, r, &req); err != nil {
		return
	}
	c, err := s.svc.SupersedeConclusion(r.Context(), cid, model.Verdict(req.Verdict), req.Summary)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (s *Server) handleListConclusionVersions(w http.ResponseWriter, r *http.Request) {
	id, err := requirePathID(w, r, "id")
	if err != nil {
		return
	}
	versions, err := s.svc.ListConclusionVersions(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"versions": versions})
}

func (s *Server) handleInferVerdict(w http.ResponseWriter, r *http.Request) {
	id, err := requirePathID(w, r, "id")
	if err != nil {
		return
	}
	verdict, err := s.svc.InferVerdict(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"verdict": verdict})
}
