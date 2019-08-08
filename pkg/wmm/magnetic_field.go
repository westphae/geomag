package wmm

import (
	"math"
	"time"

	"github.com/westphae/geomag/pkg/egm96"
	"github.com/westphae/geomag/pkg/polynomial"
)

const AGeo = 6371200 // Geomagnetic Reference Radius

// MagneticField represents a geomagnetic field and its rate of change.
type MagneticField struct {
	l egm96.Location
	x, y, z    float64
	dx, dy, dz float64
}

// Ellipsoidal returns the magnetic field in ellipsoidal coordinate axes.
// Field strengths are in nT and field strength changes in nT/Year.
func (m MagneticField) Ellipsoidal() (x, y, z, dx, dy, dz float64) {
	latS, _, _ := m.l.Spherical()
	latG, _, _ := m.l.Geodetic()
	cosDPhi := math.Cos(latS-latG)
	sinDPhi := math.Sin(latS-latG)
	x = m.x*cosDPhi - m.z*sinDPhi
	y = m.y
	z = m.x*sinDPhi + m.z*cosDPhi
	dx = m.dx*cosDPhi - m.dz*sinDPhi
	dy = m.dy
	dz = m.dx*sinDPhi + m.dz*cosDPhi
	return x, y, z, dx, dy, dz
}

// Spherical returns the magnetic field in spherical coordinate axes.
// Field strengths are in nT and field strength changes in nT/Year.
func (m MagneticField) Spherical() (x, y, z, dx, dy, dz float64) {
	return m.x, m.y, m.z, m.dx, m.dy, m.dz
}

func (m MagneticField) H() (h float64) {
	x, y, _, _, _, _ := m.Ellipsoidal()
	return math.Sqrt(x*x + y*y)
}

func (m MagneticField) F() (f float64) {
	x, y, z, _, _, _ := m.Ellipsoidal()
	return math.Sqrt(x*x + y*y + z*z)
}

func (m MagneticField) I() (f float64) {
	_, _, z, _, _, _ := m.Ellipsoidal()
	return math.Atan2(z, m.H())/egm96.Deg
}

func (m MagneticField) D() (f float64) {
	x, y, _, _, _, _ := m.Ellipsoidal()
	return math.Atan2(y, x)/egm96.Deg
}

func (m MagneticField) GV(loc egm96.Location) (f float64) {
	f = m.D()
	lat, lng, _ := loc.Geodetic()
	if lat > 55*egm96.Deg {
		f -= lng/egm96.Deg
	}
	if lat < -55*egm96.Deg {
		f += lng/egm96.Deg
	}
	return f
}

func (m MagneticField) DH() (h float64) {
	x, y, _, dx, dy, _ := m.Ellipsoidal()
	return (x*dx + y*dy)/m.H()
}

func (m MagneticField) DF() (f float64) {
	x, y, z, dx, dy, dz := m.Ellipsoidal()
	return (x*dx + y*dy + z*dz)/m.F()
}

func (m MagneticField) DI() (f float64) {
	f = m.F()
	_, _, z, _, _, dz := m.Ellipsoidal()
	return (m.H()*dz - m.DH()*z)/(f*f)/ egm96.Deg
}

func (m MagneticField) DD() (f float64) {
	f = m.H()
	x, y, _, dx, dy, _ := m.Ellipsoidal()
	return (x*dy - dx*y)/(f*f)/ egm96.Deg
}

func (m MagneticField) DGV() (f float64) {
	return m.DD()
}

var (
	curLoc   egm96.Location // Spherical
	curField MagneticField
)

func CalculateWMMMagneticField(loc egm96.Location, t time.Time) (field MagneticField, err error) {
	if !loc.Equals(curLoc) {
		curLoc = loc
		curField = *new(MagneticField)
		phi, lambda, hh := loc.Spherical()
		sinPhi := math.Sin(phi)
		cosPhi := math.Cos(phi)
		var g, h, dg, dh float64
		for n:=1; n<=MaxLegendreOrder; n++ {
			nn := float64(n+1)
			// if height varies, recalculate from here
			f := polynomial.Pow(AGeo/hh, n+2)
			for m:=0; m<=n; m++ {
				mf := float64(m)
				// if latitude varies, recalculate from here
				p := polynomial.LegendreFunction(n, m, sinPhi)
				q := polynomial.LegendreFunction(n+1, m, sinPhi)
				if m>0 {
					p *= math.Sqrt(2/polynomial.FactorialRatioFloat(n+m, n-m))
					q *= math.Sqrt(2/polynomial.FactorialRatioFloat(n+m, n-m))
				}
				dp := nn*math.Tan(phi)*p - (nn-mf)/cosPhi*q
				g, h, dg, dh, err = GetWMMCoefficients(n, m, Epoch.ToTime())
				// if longitude varies, recalculate from here
				sinMLambda := math.Sin(mf*lambda)
				cosMLambda := math.Cos(mf*lambda)
				curField.x += -f*(g*cosMLambda+h*sinMLambda)*dp
				curField.y += f/cosPhi*mf*(g*sinMLambda-h*cosMLambda)*p
				curField.z += -nn*f*(g*cosMLambda+h*sinMLambda)*p
				curField.dx += -f*(dg*cosMLambda+dh*sinMLambda)*dp
				curField.dy += f/cosPhi*mf*(dg*sinMLambda-dh*cosMLambda)*p
				curField.dz += -nn*f*(dg*cosMLambda+dh*sinMLambda)*p
			}
		}
	}
	dt := float64(TimeToDecimalYears(t)- Epoch)
	field.l = loc
	field.x = curField.x + dt*curField.dx
	field.y = curField.y + dt*curField.dy
	field.z = curField.z + dt*curField.dz
	field.dx = curField.dx
	field.dy = curField.dy
	field.dz = curField.dz
	return field, err
}
