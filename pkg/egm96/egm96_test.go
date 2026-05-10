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
		// Bicubic Catmull-Rom interpolation. Measured worst case across these
		// nine UNAVCO reference points was 0.0108 m (vs 0.0557 m for the
		// previous bilinear path); 0.02 m matches the historical 1.8× safety
		// margin the bilinear test used.
		testDiff("height", -h, hts[i], 0.02, t)
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

// TestBoundaryCorners exercises edges of the EGM96 grid, where the
// v1.2025.1 bilinear-bounds off-by-one bug would have either silently
// produced wrong results or read out-of-bounds. The grid is laid out
// top-down (Y0=+90, Y1=-90), so lat=+90 lands on the topmost row (legal)
// while lat=-90 would require a row past the bottom (correctly rejected).
func TestBoundaryCorners(t *testing.T) {
	type tc struct {
		name     string
		lat, lng float64
		wantOK   bool
	}
	cases := []tc{
		{"north-pole-prime-meridian", 90, 0, true},          // legal; bilinear top-row degenerates
		{"south-pole-prime-meridian", -90, 0, false},        // illegal; would read row past bottom
		{"just-inside-south-pole", -89.5, 0, true},
		{"just-inside-north-pole", 89.5, 0, true},
		{"date-line-near-equator", 0, 179.99, true},
		{"wrap-to-zero", 0, 359.99, true},                   // valid after normalization
		{"negative-conv-180-equator", 0, -180, true},         // becomes 180
		{"negative-conv-179.999", 0, -179.999, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			loc := NewLocationGeodetic(c.lat, c.lng, 0)
			h, err := loc.HeightAboveMSL()
			gotOK := err == nil
			if gotOK != c.wantOK {
				t.Errorf("HeightAboveMSL(%v, %v): err=%v wantOK=%v (h=%v)",
					c.lat, c.lng, err, c.wantOK, h)
			}
		})
	}
}

// TestPolarFallback exercises the bilinear fallback path inside
// interpGeoidBicubic. At lat ≈ ±89.9° the bicubic stencil's ny±1..ny+2
// reads would step off the latitude grid (which doesn't wrap), so the
// implementation defers to bilinear within one cell of each pole.
// We assert the call returns a finite, sub-100 m geoid height (EGM96
// magnitudes never exceed ±107 m) and no error.
func TestPolarFallback(t *testing.T) {
	cases := []struct {
		name     string
		lat, lng float64
	}{
		{"near-north-pole", 89.9, 0},
		{"near-north-pole-other-meridian", 89.9, 180},
		{"near-south-pole", -89.9, 0},
		{"near-south-pole-other-meridian", -89.9, 270},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h, err := NewLocationGeodetic(c.lat, c.lng, 0).HeightAboveMSL()
			if err != nil {
				t.Fatalf("HeightAboveMSL(%v, %v): %v", c.lat, c.lng, err)
			}
			if math.IsNaN(h) || math.IsInf(h, 0) {
				t.Fatalf("HeightAboveMSL(%v, %v) = %v, want finite", c.lat, c.lng, h)
			}
			// Sanity bound: -h is the geoid height; the EGM96 grid extremum
			// is roughly +85.4 m / -106.9 m, so 200 m gives plenty of slack
			// without being trivially passable.
			if math.Abs(-h) > 200 {
				t.Fatalf("HeightAboveMSL(%v, %v) = %v, magnitude looks unreasonable", c.lat, c.lng, h)
			}
		})
	}
}

// TestAntimeridianContinuity exercises the longitude-wraparound logic in
// interpGeoidBicubic. Indices 0 and egm96XN-1 represent the same meridian
// (0° = 360°), so evaluating either side of the antimeridian (e.g.
// lng = 179.9° vs lng = -179.9° vs lng = 359.9°) should return values
// matching to within float-arithmetic noise. After NewLocationGeodetic
// normalizes longitude to [0, 360°), all three forms become 179.9°,
// 180.1°, 359.9° respectively — neighboring grid cells with smoothly
// varying geoid heights, but the test specifically verifies that the
// wraparound stencil at lng ≈ 359.9° doesn't introduce a discontinuity
// vs evaluation away from the wrap.
func TestAntimeridianContinuity(t *testing.T) {
	// Three forms that exercise the modular wrap differently:
	//   359.9° lands in cell (egm96XN-2, egm96XN-1), reading wrap(0) for
	//   nx+2 — this is the one that crosses the boundary.
	hWrap, err := NewLocationGeodetic(0, 359.9, 0).HeightAboveMSL()
	if err != nil {
		t.Fatalf("HeightAboveMSL(0, 359.9): %v", err)
	}
	// 0.1° is the symmetric position one tenth of a cell on the other
	// side of the 0/360 seam.
	hSeam, err := NewLocationGeodetic(0, 0.1, 0).HeightAboveMSL()
	if err != nil {
		t.Fatalf("HeightAboveMSL(0, 0.1): %v", err)
	}
	// Negative-form equivalent of 359.9° is -0.1°. Both should normalize
	// to the same stored longitude and produce identical results.
	hNeg, err := NewLocationGeodetic(0, -0.1, 0).HeightAboveMSL()
	if err != nil {
		t.Fatalf("HeightAboveMSL(0, -0.1): %v", err)
	}
	if math.Abs(hWrap-hNeg) > 1e-9 {
		t.Errorf("antimeridian: 359.9° gave %v, -0.1° gave %v — should be identical after normalization", hWrap, hNeg)
	}
	// And 0.1° vs 359.9° should both be smooth across the seam: the
	// geoid is C¹-continuous, so values 0.2° apart in longitude at the
	// equator differ by far less than 1 m in any realistic gradient.
	if math.Abs(hWrap-hSeam) > 1.0 {
		t.Errorf("antimeridian: 359.9° gave %v, 0.1° gave %v — discontinuity larger than 1 m suggests wrap is broken", hWrap, hSeam)
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
