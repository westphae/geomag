package main

import (
	"github.com/westphae/geomag/pkg/units"
	"testing"
)

func TestMeters(t *testing.T) {
	ms := []units.Meters{1, 1.5, -0.01}
	fts := []float64{3.280840, 4.921260, -0.03280840}

	for i, m := range ms {
		ft := m.ToFeet()
		if ft - fts[i] < -EPS || ft - fts[i] > EPS {
			t.Errorf("Failure converting %6.1f m to ft, expected %6.1f, got %6.1f",
				m, fts[i], ft)
		}
		mm := units.MetersFromFeet(ft)
		if mm - m < -EPS || mm - m > EPS {
			t.Errorf("Failure converting %6.1f m to ft and back, got %6.1f",
				m, mm)
		}
	}

	for i, ft := range fts {
		m := units.MetersFromFeet(ft)
		if m - ms[i] < -EPS || m - ms[i] > EPS {
			t.Errorf("Failure converting %6.1f ft to m, expected %6.1f, got %6.1f",
				ft, ms[i], m)
		}
		fft := m.ToFeet()
		if fft - ft < -EPS || fft - ft > EPS {
			t.Errorf("Failure converting %6.1f ft to m and back, got %6.1f",
				ft, fft)
		}
	}
}

func TestDegrees(t *testing.T) {
	ds := []float64{59, 30, 20, -12, -89}
	ms := []float64{59, 12, 18, 45, 59}
	ss := []float64{59.999, 46, 31, 12, 1.25}
	dds := []units.Degrees{59.999999722, 30.212777777, 20.308611111, -12.753333333, -89.983680555}

	for i, dd := range dds {
		d, m, s := dd.ToDMS()
		if d - ds[i] < -EPS || d - ds[i] > EPS ||
			m - ms[i] < -EPS*60 || m - ms[i] > EPS*60 ||
			s - ss[i] < -EPS*3600 || s - ss[i] > EPS*3600 {
			t.Errorf("Failure converting %6.1f degrees to dms, expected %3.1f %2.1f %2.6f, got %3.1f %2.1f %2.6f",
				dd, ds[i], ms[i], ss[i], d, m, s)
		}
		ddd := units.DegreesFromDMS(d, m, s)
		if ddd - dd < -EPS || ddd - dd > EPS {
			t.Errorf("Failure converting %6.1f degrees to dms and back, got %6.1f",
				dd, ddd)
		}
	}

}
