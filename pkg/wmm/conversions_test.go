package wmm

import (
	"fmt"
	"testing"
	"time"
)

func TestDecimalYearsToTime(t *testing.T) {
	ys := []DecimalYear{1995.0, 1996-1.0/365, 1997-1.0/366, 2004.0, 2017.5}
	ts := []time.Time{
		time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1995, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(1996, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2004, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2017, 7, 2, 12, 0, 0, 0, time.UTC),
	}
	for i, y := range ys {
		d := y.ToTime()
		testDiff(fmt.Sprintf("%5.3f to date", y), float64(ts[i].Unix()), float64(d.Unix()), 1, t)
	}
}

func TestTimeToDecimalYears(t *testing.T) {
	ys := []DecimalYear{1995.0, 1996-1.0/365, 1997-1.0/366, 2004.0, 2017.5}
	ts := []time.Time{
		time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1995, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(1996, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2004, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2017, 7, 2, 12, 0, 0, 0, time.UTC),
	}
	for i, tt := range ts {
		d := TimeToDecimalYears(tt)
		testDiff(fmt.Sprintf("%v to decimal year", tt), float64(d), float64(ys[i]), 0.001, t)
	}
}

func TestTimeToDecimalYearRoundTrips(t *testing.T) {
	ys := []DecimalYear{1995.0, 1996-1.0/365, 1997-1.0/366, 2004.0, 2017.5}
	ts := []time.Time{
		time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1995, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(1996, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2004, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2017, 6, 29, 0, 0, 0, 0, time.UTC),
	}
	for _, y := range ys {
		yy := TimeToDecimalYears(y.ToTime())
		testDiff(fmt.Sprintf("%6.3f to time and back", y), float64(yy), float64(y), 0.001, t)
	}
	for _, s := range ts {
		tt := TimeToDecimalYears(s).ToTime()
		testDiff(fmt.Sprintf("%v to decimal year and back", s), float64(tt.Unix()), float64(s.Unix()), 1, t)
	}
}
