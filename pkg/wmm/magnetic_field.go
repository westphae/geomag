package wmm

import (
	"github.com/westphae/geomag/pkg/polynomial"
	"github.com/westphae/geomag/pkg/units"
	"math"
	"time"
)

type MagneticField struct {
	X, Y, Z    float64
	DX, DY, DZ float64
}

type GeocentricMagneticField MagneticField

type EllipsoidalMagneticField MagneticField

func CalculateWMMMagneticField(loc Spherical, t time.Time) (field GeocentricMagneticField) {
	phi := float64(loc.Latitude)/units.Deg
	lambda := float64(loc.Longitude)/units.Deg
	sinPhi := math.Sin(phi)
	cosPhi := math.Cos(phi)
	for n:=1; n<=MaxLegendreOrder; n++ {
		f := polynomial.Pow(float64(A/loc.Height), n+2)
		for m:=0; m<=n; m++ {
			mf := float64(m)
			sinMLambda := math.Sin(mf*lambda)
			cosMLambda := math.Cos(mf*lambda)
			p := polynomial.LegendreFunction(n, m, sinPhi)
			q := polynomial.LegendreFunction(n+1, m, sinPhi)
			g, h, dg, dh, err := GetWMMCoefficients(n, m, t)
			if err != nil {
				panic(err)
			}
			nn := float64(n+1)
			dp := nn*math.Tan(phi)*p - math.Sqrt(nn*nn-mf*mf)/math.Cos(phi)*q
			field.X += -f*(dg*cosMLambda+dh*sinMLambda)*dp
			field.Y += f/cosPhi*mf*(g*sinMLambda-h*cosMLambda)*p
			field.Z += -nn*f*(g*cosMLambda+h*sinMLambda)*p
			field.DX += -f*(g*cosMLambda+h*sinMLambda)*dp
			field.DY += f/cosPhi*mf*(dg*sinMLambda-dh*cosMLambda)*p
			field.DZ += -nn*f*(dg*cosMLambda+dh*sinMLambda)*p
		}
	}
	return field
}

func (m MagneticField) H() (h float64) {
	return math.Sqrt(m.X*m.X + m.Y*m.Y)
}

func (m MagneticField) F() (f float64) {
	h := m.H()
	return math.Sqrt(h*h + m.Z*m.Z)
}

func (m MagneticField) I() (f float64) {
	return math.Atan2(m.Z, m.H())
}

func (m MagneticField) D() (f float64) {
	return math.Atan2(m.Y, m.X)
}

func (m MagneticField) DH() (h float64) {
	return (m.X*m.DX + m.Y*m.DY)/m.H()
}

func (m MagneticField) DF() (f float64) {
	return (m.X*m.DX + m.Y*m.DY + m.Z*m.DZ)/m.F()
}

func (m MagneticField) DI() (f float64) {
	f = m.F()
	return (m.H()*m.DZ - m.DH()*m.Z)/(f*f)
}

func (m MagneticField) DD() (f float64) {
	f = m.H()
	return (m.X*m.DY - m.DX*m.Y)/(f*f)
}

func (m MagneticField) DGV() (f float64) {
	return m.DD()
}
