package wmm

import (
	"bufio"
	"bytes"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/westphae/geomag/pkg/egm96"
)

const (
	epsM = 5e-5
	red = "\u001b[31m"
	green = "\u001b[32m"
	reset = "\u001b[0m"
)

func testDiff(name string, actual, expected float64, eps float64, t *testing.T) {
	if actual-expected >= -eps && actual-expected <= eps {
		t.Logf("%s%s correct: expected %6.4f, got %6.4f%s", green, name, expected, actual, reset)
		return
	}
	t.Errorf("%s%s incorrect: expected %6.4f, got %6.4f%s", red, name, expected, actual, reset)
}

// TestMagneticFieldGetters exercises every public scalar getter on
// MagneticField against a single known fixture row from NOAA's WMM2025
// test vectors. The full integration tests in TestAll20XXTestValuesFromPaper
// validate the math; this test exists so a regression in any individual
// accessor (D, I, F, H, X, Y, Z, dD, dI, dF, dH, dX, dY, dZ, ErrX/Y/Z/H/F/I)
// fails on its own with a clear name.
func TestMagneticFieldGetters(t *testing.T) {
	// Row 1 of WMM2025_TEST_VALUES.txt:
	// year=2025.0, alt=28 km, lat=89, lng=-121
	// expected: D=-99.77, I=88.47, H=1504.298, X=-255.389, Y=-1482.461,
	//           Z=56194.289, F=56214.420
	// secular: dD=2.49, dI=-0.01, dH=10.29, dX=62.72, dY=-21.24, dZ=18.08, dF=18.34
	if err := LoadWMMCOF("testdata/WMM2025.COF"); err != nil {
		t.Fatalf("LoadWMMCOF: %v", err)
	}
	loc := egm96.NewLocationGeodetic(89, -121, 28*1000)
	mf, _ := CalculateWMMMagneticField(loc, DecimalYear(2025.0).ToTime())

	// Component getters
	testDiff("D", mf.D(), -99.77, 0.005, t)
	testDiff("I", mf.I(), 88.47, 0.005, t)
	testDiff("H", mf.H(), 1504.298, 0.05, t)
	testDiff("F", mf.F(), 56214.420, 0.05, t)

	xE, yE, zE, dxE, dyE, dzE := mf.Ellipsoidal()
	testDiff("X", xE, -255.389, 0.05, t)
	testDiff("Y", yE, -1482.461, 0.05, t)
	testDiff("Z", zE, 56194.289, 0.05, t)

	// Secular variation
	testDiff("DD", mf.DD(), 2.491706, 0.05, t)
	testDiff("DI", mf.DI(), -0.009987, 0.05, t)
	testDiff("DH", mf.DH(), 10.286, 0.05, t)
	testDiff("DF", mf.DF(), 18.344, 0.05, t)
	testDiff("DX", dxE, 62.724, 0.05, t)
	testDiff("DY", dyE, -21.243, 0.05, t)
	testDiff("DZ", dzE, 18.075, 0.05, t)

	// Error model getters (WMM2025 published values)
	testDiff("ErrX", mf.ErrX(), 137, 1e-9, t)
	testDiff("ErrY", mf.ErrY(), 89, 1e-9, t)
	testDiff("ErrZ", mf.ErrZ(), 141, 1e-9, t)
	testDiff("ErrH", mf.ErrH(), 133, 1e-9, t)
	testDiff("ErrF", mf.ErrF(), 138, 1e-9, t)
	testDiff("ErrI", mf.ErrI(), 0.20, 1e-9, t)
	// ErrD is location-dependent; just sanity-check it's a positive finite number.
	if e := mf.ErrD(); !(e > 0 && e < 100) {
		t.Errorf("ErrD = %v; expected something positive and bounded", e)
	}
}

