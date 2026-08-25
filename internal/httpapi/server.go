// Package httpapi 提供 HTTP 层：路由（/api 前缀）与 JSON 请求响应。
package httpapi

import (
	"encoding/json"
	"net/http"

	"task219-colorreview/internal/model"
	"task219-colorreview/internal/service"
	"task219-colorreview/internal/webui"
)

// Server 承载 HTTP 路由与服务依赖。
type Server struct {
	svc *service.Service
	mux *http.ServeMux
}

// New 构造 Server 并注册全部路由。
func New(svc *service.Service) *Server {
	s := &Server{svc: svc, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) routes() {
	mux := s.mux

	// 批次生命周期
	mux.HandleFunc("POST /api/batches", s.handleCreateBatch)
	mux.HandleFunc("GET /api/batches", s.handleListBatches)
	mux.HandleFunc("GET /api/batches/{id}", s.handleGetBatch)
	mux.HandleFunc("POST /api/batches/{id}/advance", s.handleAdvanceBatch)
	mux.HandleFunc("POST /api/batches/{id}/seal", s.handleSealBatch)
	mux.HandleFunc("POST /api/batches/{id}/color-space", s.handleDeclareColorSpace)

	// 浴液曲线
	mux.HandleFunc("POST /api/batches/{id}/bath-curves", s.handleSaveBathCurve)
	mux.HandleFunc("GET /api/batches/{id}/bath-curves", s.handleListBathCurves)

	// 测色点
	mux.HandleFunc("POST /api/batches/{id}/measure-points", s.handleAddMeasurePoint)
	mux.HandleFunc("GET /api/batches/{id}/measure-points", s.handleListMeasurePoints)
	mux.HandleFunc("POST /api/batches/{id}/measure-points/{mid}/reject", s.handleRejectMeasurePoint)

	// 仪器校准
	mux.HandleFunc("POST /api/instruments/calibrations", s.handleCreateCalibration)
	mux.HandleFunc("GET /api/instruments/calibrations", s.handleListCalibrations)

	// 色差计算
	mux.HandleFunc("POST /api/batches/{id}/color-diff", s.handleComputeColorDiff)

	// 工艺证据
	mux.HandleFunc("POST /api/batches/{id}/evidences", s.handleAddEvidence)
	mux.HandleFunc("GET /api/batches/{id}/evidences", s.handleListEvidences)
	mux.HandleFunc("POST /api/batches/{id}/evidences/{eid}/confirm", s.handleConfirmEvidence)

	// 复核结论
	mux.HandleFunc("POST /api/batches/{id}/conclusion", s.handleCreateConclusion)
	mux.HandleFunc("GET /api/batches/{id}/conclusion", s.handleGetConclusion)
	mux.HandleFunc("POST /api/batches/{id}/conclusion/{cid}/publish", s.handlePublishConclusion)
	mux.HandleFunc("POST /api/batches/{id}/conclusion/{cid}/supersede", s.handleSupersedeConclusion)
	mux.HandleFunc("GET /api/batches/{id}/conclusion/versions", s.handleListConclusionVersions)
	mux.HandleFunc("GET /api/batches/{id}/verdict", s.handleInferVerdict)

	// 杂项
	mux.HandleFunc("GET /api/self-check", s.handleSelfCheck)
	mux.HandleFunc("POST /api/demo/import", s.handleDemoImport)

	// 前端页面
	mux.Handle("/", webui.Handler())
}

// Handler 返回根 handler（供 http.Server 使用）。
func (s *Server) Handler() http.Handler { return s.mux }

// writeJSON 统一 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeErr 把错误映射为 HTTP 状态码并输出。
func writeErr(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case model.Is(err, model.ErrNotFound):
		status = http.StatusNotFound
	case model.Is(err, model.ErrDuplicateSample), model.Is(err, model.ErrAlreadyExists):
		status = http.StatusConflict
	case model.Is(err, model.ErrInvalidArgument), model.Is(err, model.ErrColorSpaceMissing),
		model.Is(err, model.ErrTimeInverted), model.Is(err, model.ErrRejectReasonMissing):
		status = http.StatusBadRequest
	case model.Is(err, model.ErrBatchSealed), model.Is(err, model.ErrConclusionFrozen),
		model.Is(err, model.ErrInvalidTransition), model.Is(err, model.ErrConflictUnresolved),
		model.Is(err, model.ErrCalibrationMissing):
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
