package egm96

import (
	"fmt"
	"math"
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

	for i := range lats {
		p, _ := NewLocationGeodetic(lats[i], lngs[i], 0).NearestEGM96GridPoint()

		testDiff("latitude", p.latitude/Deg, lats[i], eps, t)
		testDiff("longitude", p.longitude/Deg, lngs[i], eps, t)
		testDiff("height", p.height, hts[i], eps, t)
	}
}

func TestEGM96GridInterpolationAgainstKnown(t *testing.T) {
	lats := []float64{38, -12.25, 0, 38.628155, -14.621217, 46.874319, -23.617446, 38.625473, -0.466744}
	lngs := []float64{270, 82.75, 0, 269.779155, 305.021114, 102.448729, 133.874712, 359.999500, 0.002300}
	hts  := []float64{-30.262, -67.347, 17.162, -31.628, -2.969, -43.575, 15.871, 50.066, 17.329}

	for i := range lats {
		h, _ := NewLocationGeodetic(lats[i], lngs[i], 0).HeightAboveMSL()
		// 0.1 seems to be the error introduced by bi-linear interpolation rather than splines
		testDiff("height", -h, hts[i], 0.1, t)
	}
}

func TestNewLocationMSL(t *testing.T) {
	lats := []float64{38, -12.25, 0, 38.628155, -14.621217, 46.874319, -23.617446, 38.625473, -0.466744}
	lngs := []float64{270, 82.75, 0, 269.779155, 305.021114, 102.448729, 133.874712, 359.999500, 0.002300}
	hts  := []float64{200, -1000, 99999, 12000, 3600, -50, 8800, 1200000, -1111}

	for i := range lats {
		l, _ := NewLocationMSL(lats[i], lngs[i], hts[i])
		h, _ := l.HeightAboveMSL()
		// 0.1 seems to be the error introduced by bi-linear interpolation rather than splines
		testDiff("height", h, hts[i], eps, t)
	}
}

// TestNegativeLongitudeAccepted is the regression test for issue #3:
// callers using the GPS/WGS84 [-180, 180] longitude convention should not
// hit "requested longitude … lies outside of EGM96 longitude range".
func TestNegativeLongitudeAccepted(t *testing.T) {
	// The exact case from #3: somewhere in northern California.
	lat, lng := 39.865315, -121.32870643118366

	// Both methods that take user-supplied lat/lng should work.
	if _, err := NewLocationGeodetic(lat, lng, 0).HeightAboveMSL(); err != nil {
		t.Errorf("HeightAboveMSL with negative lng: %v", err)
	}
	if _, err := NewLocationGeodetic(lat, lng, 0).NearestEGM96GridPoint(); err != nil {
		t.Errorf("NearestEGM96GridPoint with negative lng: %v", err)
	}

	// Equivalent positive-longitude form should produce the same result.
	hNeg, _ := NewLocationGeodetic(lat, lng, 0).HeightAboveMSL()
	hPos, _ := NewLocationGeodetic(lat, lng+360, 0).HeightAboveMSL()
	if hNeg != hPos {
		t.Errorf("MSL height differs across longitude conventions: %.6f vs %.6f", hNeg, hPos)
	}

	// And over-wound longitudes (e.g. 540 = 180) should also work.
	if _, err := NewLocationGeodetic(0, 540, 0).HeightAboveMSL(); err != nil {
		t.Errorf("HeightAboveMSL with lng=540 (= 180): %v", err)
	}

	// Stored longitude should be in [0, 360°) as a radian-equivalent value.
	loc := NewLocationGeodetic(lat, lng, 0)
	_, lngStored, _ := loc.Geodetic()
	wantLngDeg := lng + 360 // -121.328… → 238.671…
	if got := lngStored / Deg; got < 0 || got >= 360 {
		t.Errorf("stored longitude %v deg not in [0,360)", got)
	} else if math.Abs(got-wantLngDeg) > 1e-9 {
		t.Errorf("stored longitude = %v deg, want %v deg", got, wantLngDeg)
	}
}

func ExampleLocation_NearestEGM96GridPoint() {
	p, _ := NewLocationGeodetic(-12.25, 82.75, 0).NearestEGM96GridPoint()
	fmt.Printf("Lat: %4.2f, Lng: %4.2f, height: %5.3f", p.latitude/Deg, p.longitude/Deg, p.height)
	// Output: Lat: -12.25, Lng: 82.75, height: -67.347
}

func ExampleLocation_HeightAboveMSL() {
	h, _ := NewLocationGeodetic(-12.25, 82.75, 1000).HeightAboveMSL()
	fmt.Printf("height above MSL: %7.3f", h)
	// Output: height above MSL: 1067.347
}