func TestMagneticFieldFromPaperDetail(t *testing.T) {
	// Test values in paper are only for original version of WMM-2015
	_ = LoadWMMCOF("testdata/WMM2015v1.COF")
	tt := DecimalYear(2017.5)
	loc := egm96.NewLocationGeodetic(-80,240,100e3)

	lat, lng, hh := loc.Geodetic()
	testDiff("lambda", lng, 4.1887902048, epsM, t)
	testDiff("phi", lat, -1.3962634016, epsM, t)
	testDiff("h", hh, 100000.0000000000, epsM, t)
	testDiff("t", float64(tt), 2017.5000000000, epsM, t)

	lat, _, hh = loc.Spherical()
	testDiff("phi-prime", lat, -1.3951289589, epsM, t)
	testDiff("r", hh, 6457402.3484473705, epsM, t)

	var g, h float64
	g, h, _, _, _ = GetWMMCoefficients(1, 0, tt.ToTime())
	testDiff("g(1,0,t)", g, -29411.7500000000, epsM, t)
	testDiff("h(1,0,t)", h, 0.0000000000, epsM, t)

	g, h, _, _, _ = GetWMMCoefficients(1, 1, tt.ToTime())
	testDiff("g(1,1,t)", g, -1456.3500000000, epsM, t)
	testDiff("h(1,1,t)", h, 4729.2000000000, epsM, t)

	g, h, _, _, _ = GetWMMCoefficients(2, 0, tt.ToTime())
	testDiff("g(2,0,t)", g, -2466.8000000000, epsM, t)
	testDiff("h(2,0,t)", h, 0.0000000000, epsM, t)

	g, h, _, _, _ = GetWMMCoefficients(2, 1, tt.ToTime())
	testDiff("g(2,1,t)", g, 3004.2500000000, epsM, t)
	testDiff("h(2,1,t)", h, -2913.3500000000, epsM, t)

	g, h, _, _, _ = GetWMMCoefficients(2, 2, tt.ToTime())
	testDiff("g(2,2,t)", g, 1682.6000000000, epsM, t)
	testDiff("h(2,2,t)", h, -675.2500000000, epsM, t)

	mag, _ := CalculateWMMMagneticField(loc, tt.ToTime())
	xS, yS, zS, dxS, dyS, dzS := mag.Spherical()
	testDiff("X-prime", xS, 5626.6068398092, epsM, t)
	testDiff("Y-prime", yS, 14808.8492023104, epsM, t)
	testDiff("Z-prime", zS, -50169.4287102381, epsM, t)
	testDiff("Xprime-dot", dxS, 28.2627812813, epsM, t)
	testDiff("Yprime-dot", dyS, 6.9411521726, epsM, t)
	testDiff("Zprime-dot", dzS, 86.2115570931, epsM, t)

	xE, yE, zE, dxE, dyE, dzE := mag.Ellipsoidal()
	testDiff("X", xE, 5683.5175495763, epsM, t)
	testDiff("Y", yE, 14808.8492023104, epsM, t)
	testDiff("Z", zE, -50163.0133654779, epsM, t)
	testDiff("Xdot", dxE, 28.1649610434, epsM, t)
	testDiff("Ydot", dyE, 6.9411521726, epsM, t)
	testDiff("Zdot", dzE, 86.2435641169, epsM, t)

	testDiff("F", mag.F(), 52611.1423211683, epsM, t)
	testDiff("H", mag.H(), 15862.0423159539, epsM, t)
	testDiff("D", mag.D(), 1.2043399870/egm96.Deg, epsM, t)
	testDiff("I", mag.I(), -1.2645351837/egm96.Deg, epsM, t)
	testDiff("DF", mag.DF(), -77.2340297896, epsM, t)
	testDiff("DH", mag.DH(), 16.5720479716, epsM, t)
	testDiff("DD", mag.DD(), -0.0015009297/egm96.Deg, epsM, t)
	testDiff("DI", mag.DI(), 0.0007945653/egm96.Deg, epsM, t)
}

