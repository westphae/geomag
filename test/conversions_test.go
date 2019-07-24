package main

import (
	"testing"
	"time"

	"github.com/westphae/geomag/pkg/wmm"
)

func TestDecimalYearsToTime(t *testing.T) {
	ys := []wmm.DecimalYear{1995.0, 1996-1.0/364, 1997-1.0/365, 2004.0, 2019.367}
	ts := []time.Time{
		time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1995, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(1996, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2004, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2019, 5, 15, 0, 0, 0, 0, time.UTC),
	}
	for i, y := range ys {
		d := y.ToTime()
		dd := float64(ts[i].Sub(d).Seconds()/wmm.SecondsPerDay)
		if dd < -0.5 || dd > 0.5 {
			t.Errorf("Conversion of %5.1f to date failed, expected %v, got %v", y, ts[i], d)
		}
	}
}

func TestTimeToDecimalYears(t *testing.T) {
	ys := []wmm.DecimalYear{1995.0, 1996-1.0/365, 1997-1.0/366, 2004.0, 2019.367}
	ts := []time.Time{
		time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1995, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(1996, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2004, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2019, 5, 15, 0, 0, 0, 0, time.UTC),
	}
	for i, tt := range ts {
		d := wmm.TimeToDecimalYears(tt)
		dd := d - ys[i]
		if dd < -0.001 || dd > 0.001 {
			t.Errorf("Conversion of %v to decimal year failed, expected %8.4f, got %8.4f", tt, ys[i], d)
		}
	}
}

func TestTimeToDecimalYearRoundTrips(t *testing.T) {
	ys := []wmm.DecimalYear{1995.0, 1996-1.0/365, 1997-1.0/366, 2004.0, 2019.367}
	ts := []time.Time{
		time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(1995, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(1996, 12, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2004, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2019, 5, 15, 0, 0, 0, 0, time.UTC),
	}
	for _, y := range ys {
		yy := wmm.TimeToDecimalYears(y.ToTime())
		dy := yy - y
		if dy < -0.001 || dy > 0.001 {
			t.Errorf("Round trip conversion of %8.4f failed, got %8.4f", y, yy)
		}
	}
	for _, s := range ts {
		tt := wmm.TimeToDecimalYears(s).ToTime()
		dt := tt.Sub(s).Seconds()/wmm.SecondsPerDay
		if dt < -0.5 || dt > 0.5 {
			t.Errorf("Round trip conversion of %v failed, got %v", s, tt)
		}
	}
}

func TestDecimalYearsSinceEpoch(t *testing.T) {
	dEpoch := wmm.DecimalYear(1995)
	epoch := dEpoch.ToTime()
	ts := []time.Time{
		time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2004, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2005, 7, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2007, 10, 1, 0, 0, 0, 0, time.UTC),
	}
	ys := []wmm.DecimalYear{0, 9, 10.5, 12.75}
	for i, y := range ts {
		z := wmm.DecimalYearsSinceEpoch(y, epoch)
		dz := z - ys[i]
		if dz < -0.01 || dz > 0.01 {
			t.Errorf("calculation of decimal years since epoch %5.1f of %v failed, got %5.1f, expected %5.1f",
				dEpoch, y, z, ys[i])
		}
	}
}
