package wmm

import (
	"fmt"
	"testing"
	"time"
)

const eps = 1e-6

// TestEmbeddedDefaultLoads exercises the embedded default coefficients via
// NewModel. It would fail if the embedded WMM.COF is corrupt or unparseable
// (the case that snuck through the earlier go-bindata-based PR #8 because
// every other test loaded explicit testdata files instead of the embed).
func TestEmbeddedDefaultLoads(t *testing.T) {
	m, err := NewModel()
	if err != nil {
		t.Fatalf("NewModel() failed to parse the embedded WMM.COF: %v", err)
	}
	if m.COFName() == "" {
		t.Errorf("embedded model has no COF name")
	}
	if m.Epoch() == 0 {
		t.Errorf("embedded model has zero epoch")
	}
	if m.ValidDate().IsZero() {
		t.Errorf("embedded model has no valid date")
	}
	// Sanity-check at least one coefficient — guards against the parser
	// silently producing empty coefficient slices.
	g, _, _, _, err := m.Coefficients(1, 0, m.ValidDate())
	if err != nil {
		t.Errorf("Coefficients(1, 0): %v", err)
	}
	if g == 0 {
		t.Errorf("g(1,0) = 0; coefficients look unloaded")
	}
}

// TestLoadWMMCOFEmptyReloadsEmbedded covers the legacy package-level path
// that PR #8's corrupt bindata.go silently broke at init time.
func TestLoadWMMCOFEmptyReloadsEmbedded(t *testing.T) {
	if err := LoadWMMCOF(""); err != nil {
		t.Fatalf("LoadWMMCOF(\"\"): %v", err)
	}
	if Epoch == 0 {
		t.Errorf("Epoch package var not populated after LoadWMMCOF(\"\")")
	}
	if COFName == "" {
		t.Errorf("COFName package var not populated after LoadWMMCOF(\"\")")
	}
	g, _, _, _, err := GetWMMCoefficients(1, 0, ValidDate)
	if err != nil {
		t.Errorf("GetWMMCoefficients(1, 0): %v", err)
	}
	if g == 0 {
		t.Errorf("g(1,0) = 0; coefficients did not load from embedded default")
	}
}

func TestGetWMM2015v2Coefficients(t *testing.T) {
	_ = LoadWMMCOF("testdata/WMM2015v2.COF")
	nms := [][]int{{1, 0}, {2, 2}, {5, 1}, {5, 4}, {12, 0}, {12, 6}, {12, 11}}
	gs := []float64{-29438.2, 1679.0, 360.1, -157.2, -2.0, 0.1, -0.9}
	hs := []float64{0.0, -638.8, 46.9, 16.0, 0.0, 0.7, -0.2}
	dgs := []float64{7.0, 0.3, 0.6, 1.2, 0.0, 0.0, 0.0}
	dhs := []float64{0.0, -17.3, 0.2, 3.3, 0.0, 0.0, 0.0}
	ts := []time.Time{
		time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	for j, tt := range ts {
		for i, nm := range nms {
			n := nm[0]
			m := nm[1]
			g, h, dg, dh, _ := GetWMMCoefficients(n, m, tt)
			testDiff(fmt.Sprintf("G(%d,%d)", n, m), g, gs[i]+float64(j)*dgs[i], eps, t)
			testDiff(fmt.Sprintf("H(%d,%d)", n, m), h, hs[i]+float64(j)*dhs[i], eps, t)
			testDiff(fmt.Sprintf("DG(%d,%d)", n, m), dg, dgs[i], eps, t)
			testDiff(fmt.Sprintf("DH(%d,%d)", n, m), dh, dhs[i], eps, t)
		}
	}
}

func TestGetWMM2020Coefficients(t *testing.T) {
	_ = LoadWMMCOF("testdata/WMM2020.COF")
	nms := [][]int{{1, 0}, {2, 2}, {5, 1}, {5, 4}, {12, 0}, {12, 6}, {12, 11}}
	gs := []float64{-29404.5, 1676.8, 363.1, -151.2, -2.0, 0.3, -1.1}
	hs := []float64{0.0, -734.8, 47.7, 32.2, 0.0, 0.7, 0.0}
	dgs := []float64{6.7, -2.2, 0.6, 1.2, 0.0, 0.0, 0.0}
	dhs := []float64{0.0, -23.9, 0.1, 3.0, 0.0, 0.0, 0.0}
	ts := []time.Time{
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	for j, tt := range ts {
		for i, nm := range nms {
			n := nm[0]
			m := nm[1]
			g, h, dg, dh, _ := GetWMMCoefficients(n, m, tt)
			testDiff(fmt.Sprintf("G(%d,%d)", n, m), g, gs[i]+float64(j)*dgs[i], eps, t)
			testDiff(fmt.Sprintf("H(%d,%d)", n, m), h, hs[i]+float64(j)*dhs[i], eps, t)
			testDiff(fmt.Sprintf("DG(%d,%d)", n, m), dg, dgs[i], eps, t)
			testDiff(fmt.Sprintf("DH(%d,%d)", n, m), dh, dhs[i], eps, t)
		}
	}
}

func TestGetWMM2025Coefficients(t *testing.T) {
	_ = LoadWMMCOF("testdata/WMM2025.COF")
	nms := [][]int{{1, 0}, {2, 2}, {5, 1}, {5, 4}, {12, 0}, {12, 6}, {12, 11}}
	gs := []float64{-29351.8, 1649.3, 368.9, -142.0, -2.0, 0.6, -1.3}
	hs := []float64{0.0, -815.1, 45.4, 43.0, 0.0, 0.6, 0.1}
	dgs := []float64{12.0, -8.0, 1.4, 2.2, 0.0, 0.1, 0.0}
	dhs := []float64{0.0, -12.1, -0.5, 1.7, 0.0, 0.0, 0.0}
	ts := []time.Time{
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	for j, tt := range ts {
		for i, nm := range nms {
			n := nm[0]
			m := nm[1]
			g, h, dg, dh, _ := GetWMMCoefficients(n, m, tt)
			testDiff(fmt.Sprintf("G(%d,%d)", n, m), g, gs[i]+float64(j)*dgs[i], eps, t)
			testDiff(fmt.Sprintf("H(%d,%d)", n, m), h, hs[i]+float64(j)*dhs[i], eps, t)
			testDiff(fmt.Sprintf("DG(%d,%d)", n, m), dg, dgs[i], eps, t)
			testDiff(fmt.Sprintf("DH(%d,%d)", n, m), dh, dhs[i], eps, t)
		}
	}
}
