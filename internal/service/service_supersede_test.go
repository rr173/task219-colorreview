package service

import (
	"context"
	"path/filepath"
	"testing"
	"regexp"

	"task219-colorreview/internal/model"
	"task219-colorreview/internal/store"
)

// TestSupersedeVersionIncrement 验证替代结论从第二版开始编号，并保留旧版本不可变快照。
func TestSupersedeVersionIncrement(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "supersede.db")
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()
	svc := New(s)
	ctx := context.Background()

	// 批次 + 初版结论 v1。
	batch, err := svc.CreateBatch(ctx, &model.DyeBatch{Code: "B-SUP", Name: "sup test", Recipe: "R1"})
	if err != nil {
		t.Fatalf("create batch: %v", err)
	}
	c1, err := svc.CreateConclusion(ctx, &model.ReviewConclusion{
		BatchID: batch.ID, Verdict: model.VerdictBath, Summary: "v1 summary",
	})
	if err != nil {
		t.Fatalf("create conclusion: %v", err)
	}
	if c1.Version != 1 {
		t.Fatalf("first conclusion must be v1, got %d", c1.Version)
	}
	published, err := svc.PublishConclusion(ctx, c1.ID)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if published.Version != 1 {
		t.Fatalf("published must be v1, got %d", published.Version)
	}

	// 建立 v1 的替代版本，应得到 v2。
	c2, err := svc.SupersedeConclusion(ctx, c1.ID, model.VerdictInstrument, "v2 summary")
	if err != nil {
		t.Fatalf("supersede: %v", err)
	}
	if c2.Version != 2 {
		t.Fatalf("supersede must start at v2, got %d", c2.Version)
	}
	if c2.Status != model.ConclusionDraft {
		t.Fatalf("supersede successor should be draft, got %s", c2.Status)
	}

	// 发布 v2 后再替代一次，应得到 v3，确认每次都递增。
	pub2, err := svc.PublishConclusion(ctx, c2.ID)
	if err != nil {
		t.Fatalf("publish v2: %v", err)
	}
	if pub2.Version != 2 {
		t.Fatalf("published v2 must be 2, got %d", pub2.Version)
	}
	c3, err := svc.SupersedeConclusion(ctx, c2.ID, model.VerdictSampling, "v3 summary")
	if err != nil {
		t.Fatalf("supersede v2: %v", err)
	}
	if c3.Version != 3 {
		t.Fatalf("supersede must continue incrementing to v3, got %d", c3.Version)
	}

	// 旧版本快照应被保留：v1、v2 各一条不可变快照。
	versions, err := svc.ListConclusionVersions(ctx, batch.ID)
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 immutable snapshots (v1,v2), got %d", len(versions))
	}
	for _, v := range versions {
		vs, _ := v["version"].(int)
		if vs != 1 && vs != 2 {
			t.Fatalf("unexpected snapshot version %d", vs)
		}
	}
	// 快照内容不可变：v1 summary 仍为初版文本。
	wantV1 := regexp.MustCompile(`v1 summary`)
	if !wantV1.MatchString(versions[0]["summary"].(string)) {
		t.Fatalf("v1 snapshot summary changed: %v", versions[0]["summary"])
	}
}
