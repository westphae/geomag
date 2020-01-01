package wmm

import (
	"fmt"
	"testing"
	"time"
)

const eps = 1e-6

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
