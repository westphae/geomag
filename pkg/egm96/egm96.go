// Package egm96 provides a representation of the 1996 Earth Gravitational Model (EGM96),
// a geopotential model of the Earth.
//
// EGM96 is the geoid reference model component of the World Geodetic System (WGS84).
// It consists of n=m=360 spherical harmonic coefficients as published by the
// National Geospatial-Intelligence Agency (NGA).  The NGA also publishes a raster grid
// of the calculated heights which can be interpolated to approximate the geoid height
// at any location.
//
// In effect, this model provides the height of sea level above the WGS84 reference ellipsoid.
// It is used, for example, in GPS navigation to provide the height above sea level.
//
// This package is based on the NGA-provided 15'x15' resolution grid encoding
// the heights of the geopotential surface at each lat/long, and interpolates between grid
// points using a bicubic Catmull-Rom spline (with bilinear fallback within one
// grid cell of the latitude poles, where the 4×4 stencil would step off the grid).
package egm96

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
)

// egm96LoadOnce guards loadEGM96Grid so it runs exactly once across all
// goroutines, even under concurrent first-use of HeightAboveMSL,
// NewLocationMSL, or NearestEGM96GridPoint. Fixes issue #4.
var egm96LoadOnce sync.Once

//go:embed embedded/ww15mgh.grd
var defaultGrid []byte

// Constants defining the WGS84 reference ellipsoid
const (
	A  = 6378137         // Equatorial radius of WGS84 reference ellipsoid in meters
	F  = 1/298.257223563 // Flattening of WGS84 reference ellipsoid
	E2 = F*(2-F)         // Eccentricity squared of WGS84 reference ellipsoid
)

// Location is a type that represents a position in space as represented
// by a latitude, a longitude and a height.
type Location struct {
	latitude  float64
	longitude float64
	height    float64
}

// normalizeLongitude wraps a longitude in degrees into [0, 360). Both the
// usual GPS/WGS84 [-180, 180] convention and any over-wound values like 540
// (= 180) round-trip correctly. Callers can rely on stored Location longitudes
// always being in [0, 360) regardless of input convention.
func normalizeLongitude(lng float64) float64 {
	lng = math.Mod(lng, 360)
	if lng < 0 {
		lng += 360
	}
	return lng
}

// NewLocationGeodetic returns a Location given an input latitude, longitude,
// and height specified in the Geodetic system.
//
// The Geodetic coordinate system is the usual latitude, longitude, and height
// above the WGS84 Reference Ellipsoid, i.e. as typically measured by a GPS
// receiver.
//
// Latitude and longitude are specified in decimal degrees and height in
// meters. Negative longitudes (e.g. -121 for the western hemisphere) are
// accepted and normalized to [0, 360); a subsequent (Location).Geodetic()
// call returns the normalized form.
//
// Geodetic coordinates are the un-primed variables φ,λ,h in the WMM paper.
func NewLocationGeodetic(latitude, longitude, height float64) (loc Location) {
	return Location{
		latitude:  latitude * Deg,
		longitude: normalizeLongitude(longitude) * Deg,
		height:    height,
	}
}

// NewLocationMSL returns a Location given an input latitude, longitude, and
// height above mean sea level.
//
// The latitude and longitude are as specified in the Geodetic Coordinate
// System, and the height is the height above mean sea level, NOT above the
// WGS84 Reference Ellipsoid.
//
// Latitude and longitude are specified in decimal degrees and height in
// meters. Negative longitudes are accepted and normalized to [0, 360).
func NewLocationMSL(latitude, longitude, height float64) (loc Location, err error) {
	egm96LoadOnce.Do(loadEGM96Grid)

	longitude = normalizeLongitude(longitude)
	nLng := int((longitude-egm96X0)/egm96DX) // Grid x just below desired x
	nLat := int((latitude-egm96Y0)/egm96DY)  // Grid y just below desired y

	// Bounds checks guard the bilinear fallback path used at the latitude
	// poles. The bicubic stencil reads a 4-cell window in latitude and is
	// guarded internally; the latitude/longitude rejection here matches the
	// historical bilinear contract so existing error messages still apply.
	if nLng < 0 || nLng > egm96XN-2 {
		return Location{}, fmt.Errorf("requested longitude %4.2f lies outside of EGM96 longitude range %4.1f to %4.1f",
			longitude, egm96X0, egm96X1)
	}
	if nLat < 0 || nLat > egm96YN-2 {
		return Location{}, fmt.Errorf("requested latitude %4.2f lies outside of EGM96 latitude range %4.1f to %4.1f",
			latitude, egm96Y0, egm96Y1)
	}

	gx := (longitude - egm96X0) / egm96DX
	gy := (latitude - egm96Y0) / egm96DY
	geoid := interpGeoidBicubic(gx, gy)

	return Location{
		latitude:  latitude * Deg,
		longitude: longitude * Deg,
		height:    height + geoid,
	}, nil
}

