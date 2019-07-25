package units

type Meters float64

type Degrees float64

const Deg = 57.29577951308232

type Location struct {
	Latitude Degrees
	Longitude Degrees
	Height Meters
}

func MetersFromFeet(ft float64) (m Meters) {
	return Meters(0.3048*ft)
}

func (m Meters) ToFeet() (ft float64) {
	return float64(m)/0.3048
}

func DegreesFromDMS(d, m, s float64) (dd Degrees) {
	var sgn float64 = 1
	if d<0 {
		sgn = -1
	}
	return Degrees(d+sgn*(m+s/60)/60)
}

func (dd Degrees) ToDMS() (d, m, s float64) {
	var sgn float64 = 1
	if dd<0 {
		sgn = -1
	}
	z := float64(dd)
	d = float64(int(z))
	z = (z-d)*60
	m = float64(int(z))
	s = (z-m)*60
	return d, sgn*m, sgn*s
}
