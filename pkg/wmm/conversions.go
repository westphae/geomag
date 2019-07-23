package wmm

import (
	"math"
	"time"

	"github.com/westphae/geomag/pkg/units"
)

const (
	A units.Meters = 6378137
	F = 1/298.257223563
	E2 = F*(2-F)
	DaysPerYear = 365.2425
	SecondsPerDay = 86400
	SecondsPerYear = SecondsPerDay*DaysPerYear
)

type Geodetic units.Location

type Spherical units.Location

func (l Geodetic) ToSpherical() (s Spherical) {
	sinPhi := math.Sin(float64(l.Latitude))
	cosPhi := math.Cos(float64(l.Latitude))
	h := float64(l.Height)
	rc := float64(A)/math.Sqrt(1-E2*sinPhi*sinPhi)
	p := (rc+h)*cosPhi
	z := (rc*(1-E2)+h)*sinPhi
	r := math.Sqrt(p*p+z*z)
	return Spherical{
		Latitude: units.Degrees(math.Asin(z/r)),
		Longitude: l.Longitude,
		Height: units.Meters(r),
	}
}

// DecimalYearsToTime converts an epoch-like float64 year like 2015.0
// to a Go time.Time.  It is accurate only to the nearest day.
func DecimalYearsToTime(y float64) (t time.Time) {
	tY := int(y)
	tD := int((y-float64(tY))*365.2425+0.5)
	return time.Date(tY, 1, tD, 0, 0, 0, 0, time.UTC)
}

func DecimalYearsSinceEpoch(t time.Time, epoch time.Time) (y float64) {
	y = t.Sub(epoch).Seconds()
	return y/SecondsPerYear
}