// Package wmm provides a representation of the World Magnetic Model (WMM),
// a mathematical model of the magnetic field produced by the Earth's core and
// its variation over time.
//
// WMM is the magnetic model component of the World Geodetic System (WGS84).
// It consists of n=m=12 spherical harmonic coefficients as published by the
// National Geospatial-Intelligence Agency (NGA).
//
// This model evaluates all magnetic field components and their rates of change
// for any location on the Earth's surface.  These field components include the
// X, Y, and Z values and the declination D and inclination I.
// The Declination is used, for example, in correcting a Magnetic Heading to a
// True Heading.
package wmm

import (
	"math"
	"time"

	"github.com/westphae/geomag/pkg/egm96"
)

const (
	AGeo = 6371200 // Geomagnetic Reference Radius
)

// MagneticField represents a geomagnetic field and its rate of change.
type MagneticField struct {
	l          egm96.Location
	x, y, z    float64
	dx, dy, dz float64
	errors     ErrorModel // copied from the source Model at evaluation time
}

// ErrorModel returns the global-average uncertainty values for the WMM
// release that produced this field. See ErrorModel for the meaning of each
// component.
func (m MagneticField) ErrorModel() ErrorModel {
	return m.errors
}

// Ellipsoidal returns the magnetic field in ellipsoidal coordinate axes.
//
// The Ellipsoidal axes are the most commonly desired axes, in which the
// horizontal directions are parallel to the WGS84 ellipsoid.
//
// Field strengths are in nT and field strength changes in nT/Year.
func (m MagneticField) Ellipsoidal() (x, y, z, dx, dy, dz float64) {
	latS, _, _ := m.l.Spherical()
	latG, _, _ := m.l.Geodetic()
	cosDPhi := math.Cos(latS - latG)
	sinDPhi := math.Sin(latS - latG)
	x = m.x*cosDPhi - m.z*sinDPhi
	y = m.y
	z = m.x*sinDPhi + m.z*cosDPhi
	dx = m.dx*cosDPhi - m.dz*sinDPhi
	dy = m.dy
	dz = m.dx*sinDPhi + m.dz*cosDPhi
	return x, y, z, dx, dy, dz
}

// Spherical returns the magnetic field in spherical coordinate axes.
//
// The spherical axes are centered on the Earth's center of mass.
// These axes won't typically be used for navigation on or near the
// Earth's surface, but might be used in space.
//
// Field strengths are in nT and field strength changes in nT/Year.
func (m MagneticField) Spherical() (x, y, z, dx, dy, dz float64) {
	return m.x, m.y, m.z, m.dx, m.dy, m.dz
}

// H returns the strength of the magnetic field in the horizontal
// direction, i.e. the component parallel to the WGS84 ellipsoid.
//
// The return value is in nT.
func (m MagneticField) H() (h float64) {
	x, y, _, _, _, _ := m.Ellipsoidal()
	return math.Sqrt(x*x + y*y)
}

// F returns the total strength of the magnetic field.
//
// The return value is in nT.
func (m MagneticField) F() (f float64) {
	x, y, z, _, _, _ := m.Spherical()
	return math.Sqrt(x*x + y*y + z*z)
}

// I returns the Inclination of the magnetic field relative to the WGS84
// ellipsoid.
//
// The inclination is the angle the field makes relative to the horizontal,
// e.g. at the Magnetic North Pole, the field has a 90 degree inclination
// and points straight down.
//
// The return value is in degrees.
func (m MagneticField) I() (f float64) {
	_, _, z, _, _, _ := m.Ellipsoidal()
	return math.Atan2(z, m.H()) / egm96.Deg
}

// D returns the Declination of the magnetic field relative to the WGS84
// ellipsoid.
//
// The declination is the angle the field makes relative to True North.
// This is the most often-used value provided for the WMM for near-Earth
// navigation.  To convert Magnetic North to True North:
//  d := field.D()
//  TrueNorth := Magnetic_North + d
//
// The return value is in degrees.
func (m MagneticField) D() (f float64) {
	x, y, _, _, _, _ := m.Ellipsoidal()
	return math.Atan2(y, x) / egm96.Deg
}

