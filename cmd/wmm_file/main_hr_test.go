//go:build !wmm_no_hr

package main

import (
	"testing"

	"github.com/westphae/geomag/pkg/wmm/wmmhr"
)

// TestAgainstNOAAHRSample is the WMMHR sibling of TestAgainstNOAASample —
// runs the same NOAA sample_coords.txt input through processLine using the
// embedded WMMHR2025 coefficients and diffs against NOAA's published
// sample_output_hr.txt. Same per-column tolerances as the WMM test (the
// foot-altitude rows still differ by ~0.1 nT due to NOAA's wmm_file.c
// transcription error in the feet-to-km conversion).
//
// Gated by the !wmm_no_hr build tag so the lean test variant
// (`go test -tags wmm_no_hr`) doesn't pull in the embedded WMMHR data.
func TestAgainstNOAAHRSample(t *testing.T) {
	compareWMMFileSample(t, wmmhr.Default(), "testdata/sample_output_hr_noaa.txt")
}
