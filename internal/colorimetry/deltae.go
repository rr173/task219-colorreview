// Package colorimetry 提供纺织测色的色差计算：CIE Lab 转换、ΔE 度量与容差判定。
package colorimetry

import "math"

// Lab 表示 CIE L*a*b* 颜色。
type Lab struct {
	L float64
	A float64
	B float64
}

// Chroma 返回彩度 C = sqrt(a^2 + b^2)。
func (c Lab) Chroma() float64 { return math.Hypot(c.A, c.B) }

// HueDeg 返回色相角（度，范围 [0, 360)）。
func (c Lab) HueDeg() float64 {
	h := math.Atan2(c.B, c.A) * 180 / math.Pi
	if h < 0 {
		h += 360
	}
	return h
}

// DeltaE76 CIE76 色差：Lab 空间欧氏距离。
func DeltaE76(a, b Lab) float64 {
	dL := a.L - b.L
	dA := a.A - b.A
	dB := a.B - b.B
	return math.Sqrt(dL*dL + dA*dA + dB*dB)
}

// DeltaE94 CIE94 色差（纺织业常用，权重因子 kL=kC=kH=1）。
func DeltaE94(a, b Lab) float64 {
	dL := a.L - b.L
	c1, c2 := a.Chroma(), b.Chroma()
	dC := c1 - c2
	dA := a.A - b.A
	dB := a.B - b.B
	dH2 := dA*dA + dB*dB - dC*dC
	if dH2 < 0 {
		dH2 = 0
	}
	dH := math.Sqrt(dH2)
	sl := 1.0
	sc := 1 + 0.045*c1
	sh := 1 + 0.015*c1
	return math.Sqrt((dL/sl)*(dL/sl) + (dC/sc)*(dC/sc) + (dH/sh)*(dH/sh))
}

// DeltaE2000 CIEDE2000 色差，是目前纺织与印刷业最准确的色差度量。
func DeltaE2000(a, b Lab) float64 {
	L1, a1, b1 := a.L, a.A, a.B
	L2, a2, b2 := b.L, b.A, b.B

	c1 := math.Hypot(a1, b1)
	c2 := math.Hypot(a2, b2)
	cBar := (c1 + c2) / 2

	g := 0.5 * (1 - math.Sqrt(math.Pow(cBar, 7)/(math.Pow(cBar, 7)+math.Pow(25, 7))))
	a1p := (1 + g) * a1
	a2p := (1 + g) * a2

	c1p := math.Hypot(a1p, b1)
	c2p := math.Hypot(a2p, b2)

	h1p := hueDeg(a1p, b1)
	h2p := hueDeg(a2p, b2)

	dLp := L2 - L1
	dCp := c2p - c1p

	var dhp float64
	switch {
	case c1p*c2p == 0:
		dhp = 0
	case math.Abs(h2p-h1p) <= 180:
		dhp = h2p - h1p
	case h2p <= h1p:
		dhp = h2p - h1p + 360
	default:
		dhp = h2p - h1p - 360
	}

	dHp := 2 * math.Sqrt(c1p*c2p) * math.Sin(dhp*math.Pi/360)

	LBar := (L1 + L2) / 2
	cBarP := (c1p + c2p) / 2

	var hBarP float64
	switch {
	case c1p*c2p == 0:
		hBarP = h1p + h2p
	case math.Abs(h1p-h2p) <= 180:
		hBarP = (h1p + h2p) / 2
	case h1p+h2p < 360:
		hBarP = (h1p + h2p + 360) / 2
	default:
		hBarP = (h1p + h2p - 360) / 2
	}

	t := 1 -
		0.17*math.Cos((hBarP-30)*math.Pi/180) +
		0.24*math.Cos(2*hBarP*math.Pi/180) +
		0.32*math.Cos((3*hBarP+6)*math.Pi/180) -
		0.20*math.Cos((4*hBarP-63)*math.Pi/180)

	dTheta := 30 * math.Exp(-math.Pow((hBarP-275)/25, 2))
	rc := 2 * math.Sqrt(math.Pow(cBarP, 7)/(math.Pow(cBarP, 7)+math.Pow(25, 7)))
	rt := -math.Sin(2*dTheta*math.Pi/180) * rc

	sl := 1 + 0.015*math.Pow(LBar-50, 2)/math.Sqrt(20+math.Pow(LBar-50, 2))
	sc := 1 + 0.045*cBarP
	sh := 1 + 0.015*cBarP*t

	term1 := dLp / sl
	term2 := dCp / sc
	term3 := dHp / sh

	return math.Sqrt(term1*term1 + term2*term2 + term3*term3 + rt*term2*term3)
}

func hueDeg(a, b float64) float64 {
	h := math.Atan2(b, a) * 180 / math.Pi
	if h < 0 {
		h += 360
	}
	return h
}
