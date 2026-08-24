package colorimetry

import (
	"fmt"
	"math"
)

// RGB 表示 8-bit sRGB 颜色（0-255）。
type RGB struct {
	R, G, B float64
}

// ToLab 将 sRGB 转换为 CIE Lab（D65 白点）。
func (c RGB) ToLab() Lab {
	// 归一化并做 gamma 校正。
	r := srgbInv(c.R / 255)
	g := srgbInv(c.G / 255)
	b := srgbInv(c.B / 255)

	// sRGB -> XYZ（D65）。
	x := r*0.4124564 + g*0.3575761 + b*0.1804375
	y := r*0.2126729 + g*0.7151522 + b*0.0721750
	z := r*0.0193339 + g*0.1191920 + b*0.9503041

	return xyzToLab(x, y, z)
}

// XYZToLab 将 XYZ 转为 CIE Lab（D65）。
func XYZToLab(x, y, z float64) Lab { return xyzToLab(x, y, z) }

func srgbInv(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func xyzToLab(x, y, z float64) Lab {
	// D65 白点。
	const xn, yn, zn = 0.95047, 1.0, 1.08883

	fx := labF(x / xn)
	fy := labF(y / yn)
	fz := labF(z / zn)

	return Lab{
		L: 116*fy - 16,
		A: 500 * (fx - fy),
		B: 200 * (fy - fz),
	}
}

func labF(t float64) float64 {
	const delta = 6.0 / 29.0
	if t > math.Pow(delta, 3) {
		return math.Cbrt(t)
	}
	return t/(3*delta*delta) + 4.0/29.0
}

// ParseLab 从三个浮点值构造 Lab，并做取值范围校验。
func ParseLab(l, a, b float64) (Lab, error) {
	if l < 0 || l > 100 {
		return Lab{}, fmt.Errorf("L must be within [0,100], got %v", l)
	}
	if a < -128 || a > 128 || b < -128 || b > 128 {
		return Lab{}, fmt.Errorf("a/b must be within [-128,128]")
	}
	return Lab{L: l, A: a, B: b}, nil
}

// ParseRGB 从三个 0-255 通道值构造 RGB。
func ParseRGB(r, g, b float64) (RGB, error) {
	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
		return RGB{}, fmt.Errorf("RGB channels must be within [0,255]")
	}
	return RGB{R: r, G: g, B: b}, nil
}
