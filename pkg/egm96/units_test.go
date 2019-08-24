package egm96

import (
	"fmt"
	"testing"
)

func TestDegrees(t *testing.T) {
	ds := []float64{59, 30, 20, -12, -89, 0, 0, 0, 0}
	ms := []float64{59, 12, 18, 45, 59, 6, -12, 0, 0}
	ss := []float64{59.999, 46, 31, 12, 1.25, 0, 54, -45, 18}
	dds := []float64{59.999999722, 30.212777777, 20.308611111, -12.753333333, -89.983680555, 0.1, -0.215, -0.0125, 0.005}

	for i, dd := range dds {
		d, m, s := DegreesToDMS(dd)
		testDiff(fmt.Sprintf("%6.1f° degrees portion", dd), d, ds[i], eps, t)
		testDiff(fmt.Sprintf("%6.1f° minutes portion", dd), m, ms[i], eps*60, t)
		testDiff(fmt.Sprintf("%6.1f° seconds portion", dd), s, ss[i], eps*60*60, t)
		ddd := DMSToDegrees(d, m, s)
		testDiff(fmt.Sprintf("%6.1f° to dms and back", dd), ddd, dd, eps, t)
	}

}