// Equals returns whether the latitude, longitude and height of the input location
// are equal to those of the caller.
func (l Location) Equals(ll Location) bool {
	return l.latitude==ll.latitude && l.longitude==ll.longitude && l.height==ll.height
}

// Geodetic returns the location's lat (latitude), lng (longitude), and h (height).
// lat and lng are in radians and r is in meters.
// Geodetic coordinates are the variables φ,λ,h in the WMM paper.
func (l Location) Geodetic() (phi, lambda, r float64) {
	return l.latitude, l.longitude, l.height
}

// Spherical returns the location's phi (φ', corresponding to latitude),
// lambda (λ, equal to geodetic longitude), and r (r, distance from center of
// WGS sphere).  phi and lambda are in radians and r is in meters.
// Spherical coordinates are the variables φ',λ,r in the WMM paper.
func (l Location) Spherical() (phi, lambda, r float64) {
	sinPhi := math.Sin(l.latitude)
	cosPhi := math.Cos(l.latitude)
	h := l.height
	rc := A/math.Sqrt(1-E2*sinPhi*sinPhi)
	p := (rc+h)*cosPhi
	z := (rc*(1-E2)+h)*sinPhi
	r = math.Sqrt(p*p+z*z)
	return math.Asin(z/r), l.longitude, r
}

// HeightAboveMSL calculates the height of the EGM96 geoid at the input Location,
// which corresponds to the height of MSL relative to the WGS84 reference ellipsoid.
// It then subtracts this height from the total height above the WGS84 reference
// ellipsoid at the input Location, giving the the height above MSL.
func (l Location) HeightAboveMSL() (h float64, err error) {
	egm96LoadOnce.Do(loadEGM96Grid)

	lng := l.longitude / Deg
	lat := l.latitude / Deg
	nLng := int((lng-egm96X0)/egm96DX) // Grid x just below desired x
	nLat := int((lat-egm96Y0)/egm96DY) // Grid y just below desired y

	// Bounds checks guard the bilinear fallback path used at the latitude
	// poles. The bicubic stencil reads a 4-cell window in latitude and is
	// guarded internally; the latitude/longitude rejection here matches the
	// historical bilinear contract so existing error messages still apply.
	if nLng < 0 || nLng > egm96XN-2 {
		return 0, fmt.Errorf("requested longitude %4.2f lies outside of EGM96 longitude range %4.1f to %4.1f",
			lng, egm96X0, egm96X1)
	}
	if nLat < 0 || nLat > egm96YN-2 {
		return 0, fmt.Errorf("requested latitude %4.2f lies outside of EGM96 latitude range %4.1f to %4.1f",
			lat, egm96Y0, egm96Y1)
	}

	gx := (lng - egm96X0) / egm96DX
	gy := (lat - egm96Y0) / egm96DY
	h = l.height - interpGeoidBicubic(gx, gy)

	return h, err
}

// catmullRom1D evaluates a 1D Catmull-Rom cubic at parameter t∈[0,1]
// on four equally-spaced samples; p1 and p2 bracket the target. The
// curve passes through every (i, p_i) sample and has C¹ continuity
// across cell boundaries.
func catmullRom1D(t, p0, p1, p2, p3 float64) float64 {
	return p1 +
		0.5*t*(p2-p0) +
		t*t*(p0-2.5*p1+2*p2-0.5*p3) +
		t*t*t*(-0.5*p0+1.5*p1-1.5*p2+0.5*p3)
}

// interpGeoidBilinear evaluates the EGM96 grid at fractional grid coordinates
// (gx, gy) using the historical bilinear formula. Used directly by
// interpGeoidBicubic as the polar fallback. Callers must have already
// validated the (gx, gy) range to permit a 2×2 stencil read.
func interpGeoidBilinear(gx, gy float64) float64 {
	nx := int(math.Floor(gx))
	ny := int(math.Floor(gy))
	fx := gx - float64(nx)
	fy := gy - float64(ny)
	row := ny * egm96XN
	h00 := egm96Grid[row+nx]
	h10 := egm96Grid[row+nx+1]
	h01 := egm96Grid[row+egm96XN+nx]
	h11 := egm96Grid[row+egm96XN+nx+1]
	return (1-fx)*(1-fy)*h00 + fx*(1-fy)*h10 + (1-fx)*fy*h01 + fx*fy*h11
}

