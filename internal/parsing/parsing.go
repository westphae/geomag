package parsing

import (
	"fmt"
	"github.com/westphae/geomag/pkg/wmm"
	"strconv"
	"strings"
	"time"
)

// ParseLatLng takes an input string in various forms, representing a latitude
// or longitude, and returns the float representation.
//
// Possible formats:
// [-]DDD.DDDD
// [-]DD MM SS.SSS
// [-]DD,MM,SS.SSS
// [NSEW]DD MM SS.SSS
// [NSEW]DD,MM,SS.SSS
func ParseLatLng(inp string) (l float64, err error) {
	sgn := 1.0
	ii:=strings.LastIndexAny(inp, "+-NSEW")
	if ii>0 {
		return 0, fmt.Errorf("%s prefix is invalid", inp[0:ii])
	}
	if ii==0 {
		if strings.ContainsAny(inp[0:1], "-SW") {
			sgn = -1
		}
		inp = inp[1:]
	}

	if ct:=strings.Count(inp, ","); ct!=0 && ct!=2 {
		return 0, fmt.Errorf("%s is not in the format D,M,S", inp)
	}
	inp = strings.ReplaceAll(inp, ",", " ")

	ls := strings.Fields(inp)
	if len(ls)==1 {
		l, err = strconv.ParseFloat(ls[0], 64)
		return l*sgn, err
	}
	if len(ls)!=3 {
		return 0, fmt.Errorf("%s is not in the format D M S", inp)
	}
	d, err := strconv.Atoi(ls[0])
	if err!=nil {
		return 0, err
	}
	m, err := strconv.Atoi(ls[1])
	if err!=nil {
		return 0, err
	}
	if m<0 || m>=60 {
		return 0, fmt.Errorf("minutes entry %s must be in the range [0,60)", ls[1])
	}
	s, err := strconv.ParseFloat(ls[2], 64)
	if err!=nil {
		return 0, err
	}
	if s<0 || s>=60 {
		return 0, fmt.Errorf("seconds entry %s must be in the range [0,60)", ls[2])
	}
	return sgn*(float64(d)+(float64(m)+s/60)/60), nil
}

// ParseAltitude takes an input string in various forms, representing an altitude,
// and returns the float representation and whether it's height above ellipsoid or
// height above sea level.
//
// Possible formats:
// [E][-]HHH.HHHH
func ParseAltitude(inp string) (height float64, hae bool, err error) {
	ii:=strings.LastIndex(inp, "E")
	if ii>0 {
		return 0, false, fmt.Errorf("%s prefix is invalid", inp[0:ii])
	}
	if ii==0 {
		hae = true
		inp = inp[1:]
	}
	height, err = strconv.ParseFloat(inp, 64)
	return
}

// ParseNOAAAltitude parses a single altitude token in NOAA's wmm_file
// convention: a single-character unit prefix (K = kilometers, M = meters,
// F = feet) followed by a signed decimal magnitude. The result is in meters.
//
// Examples: "K100" → 100000.0, "M1042" → 1042.0, "F30000" → 9144.0,
// "K-1" → -1000.0, "K1.3" → 1300.0.
func ParseNOAAAltitude(inp string) (meters float64, err error) {
	if len(inp) < 2 {
		return 0, fmt.Errorf("%q is too short to be a NOAA altitude (expected K/M/F prefix + number)", inp)
	}
	unit := inp[0]
	mag, err := strconv.ParseFloat(inp[1:], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid altitude magnitude %q: %w", inp[1:], err)
	}
	switch unit {
	case 'K', 'k':
		return mag * 1000, nil
	case 'M', 'm':
		return mag, nil
	case 'F', 'f':
		return mag * 0.3048, nil // international foot
	}
	return 0, fmt.Errorf("unknown altitude unit %q in %q (expected K, M, or F)", string(unit), inp)
}

// ParseTime takes an input string in various forms, representing a time,
// and returns the float representation.
//
// Possible formats:
// YYYY.yyy
// MM DD YYYY
// MM/DD/YYYY
func ParseTime(inp string) (dYear float64, err error) {
	inp = strings.ReplaceAll(inp, "/", " ")
	ls := strings.Fields(inp)
	if len(ls)==1 {
		return strconv.ParseFloat(inp, 64)
	}
	if len(ls)==3 {
		var y, m, d int
		y, err = strconv.Atoi(ls[2])
		if err!=nil {
			return 0, fmt.Errorf("invalid year: %s", ls[0])
		}
		m, err = strconv.Atoi(ls[0])
		if err!=nil || m<1 || m>12 {
			return 0, fmt.Errorf("invalid month: %s", ls[1])
		}
		mm := time.Month(m)
		d, err = strconv.Atoi(ls[1])
		if err!=nil {
			return 0, fmt.Errorf("invalid day: %s", ls[2])
		}
		return float64(wmm.TimeToDecimalYears(time.Date(y, mm, d,
			0, 0, 0, 0, time.UTC))), nil
	}
	return
}
