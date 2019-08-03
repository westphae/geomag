package wmm

import (
	"fmt"
	"testing"
	"time"
)

func TestDecimalYearsToTime(t *testing.T) {
	ys := []DecimalYear{1995.0, 1996-1.0/364, 1997-1.0/365, 2004.0, 2019.367}
	ts := []time.Time{
		time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1995, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(1996, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2004, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2019, 5, 15, 0, 0, 0, 0, time.UTC),
	}
	for i, y := range ys {
		d := y.ToTime()
		testDiff(fmt.Sprintf("%5.1f to date", y), float64(ts[i].Unix()), float64(d.Unix()), 0.5*SecondsPerDay, t)
	}
}

func TestTimeToDecimalYears(t *testing.T) {
	ys := []DecimalYear{1995.0, 1996-1.0/365, 1997-1.0/366, 2004.0, 2019.367}
	ts := []time.Time{
		time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1995, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(1996, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2004, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2019, 5, 15, 0, 0, 0, 0, time.UTC),
	}
	for i, tt := range ts {
		d := TimeToDecimalYears(tt)
		testDiff(fmt.Sprintf("%v to decimal year", tt), float64(d), float64(ys[i]), 0.001, t)
	}
}

func TestTimeToDecimalYearRoundTrips(t *testing.T) {
	ys := []DecimalYear{1995.0, 1996-1.0/365, 1997-1.0/366, 2004.0, 2019.367}
	ts := []time.Time{
		time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1995, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(1996, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2004, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2019, 5, 15, 0, 0, 0, 0, time.UTC),
	}
	for _, y := range ys {
		yy := TimeToDecimalYears(y.ToTime())
		testDiff(fmt.Sprintf("%6.2f to time and back", yy), float64(yy), float64(y), 0.001, t)
	}
	for _, s := range ts {
		tt := TimeToDecimalYears(s).ToTime()
		testDiff(fmt.Sprintf("%v to decimal year and back", tt), float64(tt.Unix()), float64(s.Unix()), 0.5*SecondsPerDay, t)
	}
}

func TestDecimalYearsSinceEpoch(t *testing.T) {
	dEpoch := DecimalYear(1995)
	epoch := dEpoch.ToTime()
	ts := []time.Time{
		time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2004, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2005, 7, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2007, 10, 1, 0, 0, 0, 0, time.UTC),
	}
	ys := []DecimalYear{0, 9, 10.5, 12.75}
	for i, y := range ts {
		z := DecimalYearsSinceEpoch(y, epoch)
		testDiff(fmt.Sprintf("decimal years since epoch %5.1f of %v", dEpoch, y),
			float64(z), float64(ys[i]), 0.01, t)
	}
}
