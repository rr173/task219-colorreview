package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"task219-colorreview/internal/model"
)

func TestBatchLifecyclePersistsAndRejectsDuplicateCode(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "colorreview.db")

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	b, err := s.CreateBatch(ctx, &model.DyeBatch{
		Code:   "DYE-001",
		Name:   "indigo trial",
		Recipe: "recipe-7",
	})
	if err != nil {
		s.Close()
		t.Fatalf("create batch: %v", err)
	}
	if _, err := s.CreateBatch(ctx, &model.DyeBatch{Code: "DYE-001", Name: "duplicate", Recipe: "recipe-8"}); !errors.Is(err, model.ErrAlreadyExists) {
		s.Close()
		t.Fatalf("duplicate create error = %v, want ErrAlreadyExists", err)
	}
	if err := s.SetBatchStatus(ctx, b.ID, model.BatchRecipe, model.BatchInProduction); err != nil {
		s.Close()
		t.Fatalf("advance batch: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}

	s, err = Open(dbPath)
	if err != nil {
		t.Fatalf("reopen store: %v", err)
	}
	defer s.Close()
	restored, err := s.GetBatch(ctx, b.ID)
	if err != nil {
		t.Fatalf("get restored batch: %v", err)
	}
	if restored.Status != model.BatchInProduction {
		t.Fatalf("restored status = %s, want %s", restored.Status, model.BatchInProduction)
	}
	list, err := s.ListBatches(ctx)
	if err != nil {
		t.Fatalf("list batches: %v", err)
	}
	if len(list) != 1 || list[0].Code != "DYE-001" {
		t.Fatalf("restored batch list = %#v, want one DYE-001 batch", list)
	}
	if err := s.SetBatchStatus(ctx, b.ID, model.BatchRecipe, model.BatchPendingReview); !errors.Is(err, model.ErrInvalidTransition) {
		t.Fatalf("stale transition error = %v, want ErrInvalidTransition", err)
	}
}

func TestWithTxRollsBackPartialWrite(t *testing.T) {
	ctx := context.Background()
	s, err := Open(filepath.Join(t.TempDir(), "rollback.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	rollbackErr := errors.New("force rollback")
	err = WithTx(ctx, s.DB(), func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO batches (code, name, recipe, color_space, status, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			"ROLLBACK-001", "temporary", "recipe", "", string(model.BatchRecipe), "now", "now"); err != nil {
			return err
		}
		return rollbackErr
	})
	if !errors.Is(err, rollbackErr) {
		t.Fatalf("WithTx error = %v, want rollback error", err)
	}
	list, err := s.ListBatches(ctx)
	if err != nil {
		t.Fatalf("list batches after rollback: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("rollback left %d batches, want 0", len(list))
	}
}