// interpGeoidBicubic evaluates the EGM96 grid at fractional grid coordinates
// (gx, gy) using a bicubic Catmull-Rom spline (4×4 stencil). Longitude wraps
// modularly across the antimeridian: indices 0 and egm96XN-1 represent the
// same meridian so the (nx + k) mod (egm96XN-1) form picks the correct
// neighbour without a special case. Latitude does not wrap, so within one
// grid cell of either pole (where the stencil would read off the grid) we
// fall back to bilinear; the affected band is ~0.5° around each pole.
//
// Callers must have already validated (gx, gy) for the bilinear stencil
// (HeightAboveMSL / NewLocationMSL guard this); interpGeoidBicubic adds
// no further bounds errors.
func interpGeoidBicubic(gx, gy float64) float64 {
	nx := int(math.Floor(gx))
	ny := int(math.Floor(gy))
	fx := gx - float64(nx)
	fy := gy - float64(ny)

	// Polar fallback: bicubic stencil reads ny-1..ny+2; if either bound
	// steps off the latitude grid (which doesn't wrap), fall back to bilinear.
	if ny < 1 || ny > egm96YN-3 {
		return interpGeoidBilinear(gx, gy)
	}

	mod := egm96XN - 1
	wrap := func(i int) int { return ((i%mod)+mod)%mod }
	i0, i1, i2, i3 := wrap(nx-1), wrap(nx), wrap(nx+1), wrap(nx+2)

	var col [4]float64
	for j := 0; j < 4; j++ {
		row := (ny - 1 + j) * egm96XN
		col[j] = catmullRom1D(fx,
			egm96Grid[row+i0],
			egm96Grid[row+i1],
			egm96Grid[row+i2],
			egm96Grid[row+i3],
		)
	}
	return catmullRom1D(fy, col[0], col[1], col[2], col[3])
}

var (
	egm96X0, egm96X1, egm96DX float64
	egm96Y0, egm96Y1, egm96DY float64
	egm96XN, egm96YN int
	egm96Grid []float64
)

// NearestEGM96GridPoint looks up the grid point nearest the desired location within the
// 15'x15' resolution grid data for the EGM96 geoid model.
//
// The returned Location contains the lat/long of the grid point and the height in meters of
// the geoid relative to the WGS 84 reference ellipsoid at that grid point.
//
// Ignores any height value in the input Location.
func (l Location) NearestEGM96GridPoint() (loc Location, err error) {
	egm96LoadOnce.Do(loadEGM96Grid)

	// l.longitude is already normalized to [0, 360°) at construction.
	lng := l.longitude / Deg
	lat := l.latitude / Deg
	nLng := int((lng-egm96X0)/egm96DX + 0.5)
	nLat := int((lat-egm96Y0)/egm96DY + 0.5)

	// NearestEGM96GridPoint reads only [nLat*egm96XN+nLng], so valid indices
	// are [0, egm96XN-1] / [0, egm96YN-1].
	if nLng < 0 || nLng >= egm96XN {
		return Location{},
			fmt.Errorf("requested longitude %4.2f lies outside of EGM96 longitude range %4.1f to %4.1f",
				lng, egm96X0, egm96X1)
	}
	if nLat < 0 || nLat >= egm96YN {
		return Location{},
			fmt.Errorf("requested latitude %4.2f lies outside of EGM96 latitude range %4.1f to %4.1f",
				lat, egm96Y0, egm96Y1)
	}

	return Location{
		latitude:  (egm96Y0+egm96DY*float64(nLat))*Deg,
		longitude: (egm96X0+egm96DX*float64(nLng))*Deg,
		height:    egm96Grid[nLat*egm96XN+nLng],
	}, nil
}

func loadEGM96Grid() {
	var (
		err error
		dat []string
		v   float64
		i   int
	)

	scanner := bufio.NewScanner(bytes.NewReader(defaultGrid))
	// Read and parse header
	if !scanner.Scan() {
		panic("Could not read header line from EGM96 grid file")
	}
	dat = strings.Fields(scanner.Text())
	if egm96Y1, err = strconv.ParseFloat(dat[0], 64); err != nil {
		panic("bad EGM96 grid file header for Y1")
	}
	if egm96Y0, err = strconv.ParseFloat(dat[1], 64); err != nil {
		panic("bad EGM96 grid file header for Y0")
	}
	if egm96X0, err = strconv.ParseFloat(dat[2], 64); err != nil {
		panic("bad EGM96 grid file header for X0")
	}
	if egm96X1, err = strconv.ParseFloat(dat[3], 64); err != nil {
		panic("bad EGM96 grid file header for X1")
	}
	if egm96DX, err = strconv.ParseFloat(dat[4], 64); err != nil {
		panic("bad EGM96 grid file header for DX")
	}
	if egm96DY, err = strconv.ParseFloat(dat[5], 64); err != nil {
		panic("bad EGM96 grid file header for DY")
	}

	if egm96X1 < egm96X0 {
		egm96DX *= -1
	}
	if egm96Y1 < egm96Y0 {
		egm96DY *= -1
	}
	egm96XN = int((egm96X1-egm96X0)/egm96DX+0.5)+1 // Count the ends
	egm96YN = int((egm96Y1-egm96Y0)/egm96DY+0.5)+1
	egm96Grid = make([]float64, egm96XN*egm96YN)

	// Read and parse data
	i = 0
	for scanner.Scan() {
		for _, s := range strings.Fields(scanner.Text()) {
			if v, err = strconv.ParseFloat(s, 64); err != nil {
				panic("bad EGM96 grid data")
			}
			egm96Grid[i] = v
			i++
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
