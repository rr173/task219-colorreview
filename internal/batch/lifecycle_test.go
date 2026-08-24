package batch

import (
	"testing"

	"task219-colorreview/internal/model"
)

func TestAdvanceLifecycle(t *testing.T) {
	path := []model.BatchStatus{
		model.BatchRecipe,
		model.BatchInProduction,
		model.BatchPendingReview,
		model.BatchConfirmed,
		model.BatchSealed,
	}
	for i := 0; i < len(path)-1; i++ {
		next, err := Advance(path[i])
		if err != nil {
			t.Fatalf("advance from %s: %v", path[i], err)
		}
		if next != path[i+1] {
			t.Fatalf("advance from %s = %s, want %s", path[i], next, path[i+1])
		}
	}
}

func TestAdvanceSealedFails(t *testing.T) {
	if _, err := Advance(model.BatchSealed); err == nil {
		t.Fatal("expected error advancing sealed batch")
	}
}

func TestCanTransition(t *testing.T) {
	if !CanTransition(model.BatchRecipe, model.BatchInProduction) {
		t.Fatal("recipe -> in_production should be allowed")
	}
	if CanTransition(model.BatchRecipe, model.BatchConfirmed) {
		t.Fatal("recipe -> confirmed should be rejected")
	}
	if CanTransition(model.BatchSealed, model.BatchConfirmed) {
		t.Fatal("sealed -> confirmed should be rejected")
	}
}

func TestSealable(t *testing.T) {
	if !Sealable(model.BatchPendingReview) {
		t.Fatal("pending_review should be sealable")
	}
	if !Sealable(model.BatchConfirmed) {
		t.Fatal("confirmed should be sealable")
	}
	if Sealable(model.BatchRecipe) {
		t.Fatal("recipe should not be sealable")
	}
}

func TestValidate(t *testing.T) {
	if err := Validate(&model.DyeBatch{Code: "X", Name: "n", Recipe: "r"}); err != nil {
		t.Fatalf("valid batch should pass: %v", err)
	}
	if err := Validate(&model.DyeBatch{Name: "n", Recipe: "r"}); err == nil {
		t.Fatal("missing code should fail")
	}
	if err := Validate(&model.DyeBatch{Code: "X", Name: "n", Recipe: "r", Status: "bogus"}); err == nil {
		t.Fatal("bogus status should fail")
	}
}
