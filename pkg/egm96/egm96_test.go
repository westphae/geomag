package egm96

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

func TestEGM96GridLookup(t *testing.T) {
	lats := []float64{38, -12.25, -84.75, 26, 0}
	lngs := []float64{270, 82.75, 180.5, 279.5, 0}
	hts  := []float64{-30.262, -67.347, -40.254, -26.621, 17.162}

	for i:=0; i<len(lats); i++ {
		p, _ := NewLocationGeodetic(lats[i],lngs[i],0).NearestEGM96GridPoint()

		testDiff("latitude", p.latitude/Deg, lats[i], eps, t)
		testDiff("longitude", p.longitude/Deg, lngs[i], eps, t)
		testDiff("height", p.height, hts[i], eps, t)
	}
}

func TestEGM96GridInterpolationAgainstKnown(t *testing.T) {
	lats := []float64{38, -12.25, 0, 38.628155, -14.621217, 46.874319, -23.617446, 38.625473, -0.466744}
	lngs := []float64{270, 82.75, 0, 269.779155, 305.021114, 102.448729, 133.874712, 359.999500, 0.002300}
	hts  := []float64{-30.262, -67.347, 17.162, -31.628, -2.969, -43.575, 15.871, 50.066, 17.329}

	for i:=0; i<len(lats); i++ {
		h, _ := NewLocationGeodetic(lats[i],lngs[i],0).HeightAboveMSL()
		// 0.1 seems to be the error introduced by bi-linear interpolation rather than splines
		testDiff("height", -h, hts[i], 0.1, t)
	}
}

func TestNewLocationMSL(t *testing.T) {
	lats := []float64{38, -12.25, 0, 38.628155, -14.621217, 46.874319, -23.617446, 38.625473, -0.466744}
	lngs := []float64{270, 82.75, 0, 269.779155, 305.021114, 102.448729, 133.874712, 359.999500, 0.002300}
	hts  := []float64{200, -1000, 99999, 12000, 3600, -50, 8800, 1200000, -1111}

	for i:=0; i<len(lats); i++ {
		l, _ := NewLocationMSL(lats[i],lngs[i],hts[i])
		h, _ := l.HeightAboveMSL()
		// 0.1 seems to be the error introduced by bi-linear interpolation rather than splines
		testDiff("height", h, hts[i], eps, t)
	}
}

func ExampleNearestEGM96GridPoint() {
	p, _ := NewLocationGeodetic(-12.25,82.75,0).NearestEGM96GridPoint()
	fmt.Printf("Lat: %4.2f, Lng: %4.2f, height: %5.3f", p.latitude/Deg, p.longitude/Deg, p.height)
	// Output: Lat: -12.25, Lng: 82.75, height: -67.347
}

func ExampleConvertMSLToHeightAboveWGS84() {
	h, _ := NewLocationGeodetic(-12.25,82.75,1000).HeightAboveMSL()
	fmt.Printf("height Above Ellipsoid: %7.3f", h)
	// Output: height Above Ellipsoid: 1067.347
}
