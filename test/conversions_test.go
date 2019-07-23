package main

import (
	"testing"
	"time"

	"github.com/westphae/geomag/pkg/wmm"
)

func TestDecimalYearsToTime(t *testing.T) {
	ys := []float64{1995.0, 2004.0, 2005.5, 2007.75}
	ts := []time.Time{
		time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2004, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2005, 7, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2007, 10, 1, 0, 0, 0, 0, time.UTC),
	}
	for i, y := range ys {
		d := wmm.DecimalYearsToTime(y)
		dd := float64(ts[i].Sub(d).Seconds()/wmm.SecondsPerDay)
		if dd < -1 || dd > 1 {
			t.Errorf("Conversion of %5.1f to date failed, expected %v, got %v", y, ts[i], d)
		}
	}
}

func TestDecimalYearsSinceEpoch(t *testing.T) {
	dEpoch := 1995.0
	epoch := wmm.DecimalYearsToTime(dEpoch)
	ts := []time.Time{
		time.Date(1995, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2004, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2005, 7, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2007, 10, 1, 0, 0, 0, 0, time.UTC),
	}
	ys := []float64{0, 9, 10.5, 12.75}
	for i, y := range ts {
		z := wmm.DecimalYearsSinceEpoch(y, epoch)
		dz := z - ys[i]
		if dz < -0.01 || dz > 0.01 {
			t.Errorf("calculation of decimal years since epoch %5.1f of %v failed, got %5.1f, expected %5.1f",
				dEpoch, y, z, ys[i])
		}
	}
}
