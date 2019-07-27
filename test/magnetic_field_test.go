package main

import (
	"bufio"
	"bytes"
	"github.com/westphae/geomag/pkg/units"
	"github.com/westphae/geomag/pkg/wmm"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
)

func testDiff(name string, actual, expected float64, eps float64, t *testing.T) {
	if actual - expected < -eps || actual - expected > eps {
		t.Errorf("%s incorrect: expected %8.4f, got %8.4f", name, expected, actual)
	}
}

func TestMagneticFieldFromPaperDetail(t *testing.T) {
	// Test values in paper are only for original version of WMM-2015
	wmm.LoadWMMCOF("data/WMM2015v1.COF")
	tt := wmm.DecimalYear(2017.5)
	loc := wmm.Geodetic{
		Latitude: units.Degrees(-80),
		Longitude: units.Degrees(240),
		Height: units.Meters(100e3),
	}

	testDiff("lambda", float64(loc.Longitude)*wmm.Deg, 4.1887902048, 1e-10, t)
	testDiff("phi", float64(loc.Latitude)*wmm.Deg, -1.3962634016, 1e-10, t)
	testDiff("h", float64(loc.Height), 100000.0000000000, 1e-10, t)
	testDiff("t", float64(tt), 2017.5000000000, 1e-10, t)

	locS := loc.ToSpherical()
	testDiff("phi-prime", float64(locS.Latitude)*wmm.Deg, -1.3951289589, 1e-10, t)
	testDiff("r", float64(locS.Height), 6457402.3484473705, 1e-10, t)

	var g, h float64
	ghEps := 0.05
	g, h, _, _, _ = wmm.GetWMMCoefficients(1, 0, tt.ToTime())
	testDiff("g(1,0,t)", g, -29411.7500000000, ghEps, t)
	testDiff("h(1,0,t)", h, 0.0000000000, ghEps, t)

	g, h, _, _, _ = wmm.GetWMMCoefficients(1, 1, tt.ToTime())
	testDiff("g(1,1,t)", g, -1456.3500000000, ghEps, t)
	testDiff("h(1,1,t)", h, 4729.2000000000, ghEps, t)

	g, h, _, _, _ = wmm.GetWMMCoefficients(2, 0, tt.ToTime())
	testDiff("g(2,0,t)", g, -2466.8000000000, ghEps, t)
	testDiff("h(2,0,t)", h, 0.0000000000, ghEps, t)

	g, h, _, _, _ = wmm.GetWMMCoefficients(2, 1, tt.ToTime())
	testDiff("g(2,1,t)", g, 3004.2500000000, ghEps, t)
	testDiff("h(2,1,t)", h, -2913.3500000000, ghEps, t)

	g, h, _, _, _ = wmm.GetWMMCoefficients(2, 2, tt.ToTime())
	testDiff("g(2,2,t)", g, 1682.6000000000, ghEps, t)
	testDiff("h(2,2,t)", h, -675.2500000000, ghEps, t)

	magS := wmm.CalculateWMMMagneticField(locS, tt.ToTime())
	testDiff("X-prime", magS.X, 5626.6068398092, 1e-10, t)
	testDiff("Y-prime", magS.Y, 14808.8492023104, 1e-10, t)
	testDiff("Z-prime", magS.Z, -50169.4287102381, 1e-10, t)
	testDiff("Xprime-dot", magS.DX, 28.2627812813, 1e-10, t)
	testDiff("Yprime-dot", magS.DY, 6.9411521726, 1e-10, t)
	testDiff("Zprime-dot", magS.DZ, 86.2115570931, 1e-10, t)

	mag := magS.ToEllipsoidal(loc)
	testDiff("X", mag.X, 5683.5175495763, 1e-10, t)
	testDiff("Y", mag.Y, 14808.8492023104, 1e-10, t)
	testDiff("Z", mag.Z, -50163.0133654779, 1e-10, t)
	testDiff("Xdot", mag.DX, 28.1649610434, 1e-10, t)
	testDiff("Ydot", mag.DY, 6.9411521726, 1e-10, t)
	testDiff("Zdot", mag.DZ, 86.2435641169, 1e-10, t)

	testDiff("F", wmm.MagneticField(mag).F(), 52611.1423211683, 1e-10, t)
	testDiff("H", wmm.MagneticField(mag).H(), 15862.0423159539, 1e-10, t)
	testDiff("D", wmm.MagneticField(mag).D(), 1.2043399870/wmm.Deg, 1e-10, t)
	testDiff("I", wmm.MagneticField(mag).I(), -1.2645351837/wmm.Deg, 1e-10, t)
	testDiff("DF", wmm.MagneticField(mag).DF(), -77.2340297896, 1e-10, t)
	testDiff("DH", wmm.MagneticField(mag).DH(), 16.5720479716, 1e-10, t)
	testDiff("DD", wmm.MagneticField(mag).DD(), -0.0015009297/wmm.Deg, 1e-10, t)
	testDiff("DI", wmm.MagneticField(mag).DI(), 0.0007945653/wmm.Deg, 1e-10, t)
}

