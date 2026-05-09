package wmm

import (
	"testing"
	"time"

	"github.com/westphae/geomag/pkg/egm96"
)

// fixedTestPoint is the WMM2025 test row 1 (lat=89, lng=-121, alt=28km, t=2025.0).
// Picked because it's far from any equator/pole singularity but at high latitude
// where HR-vs-WMM differences are meaningful.
var (
	benchLoc = egm96.NewLocationGeodetic(89, -121, 28*1000)
	benchT   = DecimalYear(2025.0).ToTime()
)

// BenchmarkMagneticField_WMM measures the full single-point evaluation cost
// at the standard-WMM (degree 12) hot path. Resets the per-location cache
// each iteration so we measure the actual spherical-harmonic computation,
// not a cache hit.
func BenchmarkMagneticField_WMM(b *testing.B) {
	m := Default()
	// Drain the cache so each iteration recomputes.
	for i := 0; i < b.N; i++ {
		m.cacheMu.Lock()
		m.haveCache = false
		m.cacheMu.Unlock()
		_, _ = m.MagneticField(benchLoc, benchT)
	}
}

// BenchmarkGetters measures the cost of extracting every public scalar
// from a single MagneticField — the access pattern wmm_file uses for
// every output row.
func BenchmarkGetters(b *testing.B) {
	m := Default()
	mf, _ := m.MagneticField(benchLoc, benchT)
	for i := 0; i < b.N; i++ {
		_ = mf.D()
		_ = mf.I()
		_ = mf.F()
		_ = mf.H()
		_, _, _, _, _, _ = mf.Ellipsoidal()
		_ = mf.DD()
		_ = mf.DI()
		_ = mf.DH()
		_ = mf.DF()
		_ = mf.GV(benchLoc)
		_ = mf.ErrD()
	}
}

// BenchmarkGrid_WMM measures end-to-end batch evaluation cost: 5×5 lat/lng
// at one altitude and time, with the per-location cache cycling on every
// (lat, lng) pair. Mirrors the wmm_grid use pattern.
func BenchmarkGrid_WMM(b *testing.B) {
	benchGrid(b, Default())
}

func benchGrid(b *testing.B, m *Model) {
	b.Helper()
	t := benchT
	for i := 0; i < b.N; i++ {
		for lat := -60.0; lat <= 60.0; lat += 30 {
			for lng := -120.0; lng <= 120.0; lng += 60 {
				loc := egm96.NewLocationGeodetic(lat, lng, 0)
				_, _ = m.MagneticField(loc, t)
			}
		}
	}
}

// BenchmarkParseLoad measures cold-start cost: parse + first MagneticField
// evaluation.
func BenchmarkParseLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m, err := NewModel()
		if err != nil {
			b.Fatal(err)
		}
		_, _ = m.MagneticField(benchLoc, benchT)
	}
}

// epoch sentinel kept to validate benchT independently.
var _ = time.Time{}
