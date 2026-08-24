package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"task219-colorreview/internal/model"
)

// decodeBody 解析 JSON 请求体到 v。
func decodeBody(w http.ResponseWriter, r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeErr(w, model.Wrap("decode body", model.ErrInvalidArgument))
		return err
	}
	return nil
}

// pathID 从路径参数取值并解析为 int64。
func pathID(r *http.Request, name string) (int64, error) {
	raw := r.PathValue(name)
	if raw == "" {
		return 0, model.ErrInvalidArgument
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, model.ErrInvalidArgument
	}
	return id, nil
}

// requirePathID 解析路径参数，失败时写错误并返回哨兵错误。
func requirePathID(w http.ResponseWriter, r *http.Request, name string) (int64, error) {
	id, err := pathID(r, name)
	if err != nil {
		writeErr(w, err)
		return 0, errors.New("invalid path id")
	}
	return id, nil
}

func (s *Server) handleCreateBatch(w http.ResponseWriter, r *http.Request) {
	var b model.DyeBatch
	if err := decodeBody(w, r, &b); err != nil {
		return
	}
	created, err := s.svc.CreateBatch(r.Context(), &b)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleListBatches(w http.ResponseWriter, r *http.Request) {
	batches, err := s.svc.ListBatches(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"batches": batches})
}

func (s *Server) handleGetBatch(w http.ResponseWriter, r *http.Request) {
	id, err := requirePathID(w, r, "id")
	if err != nil {
		return
	}
	b, err := s.svc.GetBatch(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (s *Server) handleAdvanceBatch(w http.ResponseWriter, r *http.Request) {
	id, err := requirePathID(w, r, "id")
	if err != nil {
		return
	}
	b, err := s.svc.AdvanceBatch(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (s *Server) handleSealBatch(w http.ResponseWriter, r *http.Request) {
	id, err := requirePathID(w, r, "id")
	if err != nil {
		return
	}
	b, err := s.svc.SealBatch(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (s *Server) handleDeclareColorSpace(w http.ResponseWriter, r *http.Request) {
	id, err := requirePathID(w, r, "id")
	if err != nil {
		return
	}
	var body struct {
		ColorSpace string `json:"color_space"`
	}
	if err := decodeBody(w, r, &body); err != nil {
		return
	}
	b, err := s.svc.DeclareColorSpace(r.Context(), id, body.ColorSpace)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}
