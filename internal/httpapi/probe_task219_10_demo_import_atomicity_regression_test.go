package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"task219-colorreview/internal/httpapi"
	"task219-colorreview/internal/service"
	"task219-colorreview/internal/store"
)

func TestDemoImportDoesNotLeavePartialBatchOnCalibrationFailure(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer st.Close()
	if _, err := st.DB().ExecContext(ctx, `CREATE TRIGGER fail_demo_calibration BEFORE INSERT ON instrument_calibrations BEGIN SELECT RAISE(ABORT, 'calibration blocked'); END`); err != nil { t.Fatal(err) }
	h := httpapi.New(service.New(st)).Handler()
	res := httptest.NewRecorder()
	h.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/api/demo/import", nil))
	if res.Code < 400 { t.Fatalf("demo import status=%d body=%s", res.Code, res.Body.String()) }
	batches, err := service.New(st).ListBatches(ctx)
	if err != nil { t.Fatal(err) }
	if len(batches) != 0 { t.Fatalf("partial batches=%d, want rollback", len(batches)) }
}