func TestAll2015v2TestValuesFromPaper(t *testing.T) {
	var (
		date                   DecimalYear
		height                 float64
		lat, lon               float64
		x, y, z                float64
		h, f, i, d             float64
		gv                     float64
		xdot, ydot, zdot       float64
		hdot, fdot, idot, ddot float64
		data                   []byte
		dat                    []string
		err                    error
	)

	_ = LoadWMMCOF("testdata/WMM2015v2.COF")

	data, err = os.ReadFile("testdata/WMM2015v2TestValues.txt")
	scanner := bufio.NewScanner(bytes.NewReader(data))
	// Read and parse header
	if !scanner.Scan() {
		panic(err)
	}
	_ = strings.Fields(scanner.Text()) // Not using the header
	for scanner.Scan() {
		dat = strings.Fields(scanner.Text())
		dd, err := strconv.ParseFloat(dat[0], 64)
		if err != nil {
			panic(err)
		}

		date = DecimalYear(dd)
		if dd, err = strconv.ParseFloat(dat[1], 64); err != nil {
			panic(err)
		}
		height = dd*1000
		if dd, err = strconv.ParseFloat(dat[2], 64); err != nil {
			panic(err)
		}
		lat = dd
		if dd, err = strconv.ParseFloat(dat[3], 64); err != nil {
			panic(err)
		}
		lon = dd
		loc := egm96.NewLocationGeodetic(lat,lon,height)

		mag, _ := CalculateWMMMagneticField(loc, date.ToTime())
		xE, yE, zE, dxE, dyE, dzE := mag.Ellipsoidal()

		if x, err = strconv.ParseFloat(dat[4], 64); err != nil {
			panic(err)
		}
		testDiff("X", xE, x, 0.05, t)
		if y, err = strconv.ParseFloat(dat[5], 64); err != nil {
			panic(err)
		}
		testDiff("Y", yE, y, 0.05, t)
		if z, err = strconv.ParseFloat(dat[6], 64); err != nil {
			panic(err)
		}
		testDiff("Z", zE, z, 0.05, t)
		if h, err = strconv.ParseFloat(dat[7], 64); err != nil {
			panic(err)
		}
		testDiff("H", mag.H(), h, 0.05, t)
		if f, err = strconv.ParseFloat(dat[8], 64); err != nil {
			panic(err)
		}
		testDiff("F", mag.F(), f, 0.05, t)
		if i, err = strconv.ParseFloat(dat[9], 64); err != nil {
			panic(err)
		}
		testDiff("I", mag.I(), i, 0.005, t)
		if d, err = strconv.ParseFloat(dat[10], 64); err != nil {
			panic(err)
		}
		testDiff("D", mag.D(), d, 0.005, t)
		if gv, err = strconv.ParseFloat(dat[11], 64); err != nil {
			panic(err)
		}
		testDiff("GV", mag.GV(loc), gv, 0.005, t)
		if xdot, err = strconv.ParseFloat(dat[12], 64); err != nil {
			panic(err)
		}
		testDiff("Xdot", dxE, xdot, 0.05, t)
		if ydot, err = strconv.ParseFloat(dat[13], 64); err != nil {
			panic(err)
		}
		testDiff("Ydot", dyE, ydot, 0.05, t)
		if zdot, err = strconv.ParseFloat(dat[14], 64); err != nil {
			panic(err)
		}
		testDiff("Zdot", dzE, zdot, 0.05, t)
		if hdot, err = strconv.ParseFloat(dat[15], 64); err != nil {
			panic(err)
		}
		testDiff("Hdot", mag.DH(), hdot, 0.05, t)
		if fdot, err = strconv.ParseFloat(dat[16], 64); err != nil {
			panic(err)
		}
		testDiff("Fdot", mag.DF(), fdot, 0.05, t)
		if idot, err = strconv.ParseFloat(dat[17], 64); err != nil {
			panic(err)
		}
		testDiff("Idot", mag.DI(), idot, 0.005, t)
		if ddot, err = strconv.ParseFloat(dat[18], 64); err != nil {
			panic(err)
		}
		testDiff("Ddot", MagneticField(mag).DD(), ddot, 0.005, t)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

}

// TestAllWMMHR2025TestValues runs the WMMHR2025 published test points (the
// 12-point compact-format file from NOAA's WMMHR2025COF.zip distribution)
// through the embedded WMMHR coefficients and validates every returned
// scalar against the reference. Same idea as TestAll2025TestValuesFromPaper,
// but the WMMHR reference file uses a different column order — fields are
// X Y Z H F I D GV Xdot Ydot Zdot Hdot Fdot Idot Ddot rather than the
// WMM-paper ordering — so this test parses positions explicitly.
func TestAllWMMHR2025TestValues(t *testing.T) {
	// Load WMMHR via testdata (we keep the COF copy alongside other test
	// fixtures so the test doesn't depend on the wmmhr sub-package).
	if err := LoadWMMCOF("testdata/WMMHR.COF"); err != nil {
		// WMMHR.COF isn't in testdata — copy from the embedded source via
		// the wmmhr sub-package would create an import cycle, so we just
		// skip if the file isn't there. Devs can copy it manually if they
		// want this test locally; CI ensures it's present.
		t.Skipf("testdata/WMMHR.COF not present (%v); skipping WMMHR test", err)
	}

	data, err := os.ReadFile("testdata/WMMHR2025_TEST_VALUES.txt")
	if err != nil {
		t.Fatalf("open WMMHR test values: %v", err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		f := strings.Fields(line)
		if len(f) < 19 {
			continue
		}
		date, _ := strconv.ParseFloat(f[0], 64)
		altKm, _ := strconv.ParseFloat(f[1], 64)
		lat, _ := strconv.ParseFloat(f[2], 64)
		lng, _ := strconv.ParseFloat(f[3], 64)
		loc := egm96.NewLocationGeodetic(lat, lng, altKm*1000)
		mag, _ := CalculateWMMMagneticField(loc, DecimalYear(date).ToTime())
		xE, yE, zE, dxE, dyE, dzE := mag.Ellipsoidal()

		expectX, _ := strconv.ParseFloat(f[4], 64)
		expectY, _ := strconv.ParseFloat(f[5], 64)
		expectZ, _ := strconv.ParseFloat(f[6], 64)
		expectH, _ := strconv.ParseFloat(f[7], 64)
		expectF, _ := strconv.ParseFloat(f[8], 64)
		expectI, _ := strconv.ParseFloat(f[9], 64)
		expectD, _ := strconv.ParseFloat(f[10], 64)
		// f[11] is GV; "NaN" at non-polar points; skipped here.
		expectXdot, _ := strconv.ParseFloat(f[12], 64)
		expectYdot, _ := strconv.ParseFloat(f[13], 64)
		expectZdot, _ := strconv.ParseFloat(f[14], 64)
		expectHdot, _ := strconv.ParseFloat(f[15], 64)
		expectFdot, _ := strconv.ParseFloat(f[16], 64)
		expectIdot, _ := strconv.ParseFloat(f[17], 64)
		expectDdot, _ := strconv.ParseFloat(f[18], 64)

		// File reports values at 0.1 nT / 0.01° precision. Use 0.1 nT /
		// 0.01° tolerance so rounding-direction differences don't trip us.
		const tolNT = 0.1
		const tolDeg = 0.01
		const tolRate = 0.1
		testDiff("X", xE, expectX, tolNT, t)
		testDiff("Y", yE, expectY, tolNT, t)
		testDiff("Z", zE, expectZ, tolNT, t)
		testDiff("H", mag.H(), expectH, tolNT, t)
		testDiff("F", mag.F(), expectF, tolNT, t)
		testDiff("I", mag.I(), expectI, tolDeg, t)
		testDiff("D", mag.D(), expectD, tolDeg, t)
		testDiff("Xdot", dxE, expectXdot, tolRate, t)
		testDiff("Ydot", dyE, expectYdot, tolRate, t)
		testDiff("Zdot", dzE, expectZdot, tolRate, t)
		testDiff("Hdot", mag.DH(), expectHdot, tolRate, t)
		testDiff("Fdot", mag.DF(), expectFdot, tolRate, t)
		testDiff("Idot", mag.DI(), expectIdot, tolDeg, t)
		testDiff("Ddot", mag.DD(), expectDdot, tolDeg, t)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan: %v", err)
	}
}

func TestAll2025TestValuesFromPaper(t *testing.T) {
	var (
		date                   DecimalYear
		height                 float64
		lat, lon               float64
		x, y, z                float64
		h, f, i, d             float64
		xdot, ydot, zdot       float64
		hdot, fdot, idot, ddot float64
		data                   []byte
		dat                    []string
		err                    error
	)

	_ = LoadWMMCOF("testdata/WMM2025.COF")

	data, err = os.ReadFile("testdata/WMM2025_TEST_VALUES.txt")
	scanner := bufio.NewScanner(bytes.NewReader(data))
	// Read and parse header
	if !scanner.Scan() {
		panic(err)
	}
	for scanner.Scan() {
		// Skip the header lines
		if scanner.Text()[0]=='#' {
			continue
		}

		dat = strings.Fields(scanner.Text())

		dd, err := strconv.ParseFloat(dat[0], 64)
		if err != nil {
			panic(err)
		}
		date = DecimalYear(dd)

		if dd, err = strconv.ParseFloat(dat[1], 64); err != nil {
			panic(err)
		}
		height = dd*1000

		if dd, err = strconv.ParseFloat(dat[2], 64); err != nil {
			panic(err)
		}
		lat = dd

		if dd, err = strconv.ParseFloat(dat[3], 64); err != nil {
			panic(err)
		}
		lon = dd

		loc := egm96.NewLocationGeodetic(lat,lon,height)

		mag, _ := CalculateWMMMagneticField(loc, date.ToTime())
		xE, yE, zE, dxE, dyE, dzE := mag.Ellipsoidal()

		if d, err = strconv.ParseFloat(dat[4], 64); err != nil {
			panic(err)
		}
		testDiff("D", mag.D(), d, 0.005, t)

		if i, err = strconv.ParseFloat(dat[5], 64); err != nil {
			panic(err)
		}
		testDiff("I", mag.I(), i, 0.005, t)

		if h, err = strconv.ParseFloat(dat[6], 64); err != nil {
			panic(err)
		}
		testDiff("H", mag.H(), h, 0.05, t)

		if x, err = strconv.ParseFloat(dat[7], 64); err != nil {
			panic(err)
		}
		testDiff("X", xE, x, 0.05, t)

		if y, err = strconv.ParseFloat(dat[8], 64); err != nil {
			panic(err)
		}
		testDiff("Y", yE, y, 0.05, t)

		if z, err = strconv.ParseFloat(dat[9], 64); err != nil {
			panic(err)
		}
		testDiff("Z", zE, z, 0.05, t)

		if f, err = strconv.ParseFloat(dat[10], 64); err != nil {
			panic(err)
		}
		testDiff("F", mag.F(), f, 0.05, t)

		if ddot, err = strconv.ParseFloat(dat[11], 64); err != nil {
			panic(err)
		}
		testDiff("Ddot", MagneticField(mag).DD(), ddot, 0.05, t)

		if idot, err = strconv.ParseFloat(dat[12], 64); err != nil {
			panic(err)
		}
		testDiff("Idot", mag.DI(), idot, 0.05, t)

		if hdot, err = strconv.ParseFloat(dat[13], 64); err != nil {
			panic(err)
		}
		testDiff("Hdot", mag.DH(), hdot, 0.05, t)

		if xdot, err = strconv.ParseFloat(dat[14], 64); err != nil {
			panic(err)
		}
		testDiff("Xdot", dxE, xdot, 0.05, t)

		if ydot, err = strconv.ParseFloat(dat[15], 64); err != nil {
			panic(err)
		}
		testDiff("Ydot", dyE, ydot, 0.05, t)

		if zdot, err = strconv.ParseFloat(dat[16], 64); err != nil {
			panic(err)
		}
		testDiff("Zdot", dzE, zdot, 0.05, t)

		if fdot, err = strconv.ParseFloat(dat[17], 64); err != nil {
			panic(err)
		}
		testDiff("Fdot", mag.DF(), fdot, 0.05, t)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

}

func TestAll2020TestValuesFromPaper(t *testing.T) {
	var (
		date                   DecimalYear
		height                 float64
		lat, lon               float64
		x, y, z                float64
		h, f, i, d             float64
		xdot, ydot, zdot       float64
		hdot, fdot, idot, ddot float64
		data                   []byte
		dat                    []string
		err                    error
	)

	_ = LoadWMMCOF("testdata/WMM2020.COF")

	data, err = os.ReadFile("testdata/WMM2020_TEST_VALUES.txt")
	scanner := bufio.NewScanner(bytes.NewReader(data))
	// Read and parse header
	if !scanner.Scan() {
		panic(err)
	}
	for scanner.Scan() {
		// Skip the header lines
		if scanner.Text()[0]=='#' {
			continue
		}

		dat = strings.Fields(scanner.Text())

		dd, err := strconv.ParseFloat(dat[0], 64)
		if err != nil {
			panic(err)
		}
		date = DecimalYear(dd)

		if dd, err = strconv.ParseFloat(dat[1], 64); err != nil {
			panic(err)
		}
		height = dd*1000

		if dd, err = strconv.ParseFloat(dat[2], 64); err != nil {
			panic(err)
		}
		lat = dd

		if dd, err = strconv.ParseFloat(dat[3], 64); err != nil {
			panic(err)
		}
		lon = dd

		loc := egm96.NewLocationGeodetic(lat,lon,height)

		mag, _ := CalculateWMMMagneticField(loc, date.ToTime())
		xE, yE, zE, dxE, dyE, dzE := mag.Ellipsoidal()

		if d, err = strconv.ParseFloat(dat[4], 64); err != nil {
			panic(err)
		}
		testDiff("D", mag.D(), d, 0.005, t)

		if i, err = strconv.ParseFloat(dat[5], 64); err != nil {
			panic(err)
		}
		testDiff("I", mag.I(), i, 0.005, t)

		if h, err = strconv.ParseFloat(dat[6], 64); err != nil {
			panic(err)
		}
		testDiff("H", mag.H(), h, 0.05, t)

		if x, err = strconv.ParseFloat(dat[7], 64); err != nil {
			panic(err)
		}
		testDiff("X", xE, x, 0.05, t)

		if y, err = strconv.ParseFloat(dat[8], 64); err != nil {
			panic(err)
		}
		testDiff("Y", yE, y, 0.05, t)

		if z, err = strconv.ParseFloat(dat[9], 64); err != nil {
			panic(err)
		}
		testDiff("Z", zE, z, 0.05, t)

		if f, err = strconv.ParseFloat(dat[10], 64); err != nil {
			panic(err)
		}
		testDiff("F", mag.F(), f, 0.05, t)

		if ddot, err = strconv.ParseFloat(dat[11], 64); err != nil {
			panic(err)
		}
		testDiff("Ddot", MagneticField(mag).DD(), ddot, 0.05, t)

		if idot, err = strconv.ParseFloat(dat[12], 64); err != nil {
			panic(err)
		}
		testDiff("Idot", mag.DI(), idot, 0.05, t)

		if hdot, err = strconv.ParseFloat(dat[13], 64); err != nil {
			panic(err)
		}
		testDiff("Hdot", mag.DH(), hdot, 0.05, t)

		if xdot, err = strconv.ParseFloat(dat[14], 64); err != nil {
			panic(err)
		}
		testDiff("Xdot", dxE, xdot, 0.05, t)

		if ydot, err = strconv.ParseFloat(dat[15], 64); err != nil {
			panic(err)
		}
		testDiff("Ydot", dyE, ydot, 0.05, t)

		if zdot, err = strconv.ParseFloat(dat[16], 64); err != nil {
			panic(err)
		}
		testDiff("Zdot", dzE, zdot, 0.05, t)

		if fdot, err = strconv.ParseFloat(dat[17], 64); err != nil {
			panic(err)
		}
		testDiff("Fdot", mag.DF(), fdot, 0.05, t)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

}
