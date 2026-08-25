package store_test

import (
	"context"
	"sync"
	"testing"

	"task219-colorreview/internal/model"
	"task219-colorreview/internal/store"
)

func TestConcurrentBatchCASHasSingleWinner(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(":memory:")
	if err != nil { t.Fatal(err) }
	defer st.Close()
	b, err := st.CreateBatch(ctx, &model.DyeBatch{Code: "CAS-001", Name: "CAS批次", Recipe: "R1", ColorSpace: "lab"})
	if err != nil { t.Fatal(err) }
	const workers = 20
	start := make(chan struct{})
	var wg sync.WaitGroup
	var mu sync.Mutex
	success := 0
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if err := st.SetBatchStatus(ctx, b.ID, model.BatchRecipe, model.BatchInProduction); err == nil {
				mu.Lock(); success++; mu.Unlock()
			}
		}()
	}
	close(start)
	wg.Wait()
	final, err := st.GetBatch(ctx, b.ID)
	if err != nil { t.Fatal(err) }
	if success != 1 || final.Status != model.BatchInProduction {
		t.Fatalf("success=%d final=%s, want one CAS winner and in_production", success, final.Status)
	}
}
