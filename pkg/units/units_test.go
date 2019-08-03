package units

import (
	"fmt"
	"testing"
)

const eps = 1e-6

func testDiff(name string, actual, expected float64, eps float64, t *testing.T) {
	if actual - expected > -eps && actual - expected < eps {
		t.Logf("%s correct: expected %8.4f, got %8.4f", name, expected, actual)
		return
	}
	t.Errorf("%s incorrect: expected %8.4f, got %8.4f", name, expected, actual)
}

func TestMeters(t *testing.T) {
	ms := []Meters{1, 1.5, -0.01}
	fts := []float64{3.280840, 4.921260, -0.03280840}

	for i, m := range ms {
		ft := m.ToFeet()
		testDiff(fmt.Sprintf("%6.2fm", m), ft, fts[i], eps, t)
		mm := MetersFromFeet(ft)
		testDiff(fmt.Sprintf("%6.2fm to feet and back", m), float64(mm), float64(m), eps, t)
	}

	for i, ft := range fts {
		m := MetersFromFeet(ft)
		testDiff(fmt.Sprintf("%6.1fft", ft), float64(m), float64(ms[i]), eps, t)
		fft := m.ToFeet()
		testDiff(fmt.Sprintf("%6.1fft to m and back", ft), fft, ft, eps, t)
	}
}

func TestDegrees(t *testing.T) {
	ds := []float64{59, 30, 20, -12, -89}
	ms := []float64{59, 12, 18, 45, 59}
	ss := []float64{59.999, 46, 31, 12, 1.25}
	dds := []Degrees{59.999999722, 30.212777777, 20.308611111, -12.753333333, -89.983680555}

	for i, dd := range dds {
		d, m, s := dd.ToDMS()
		testDiff(fmt.Sprintf("%6.1f° degrees portion", dd), d, ds[i], eps, t)
		testDiff(fmt.Sprintf("%6.1f° minutes portion", dd), m, ms[i], eps*60, t)
		testDiff(fmt.Sprintf("%6.1f° seconds portion", dd), s, ss[i], eps*60*60, t)
		ddd := DegreesFromDMS(d, m, s)
		testDiff(fmt.Sprintf("%6.1f° to dms and back", dd), float64(ddd), float64(dd), eps, t)
	}

}