func TestAllTestValuesFromPaper(t *testing.T) {
	var (
		date wmm.DecimalYear
		height units.Meters
		lat, lon units.Degrees
		x, y, z float64
		h, f, i, d float64
		//gv float64
		xdot, ydot, zdot float64
		hdot, fdot, idot, ddot float64
		data []byte
		dat []string
		err  error
	)

	data, err = ioutil.ReadFile("data/WMM2015v2TestValues.txt")
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

		date = wmm.DecimalYear(dd)
		if dd, err = strconv.ParseFloat(dat[1], 64); err != nil {
			panic(err)
		}
		height = units.Meters(dd*1000)
		if dd, err = strconv.ParseFloat(dat[2], 64); err != nil {
			panic(err)
		}
		lat = units.Degrees(dd)
		if dd, err = strconv.ParseFloat(dat[3], 64); err != nil {
			panic(err)
		}
		lon = units.Degrees(dd)
		loc := wmm.Geodetic{Latitude: lat, Longitude: lon, Height: height}

		magS := wmm.CalculateWMMMagneticField(loc.ToSpherical(), date.ToTime())
		mag := magS.ToEllipsoidal(loc)

		if x, err = strconv.ParseFloat(dat[4], 64); err != nil {
			panic(err)
		}
		testDiff("X", wmm.MagneticField(mag).X, x, 0.1, t)
		if y, err = strconv.ParseFloat(dat[5], 64); err != nil {
			panic(err)
		}
		testDiff("Y", wmm.MagneticField(mag).Y, y, 0.1, t)
		if z, err = strconv.ParseFloat(dat[6], 64); err != nil {
			panic(err)
		}
		testDiff("Z", wmm.MagneticField(mag).Z, z, 0.1, t)
		if h, err = strconv.ParseFloat(dat[7], 64); err != nil {
			panic(err)
		}
		testDiff("H", wmm.MagneticField(mag).H(), h, 0.1, t)
		if f, err = strconv.ParseFloat(dat[8], 64); err != nil {
			panic(err)
		}
		testDiff("F", wmm.MagneticField(mag).F(), f, 0.1, t)
		if i, err = strconv.ParseFloat(dat[9], 64); err != nil {
			panic(err)
		}
		testDiff("I", wmm.MagneticField(mag).I(), i, 0.01, t)
		if d, err = strconv.ParseFloat(dat[10], 64); err != nil {
			panic(err)
		}
		testDiff("D", wmm.MagneticField(mag).D(), d, 0.01, t)
		//if gv, err = strconv.ParseFloat(dat[11], 64); err != nil {
		//	panic(err)
		//}
		//testDiff("GV", wmm.MagneticField(mag).GV(), gv, 0.01, t)
		if xdot, err = strconv.ParseFloat(dat[12], 64); err != nil {
			panic(err)
		}
		testDiff("Xdot", wmm.MagneticField(mag).DX, xdot, 0.1, t)
		if ydot, err = strconv.ParseFloat(dat[13], 64); err != nil {
			panic(err)
		}
		testDiff("Ydot", wmm.MagneticField(mag).DY, ydot, 0.1, t)
		if zdot, err = strconv.ParseFloat(dat[14], 64); err != nil {
			panic(err)
		}
		testDiff("Zdot", wmm.MagneticField(mag).DZ, zdot, 0.1, t)
		if hdot, err = strconv.ParseFloat(dat[15], 64); err != nil {
			panic(err)
		}
		testDiff("Hdot", wmm.MagneticField(mag).DH(), hdot, 0.1, t)
		if fdot, err = strconv.ParseFloat(dat[16], 64); err != nil {
			panic(err)
		}
		testDiff("Fdot", wmm.MagneticField(mag).DF(), fdot, 0.1, t)
		if idot, err = strconv.ParseFloat(dat[17], 64); err != nil {
			panic(err)
		}
		testDiff("Idot", wmm.MagneticField(mag).DI(), idot, 0.01, t)
		if ddot, err = strconv.ParseFloat(dat[18], 64); err != nil {
			panic(err)
		}
		testDiff("Ddot", wmm.MagneticField(mag).DD(), ddot, 0.01, t)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

}