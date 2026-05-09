package wmm

import (
	"testing"
	"time"

	"github.com/westphae/geomag/pkg/egm96"
)

// TestErrorModelDefault checks that the embedded default Model carries the
// WMM2025 error values. This pairs with TestEmbeddedDefaultLoads (which
// covers the coefficient blob) — together they assert that the default
// model is fully populated.
func TestErrorModelDefault(t *testing.T) {
	m, err := NewModel()
	if err != nil {
		t.Fatalf("NewModel: %v", err)
	}
	em := m.ErrorModel()
	want := ErrorModel{X: 137, Y: 89, Z: 141, H: 133, F: 138, I: 0.20, DA: 0.26, DB: 5417}
	if em != want {
		t.Errorf("default ErrorModel = %+v, want %+v", em, want)
	}
}

// TestErrorModelSwitchesWithCOF is the regression test for issue #1: the
// error model must move with the loaded COF. Loading WMM2020 testdata on a
// binary built for WMM2025 should produce WMM2020 error values, not the
// WMM2025 ones the binary was compiled with.
func TestErrorModelSwitchesWithCOF(t *testing.T) {
	m, err := LoadModel("testdata/WMM2020.COF")
	if err != nil {
		t.Fatalf("LoadModel(WMM2020): %v", err)
	}
	em := m.ErrorModel()
	want := ErrorModel{X: 131, Y: 94, Z: 157, H: 128, F: 148, I: 0.21, DA: 0.26, DB: 5625}
	if em != want {
		t.Errorf("WMM2020 ErrorModel = %+v, want %+v", em, want)
	}

	// And the resulting MagneticField should carry WMM2020 errors.
	mf, _ := m.MagneticField(egm96.NewLocationGeodetic(0, 0, 0),
		time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC))
	if mf.ErrF() != 148 {
		t.Errorf("MagneticField.ErrF() = %v, want 148 (WMM2020)", mf.ErrF())
	}
	if mf.ErrX() != 131 {
		t.Errorf("MagneticField.ErrX() = %v, want 131 (WMM2020)", mf.ErrX())
	}
}

// TestErrorModelUnknownCOFIsZero documents the fallback behavior for COFs
// not in defaultErrorModels — the model loads fine but the error fields
// are zero until SetErrorModel is called.
func TestErrorModelUnknownCOFIsZero(t *testing.T) {
	// WMM2015v1.COF predates the lookup table; we don't know its error
	// constants offhand. Loading should succeed but the error model should
	// be the zero value.
	m, err := LoadModel("testdata/WMM2015v1.COF")
	if err != nil {
		t.Fatalf("LoadModel(WMM2015v1): %v", err)
	}
	if em := m.ErrorModel(); em != (ErrorModel{}) {
		t.Errorf("unknown COF ErrorModel = %+v, want zero value", em)
	}

	// SetErrorModel lets the user supply one.
	custom := ErrorModel{X: 100, Y: 100, Z: 100, H: 100, F: 100, I: 0.5, DA: 0.5, DB: 5000}
	m.SetErrorModel(custom)
	if em := m.ErrorModel(); em != custom {
		t.Errorf("after SetErrorModel ErrorModel = %+v, want %+v", em, custom)
	}

	// And subsequent MagneticField calls reflect the override.
	mf, _ := m.MagneticField(egm96.NewLocationGeodetic(45, 45, 0),
		time.Date(2017, 6, 1, 0, 0, 0, 0, time.UTC))
	if mf.ErrF() != 100 {
		t.Errorf("MagneticField.ErrF() = %v, want 100 (custom override)", mf.ErrF())
	}
}