// GV returns the Grid Variation of the magnetic field.
//
// It is useful for specifying the magnetic field near the field poles.
//
// The return value is in degrees.
func (m MagneticField) GV(loc egm96.Location) (f float64) {
	f = m.D()
	lat, lng, _ := loc.Geodetic()
	if lat > 55*egm96.Deg {
		f -= lng / egm96.Deg
	}
	if lat < -55*egm96.Deg {
		f += lng / egm96.Deg
	}
	return f
}

// DH returns the rate of change of the strength of the magnetic field in the
// horizontal direction, i.e. the component parallel to the WGS84 ellipsoid.
//
// The return value is in nT/yr.
func (m MagneticField) DH() (h float64) {
	x, y, _, dx, dy, _ := m.Ellipsoidal()
	return (x*dx + y*dy) / m.H()
}

// DF returns the rate of change of the total strength of the magnetic field.
//
// The return value is in nT/yr.
func (m MagneticField) DF() (f float64) {
	x, y, z, dx, dy, dz := m.Ellipsoidal()
	return (x*dx + y*dy + z*dz) / m.F()
}

// DI returns the rate of change of the Inclination of the magnetic field
// relative to the WGS84 ellipsoid.
//
// The return value is in degrees/yr.
func (m MagneticField) DI() (f float64) {
	f = m.F()
	_, _, z, _, _, dz := m.Ellipsoidal()
	return (m.H()*dz - m.DH()*z) / (f * f) / egm96.Deg
}

// DD returns the rate of change of the Declination of the magnetic field
// relative to the WGS84 ellipsoid.
//
// The return value is in degrees/yr.
func (m MagneticField) DD() (f float64) {
	f = m.H()
	x, y, _, dx, dy, _ := m.Ellipsoidal()
	return (x*dy - dx*y) / (f * f) / egm96.Deg
}

// DGV returns the rate of change of the Grid Variation of the magnetic field.
//
// The return value is in degrees/yr.
func (m MagneticField) DGV() (f float64) {
	return m.DD()
}

// ErrX returns the uncertainty in the X component of the magnetic field.
//
// The WMM specifies this uncertainty as an average over the global surface.
func (m MagneticField) ErrX() float64 { return m.errors.X }

// ErrY returns the uncertainty in the Y component of the magnetic field.
//
// The WMM specifies this uncertainty as an average over the global surface.
func (m MagneticField) ErrY() float64 { return m.errors.Y }

// ErrZ returns the uncertainty in the Z component of the magnetic field.
//
// The WMM specifies this uncertainty as an average over the global surface.
func (m MagneticField) ErrZ() float64 { return m.errors.Z }

// ErrF returns the uncertainty in the total magnetic field F.
//
// The WMM specifies this uncertainty as an average over the global surface.
func (m MagneticField) ErrF() float64 { return m.errors.F }

// ErrH returns the uncertainty in the horizontal component H of the magnetic field.
//
// The WMM specifies this uncertainty as an average over the global surface.
func (m MagneticField) ErrH() float64 { return m.errors.H }

// ErrI returns the uncertainty in the inclination I of the magnetic field.
//
// The WMM specifies this uncertainty as an average over the global surface.
func (m MagneticField) ErrI() float64 { return m.errors.I }

// ErrD returns the uncertainty in the Declination of the magnetic field at
// the given location.
//
// All other reported model uncertainties are given as the surface average.
// Because the H field can be close to zero near the poles, the D uncertainty
// can become very large and must be reported by location.
func (m MagneticField) ErrD() float64 {
	h := m.H()
	return math.Sqrt(m.errors.DA*m.errors.DA + m.errors.DB*m.errors.DB/(h*h))
}

// CalculateWMMMagneticField returns the magnetic field at the input location
// and time, using the package-level default Model.
//
// The WMM is valid at heights from -1km to +850km relative to the WGS84
// ellipsoid, and within five years of the model's epoch. Outside that range
// the function returns an informational error along with the calculated field.
//
// For batch evaluation across many points, prefer Default().MagneticField or
// hold a *Model directly — both expose a per-location cache that makes
// time-sweep loops at a fixed location effectively free.
func CalculateWMMMagneticField(loc egm96.Location, t time.Time) (MagneticField, error) {
	return Default().MagneticField(loc, t)
}
