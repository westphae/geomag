package parsing

import "testing"

const (
	eps = 1e-9
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

func TestDMSGood(t *testing.T) {
	inps := []string{
		"3.123", "-12.567",
		"0 0 0", "-150 59 59.5",
		"5,30,15", "-18,30, 30.9",
		"S112.531", "E89.183",
		"N5 9 0", "W011 29 31",
		"N15,15,15", "E11, 12, 13",
	}
	outs := []float64{
		3.123, -12.567,
		0, -150.999861111,
		5.504166666, -18.508583333,
		-112.531, 89.183,
		5.15, -11.491944444,
		15.254166666, 11.203611111,
	}

	for i, inp := range inps {
		out, err := ParseLatLng(inp)
		if err!=nil {
			t.Errorf("ParseLatLng got error %s", err)
		}
		testDiff(inp, out, outs[i], eps, t)
	}
}

func TestDMSBad(t *testing.T) {
	inps := []string{
		"NW3.123", "-E12.567",
		"0 0", "-150 59.2 59",
		"0,,0,5", "0,1,0,0", "0, 2", "0,,1", "-1,-2,-3",
		"5,61,0", "5,59,60",
		"ABC123",
	}

	for _, inp := range inps {
		_, err := ParseLatLng(inp)
		if err==nil {
			t.Errorf("%sParseLatLng incorrectly thought it could parse %s%s", red, inp, reset)
		} else {
			t.Logf("%sParseLatLng correctly rejected %s%s", green, inp, reset)
		}
	}
}

func TestAltitudeGood(t *testing.T) {
	inps := []string{
		"99.95",
		"-123.45",
		"E54.22",
		"E-800.2",
	}
	outs := []float64{
		99.95,
		-123.45,
		54.22,
		-800.2,
	}
	outh := []bool{
		false,
		false,
		true,
		true,
	}

	for i, inp := range inps {
		out, hae, err := ParseAltitude(inp)
		if err!=nil {
			t.Errorf("ParseLatLng got error %s", err)
		}
		if hae!=outh[i] {
			t.Errorf("%sParseAltitude got hae wrong for %s%s", red, inp, reset)
		} else {
			t.Logf("%sParseAltitude got hae correct for %s%s", green, inp, reset)
		}
		testDiff(inp, out, outs[i], eps, t)
	}
}

func TestAltitudeBad(t *testing.T) {
	inps := []string{
		"-E111.2",
		"EE12",
		"99E99",
		"ABC123",
	}

	for _, inp := range inps {
		_, _, err := ParseAltitude(inp)
		if err==nil {
			t.Errorf("%sParseAltitude incorrectly thought it could parse %s%s", red, inp, reset)
		} else {
			t.Logf("%sParseAltitude correctly rejected %s%s", green, inp, reset)
		}
	}
}
