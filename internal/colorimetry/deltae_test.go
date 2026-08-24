package colorimetry

import (
	"math"
	"testing"
)

// 参考 CIEDE2000 标准测试用例（Sharma, Wu & Dalal 2005）。
var sharma2000Cases = []struct {
	l1, a1, b1 float64
	l2, a2, b2 float64
	want       float64
}{
	{50.0000, 2.6772, -79.7751, 50.0000, 0.0000, -82.7485, 2.0425},
	{50.0000, 3.1571, -77.2803, 50.0000, 0.0000, -82.7485, 2.8615},
	{50.0000, 2.8361, -74.0200, 50.0000, 0.0000, -82.7485, 3.4412},
	{50.0000, -1.3802, -84.2814, 50.0000, 0.0000, -82.7485, 1.0000},
}

func TestDeltaE2000Sharma(t *testing.T) {
	for i, c := range sharma2000Cases {
		got := DeltaE2000(Lab{L: c.l1, A: c.a1, B: c.b1}, Lab{L: c.l2, A: c.a2, B: c.b2})
		if math.Abs(got-c.want) > 0.01 {
			t.Fatalf("case %d: DeltaE2000 = %.4f, want %.4f", i, got, c.want)
		}
	}
}

func TestDeltaE76Basics(t *testing.T) {
	if d := DeltaE76(Lab{L: 50, A: 0, B: 0}, Lab{L: 50, A: 0, B: 0}); d != 0 {
		t.Fatalf("identical colors should have ΔE76 = 0, got %f", d)
	}
	if d := DeltaE76(Lab{L: 50, A: 10, B: 0}, Lab{L: 50, A: 0, B: 0}); math.Abs(d-10) > 1e-9 {
		t.Fatalf("ΔE76 for a-only diff should be 10, got %f", d)
	}
}

func TestDeltaE94NonNegative(t *testing.T) {
	a := Lab{L: 40, A: 15, B: -20}
	b := Lab{L: 52, A: -5, B: 12}
	if d := DeltaE94(a, b); d < 0 {
		t.Fatalf("ΔE94 should be non-negative, got %f", d)
	}
}

func TestParseLabValidation(t *testing.T) {
	if _, err := ParseLab(101, 0, 0); err == nil {
		t.Fatal("expected error for L=101")
	}
	if _, err := ParseLab(50, 200, 0); err == nil {
		t.Fatal("expected error for a=200")
	}
	if _, err := ParseLab(50, 0, 0); err != nil {
		t.Fatalf("unexpected error for valid Lab: %v", err)
	}
}

func TestRGBToLabWhiteAndBlack(t *testing.T) {
	white := (RGB{R: 255, G: 255, B: 255}).ToLab()
	if math.Abs(white.L-100) > 1 {
		t.Fatalf("white L should be ~100, got %f", white.L)
	}
	black := (RGB{R: 0, G: 0, B: 0}).ToLab()
	if math.Abs(black.L-0) > 1 {
		t.Fatalf("black L should be ~0, got %f", black.L)
	}
}
