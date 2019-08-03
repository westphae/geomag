package egm96

import (
	"testing"

	"github.com/westphae/geomag/pkg/units"
)

const eps = 1e-6

func testDiff(name string, actual, expected float64, eps float64, t *testing.T) {
	if actual - expected > -eps && actual - expected < eps {
		t.Logf("%s correct: expected %8.4f, got %8.4f", name, expected, actual)
		return
	}
	t.Errorf("%s incorrect: expected %8.4f, got %8.4f", name, expected, actual)
}

func TestEGM96GridLookup(t *testing.T) {
	lats := []units.Degrees{38, -12.25, -84.75, 26, 0}
	lngs := []units.Degrees{270, 82.75, 180.5, 279.5, 0}
	hts  := []units.Meters{-30.262, -67.347, -40.254, -26.621, 17.162}

	for i:=0; i<len(lats); i++ {
		_, p := GetEGM96GridPoint(units.Location{
			Latitude: lats[i],
			Longitude: lngs[i],
			Height: 0,
		})

		testDiff("latitude", float64(p.Latitude), float64(lats[i]), eps, t)
		testDiff("longitude", float64(p.Longitude), float64(lngs[i]), eps, t)
		testDiff("height", float64(p.Height), float64(hts[i]), eps, t)
	}
}

func TestEGM96GridInterpolationAgainstKnown(t *testing.T) {
	lats := []units.Degrees{38, -12.25, 0, 38.628155, -14.621217, 46.874319, -23.617446, 38.625473, -0.466744}
	lngs := []units.Degrees{270, 82.75, 0, 269.779155, 305.021114, 102.448729, 133.874712, 359.999500, 0.002300}
	hts  := []units.Meters{-30.262, -67.347, 17.162, -31.628, -2.969, -43.575, 15.871, 50.066, 17.329}

	for i:=0; i<len(lats); i++ {
		_, h := ConvertMSLToHeightAboveWGS84(units.Location{
			Latitude: lats[i],
			Longitude: lngs[i],
			Height: 0,
		})
		// 0.1 seems to be the error introduced by bi-linear interpolation rather than splines
		testDiff("height", float64(h), float64(hts[i]), 0.1, t)
	}
}
