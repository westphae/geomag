package wmm

import (
	"bufio"
	"bytes"
	"github.com/westphae/geomag/pkg/egm96"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
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

func TestMagneticFieldFromPaperDetail(t *testing.T) {
	// Test values in paper are only for original version of WMM-2015
	LoadWMMCOF("testdata/WMM2015v1.COF")
	tt := DecimalYear(2017.5)
	loc := egm96.NewLocationGeodetic(-80,240,100e3)

	lat, lng, hh := loc.Geodetic()
	testDiff("lambda", lng, 4.1887902048, epsM, t)
	testDiff("phi", lat, -1.3962634016, epsM, t)
	testDiff("h", hh, 100000.0000000000, epsM, t)
	testDiff("t", float64(tt), 2017.5000000000, epsM, t)

	lat, lng, hh = loc.Spherical()
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

	testDiff("F", MagneticField(mag).F(), 52611.1423211683, epsM, t)
	testDiff("H", MagneticField(mag).H(), 15862.0423159539, epsM, t)
	testDiff("D", MagneticField(mag).D(), 1.2043399870/egm96.Deg, epsM, t)
	testDiff("I", MagneticField(mag).I(), -1.2645351837/egm96.Deg, epsM, t)
	testDiff("DF", MagneticField(mag).DF(), -77.2340297896, epsM, t)
	testDiff("DH", MagneticField(mag).DH(), 16.5720479716, epsM, t)
	testDiff("DD", MagneticField(mag).DD(), -0.0015009297/egm96.Deg, epsM, t)
	testDiff("DI", MagneticField(mag).DI(), 0.0007945653/egm96.Deg, epsM, t)
}

func TestAllTestValuesFromPaper(t *testing.T) {
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

	LoadWMMCOF("testdata/WMM2015v2.COF")

	data, err = ioutil.ReadFile("testdata/WMM2015v2TestValues.txt")
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