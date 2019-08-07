package wmm

import (
	"github.com/westphae/geomag/pkg/egm96"
	"math"
	"time"

	"github.com/westphae/geomag/pkg/polynomial"
)

const AGeo egm96.Meters = 6371200 // Geomagnetic Reference Radius

type MagneticField struct {
	X, Y, Z    float64
	DX, DY, DZ float64
}

type GeocentricMagneticField MagneticField

type EllipsoidalMagneticField MagneticField

func (m MagneticField) H() (h float64) {
	return math.Sqrt(m.X*m.X + m.Y*m.Y)
}

func (m MagneticField) F() (f float64) {
	h := m.H()
	return math.Sqrt(h*h + m.Z*m.Z)
}

func (m MagneticField) I() (f float64) {
	return math.Atan2(m.Z, m.H())/ egm96.Deg
}

func (m MagneticField) D() (f float64) {
	return math.Atan2(m.Y, m.X)/ egm96.Deg
}

func (m MagneticField) GV(loc egm96.Geodetic) (f float64) {
	f = m.D()
	if loc.latitude > 55 {
		f -= float64(loc.longitude)
	}
	if loc.latitude < -55 {
		f += float64(loc.longitude)
	}
	return f
}

func (m MagneticField) DH() (h float64) {
	return (m.X*m.DX + m.Y*m.DY)/m.H()
}

func (m MagneticField) DF() (f float64) {
	return (m.X*m.DX + m.Y*m.DY + m.Z*m.DZ)/m.F()
}

func (m MagneticField) DI() (f float64) {
	f = m.F()
	return (m.H()*m.DZ - m.DH()*m.Z)/(f*f)/ egm96.Deg
}

func (m MagneticField) DD() (f float64) {
	f = m.H()
	return (m.X*m.DY - m.DX*m.Y)/(f*f)/ egm96.Deg
}

func (m MagneticField) DGV() (f float64) {
	return m.DD()
}

var (
	curLoc   egm96.Spherical
	curField GeocentricMagneticField
)

func CalculateWMMMagneticField(loc egm96.Spherical, t time.Time) (field GeocentricMagneticField, err error) {
	if loc.longitude !=curLoc.longitude || loc.latitude !=curLoc.latitude || loc.height !=curLoc.height {
		curLoc = loc
		curField = *new(GeocentricMagneticField)
		phi := float64(loc.latitude)* egm96.Deg
		lambda := float64(loc.longitude)* egm96.Deg
		sinPhi := math.Sin(phi)
		cosPhi := math.Cos(phi)
		var g, h, dg, dh float64
		for n:=1; n<=MaxLegendreOrder; n++ {
			nn := float64(n+1)
			// if height varies, recalculate from here
			f := polynomial.Pow(float64(AGeo/loc.height), n+2)
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
				curField.X += -f*(g*cosMLambda+h*sinMLambda)*dp
				curField.Y += f/cosPhi*mf*(g*sinMLambda-h*cosMLambda)*p
				curField.Z += -nn*f*(g*cosMLambda+h*sinMLambda)*p
				curField.DX += -f*(dg*cosMLambda+dh*sinMLambda)*dp
				curField.DY += f/cosPhi*mf*(dg*sinMLambda-dh*cosMLambda)*p
				curField.DZ += -nn*f*(dg*cosMLambda+dh*sinMLambda)*p
			}
		}
	}
	dt := float64(TimeToDecimalYears(t)- Epoch)
	field.X = curField.X + dt*curField.DX
	field.Y = curField.Y + dt*curField.DY
	field.Z = curField.Z + dt*curField.DZ
	field.DX = curField.DX
	field.DY = curField.DY
	field.DZ = curField.DZ
	return field, err
}
