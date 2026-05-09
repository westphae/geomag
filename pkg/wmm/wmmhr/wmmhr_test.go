package wmmhr

import (
	"testing"
	"time"

	"github.com/westphae/geomag/pkg/egm96"
	"github.com/westphae/geomag/pkg/wmm"
)

// TestEmbeddedHRLoads confirms the embedded WMMHR.COF parses cleanly and
// that the resulting *wmm.Model carries the expected metadata. Parallel to
// pkg/wmm's TestEmbeddedDefaultLoads — guards against silent corruption of
// the embedded HR blob the same way.
func TestEmbeddedHRLoads(t *testing.T) {
	m, err := New()
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if got := m.COFName(); got != "WMMHR-2025" {
		t.Errorf("COFName = %q, want WMMHR-2025", got)
	}
	if got := m.Epoch(); got != 2025.0 {
		t.Errorf("Epoch = %v, want 2025.0", got)
	}
	if got := m.MaxN(); got != 133 {
		t.Errorf("MaxN = %d, want 133", got)
	}
	if got := m.ValidDate(); got.IsZero() {
		t.Errorf("ValidDate not populated")
	}
	em := m.ErrorModel()
	want := wmm.ErrorModel{X: 135, Y: 85, Z: 134, H: 130, F: 134, I: 0.19, DA: 0.25, DB: 5205}
	if em != want {
		t.Errorf("ErrorModel = %+v, want %+v", em, want)
	}
}

// TestDefaultIsShared verifies that Default() returns the same *wmm.Model
// across calls so its per-location cache can be hit across the program.
func TestDefaultIsShared(t *testing.T) {
	a := Default()
	b := Default()
	if a != b {
		t.Errorf("Default() returned different instances on successive calls: %p vs %p", a, b)
	}
}

// TestComputesField is a smoke test that catches catastrophic regressions
// in the high-degree spherical-harmonic math: at one geographic point the
// computed field must be finite and roughly in line with WMM's value
// (HR captures additional crustal-scale variations but the global field
// magnitude shouldn't differ by more than a few hundred nT).
func TestComputesField(t *testing.T) {
	m := Default()
	loc := egm96.NewLocationGeodetic(45, -75, 0) // Ottawa-ish
	t0 := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	mf, _ := m.MagneticField(loc, t0)
	if f := mf.F(); f < 30000 || f > 70000 {
		t.Errorf("F = %v at (45,-75); expected roughly 50000 nT", f)
	}
	// Sanity-check that a couple of getters return finite values.
	if d := mf.D(); d != d { // NaN check
		t.Errorf("D is NaN")
	}
	if errF := mf.ErrF(); errF != 134 {
		t.Errorf("ErrF() = %v, want 134 (WMMHR2025 published)", errF)
	}
}
