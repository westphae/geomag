package main

import (
	"testing"
	"time"

	"github.com/westphae/geomag/pkg/wmm"
)

func TestGetWMMCoefficients(t *testing.T) {
	wmm.LoadWMMCOF("data/WMM2015v2.COF")
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
			g, h, dg, dh, _ := wmm.GetWMMCoefficients(n, m, tt)
			gErr := g - (gs[i]+float64(j)*dgs[i])
			hErr := h - (hs[i]+float64(j)*dhs[i])
			dgErr := dg - dgs[i]
			dhErr := dh - dhs[i]
			if gErr != 0 {
				t.Errorf("got G(%d,%d)=%6.1f at %v, expecting %6.1f", n, m, g, tt, gs[i])
			}
			if hErr != 0 {
				t.Errorf("got H(%d,%d)=%6.1f at %v, expecting %6.1f", n, m, h, tt, hs[i])
			}
			if dgErr != 0 {
				t.Errorf("got DG(%d,%d)=%6.1f at %v, expecting %6.1f", n, m, dg, tt, dgs[i])
			}
			if dhErr != 0 {
				t.Errorf("got DH(%d,%d)=%6.1f at %v, expecting %6.1f", n, m, dh, tt, dhs[i])
			}
		}
	}
}