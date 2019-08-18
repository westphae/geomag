package parsing

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseLatLng takes an input string in various forms, represent a latitude
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
		inp = inp[1:len(inp)]
	}

	if ct:=strings.Count(inp, ","); ct!=0 && ct!=2 {
		return 0, fmt.Errorf("incorrect number of fields in %s", inp)
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
