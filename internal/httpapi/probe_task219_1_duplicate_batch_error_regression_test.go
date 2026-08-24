package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"task219-colorreview/internal/httpapi"
	"task219-colorreview/internal/service"
	"task219-colorreview/internal/store"
)

func TestDuplicateBatchUsesConflictResponse(t *testing.T) {
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	h := httpapi.New(service.New(st)).Handler()
	body := []byte(`{"code":"DUP-001","name":"重复批次","recipe":"R1","color_space":"lab"}`)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/batches", bytes.NewReader(body)).WithContext(context.Background())
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)
		if i == 0 && res.Code != http.StatusCreated {
			t.Fatalf("first create status=%d body=%s", res.Code, res.Body.String())
		}
		if i == 1 {
			if res.Code != http.StatusConflict {
				t.Fatalf("duplicate create status=%d body=%s", res.Code, res.Body.String())
			}
			var payload map[string]string
			if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			if payload["error"] == "" {
				t.Fatalf("duplicate response missing domain error: %s", res.Body.String())
			}
		}
	}
}
