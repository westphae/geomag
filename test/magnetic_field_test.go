package main

import (
	"github.com/westphae/geomag/pkg/units"
	"github.com/westphae/geomag/pkg/wmm"
	"testing"
)

func TestMagneticFieldFromPaper(t *testing.T) {
	tt := wmm.DecimalYear(2017.5)
	loc := wmm.Geodetic{
		Latitude: units.Degrees(-80),
		Longitude: units.Degrees(240),
		Height: units.Meters(100e6),
	}
}