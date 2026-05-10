package egm96

import "testing"

// BenchmarkHeightAboveMSL_Bicubic measures the production bicubic
// Catmull-Rom path. Picked at lat=10.66°N, lng=286.19°E — the worst-case
// region from the wider sweep where bicubic-vs-bilinear differs most
// (Caribbean geoid gradient). Off-grid by 0.66° / 0.19° to ensure we
// hit interior fractional offsets.
func BenchmarkHeightAboveMSL_Bicubic(b *testing.B) {
	egm96LoadOnce.Do(loadEGM96Grid)
	gx := (286.19 - egm96X0) / egm96DX
	gy := (10.66 - egm96Y0) / egm96DY
	for i := 0; i < b.N; i++ {
		_ = interpGeoidBicubic(gx, gy)
	}
}

// BenchmarkHeightAboveMSL_Bilinear keeps the historical bilinear cost
// measurable for comparison; same point as the bicubic bench.
func BenchmarkHeightAboveMSL_Bilinear(b *testing.B) {
	egm96LoadOnce.Do(loadEGM96Grid)
	gx := (286.19 - egm96X0) / egm96DX
	gy := (10.66 - egm96Y0) / egm96DY
	for i := 0; i < b.N; i++ {
		_ = interpGeoidBilinear(gx, gy)
	}
}
