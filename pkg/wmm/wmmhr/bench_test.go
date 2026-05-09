package wmmhr_test

import (
	"testing"

	"github.com/westphae/geomag/pkg/egm96"
	"github.com/westphae/geomag/pkg/wmm"
	"github.com/westphae/geomag/pkg/wmm/wmmhr"
)

var (
	benchLoc = egm96.NewLocationGeodetic(89, -121, 28*1000)
	benchT   = wmm.DecimalYear(2025.0).ToTime()
)

// BenchmarkMagneticField_WMMHR measures the full single-point evaluation
// cost at WMMHR's degree 133. The per-location cache is invalidated each
// iteration by toggling the location, so each call re-runs the full
// spherical-harmonic sum without re-parsing the COF.
func BenchmarkMagneticField_WMMHR(b *testing.B) {
	m := wmmhr.Default()
	locA := benchLoc
	locB := egm96.NewLocationGeodetic(45, 75, 1000)
	for i := 0; i < b.N; i++ {
		var loc egm96.Location
		if i&1 == 0 {
			loc = locA
		} else {
			loc = locB
		}
		_, _ = m.MagneticField(loc, benchT)
	}
}

// BenchmarkMagneticField_WMMHR_Cached measures repeat evaluation at the
// same location. Should be ~free vs the uncached version above; the gap
// quantifies what the per-location cache buys us.
func BenchmarkMagneticField_WMMHR_Cached(b *testing.B) {
	m := wmmhr.Default()
	// Prime the cache.
	_, _ = m.MagneticField(benchLoc, benchT)
	for i := 0; i < b.N; i++ {
		_, _ = m.MagneticField(benchLoc, benchT)
	}
}

// BenchmarkGrid_WMMHR mirrors BenchmarkGrid_WMM in the parent package.
func BenchmarkGrid_WMMHR(b *testing.B) {
	m, err := wmmhr.New()
	if err != nil {
		b.Fatal(err)
	}
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
