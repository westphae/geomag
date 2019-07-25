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

type DecimalYear float64

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

// DecimalYearsToTime converts an epoch-like float64 year like 2015.0 to a Go time.Time.
// Per document MIL-PRF-89500B Section 3.2, "Time is referenced in decimal years
// (e.g., 15 May 2019 is 2019.367). Note that the day-of-year (DOY) of January 1st is zero
// and December 31st is 364 for non-leap year. For a leap year, DOY of December 31st is 365."
func (y DecimalYear) ToTime() (t time.Time) {
	tYear := int(y)
	yearDays := float64(time.Date(tYear, 12, 31, 0, 0, 0, 0, time.UTC).YearDay())
	tDays := int((float64(y)-float64(tYear))*yearDays+0.5)
	return time.Date(tYear, 1, tDays+1, 0, 0, 0, 0, time.UTC)
}

// TimeToDecimalYears converts a Go time.Time to an epoch-like float64 year like 2015.0.
// Per document MIL-PRF-89500B Section 3.2, "Time is referenced in decimal years
// (e.g., 15 May 2019 is 2019.367). Note that the day-of-year (DOY) of January 1st is zero
// and December 31st is 364 for non-leap year. For a leap year, DOY of December 31st is 365."
func TimeToDecimalYears(t time.Time) (y DecimalYear) {
	tYear := t.Year()
	tDay := float64(t.YearDay()-1)
	yearDays := float64(time.Date(tYear, 12, 31, 0, 0, 0, 0, time.UTC).YearDay())
	return DecimalYear(tYear) + DecimalYear(tDay/yearDays)
}

// DecimalYearsSinceEpoch converts a Go time.Time (like time.Now()) to a decimal number
// of years since the epoch, per MIL-PRF-89500B Section 3.2.
func DecimalYearsSinceEpoch(t time.Time, epoch time.Time) (y DecimalYear) {
	y = DecimalYear(t.Sub(epoch).Seconds()/SecondsPerYear)
	return y
}

func (f GeocentricMagneticField) ToEllipsoidal(l Geodetic) (g EllipsoidalMagneticField) {
	ll := l.ToSpherical()
	cosDPhi := math.Cos(float64(ll.Latitude-l.Latitude))
	sinDPhi := math.Sin(float64(ll.Latitude-l.Latitude))
	g.X = f.X*cosDPhi - f.Z*sinDPhi
	g.Y = f.Y
	g.Z = f.X*sinDPhi + f.Z*cosDPhi
	g.DX = f.DX*cosDPhi - f.DZ*sinDPhi
	g.DY = f.DY
	g.DZ = f.DX*sinDPhi + f.DZ*cosDPhi
	return g
}