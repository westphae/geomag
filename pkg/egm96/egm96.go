package egm96

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/westphae/geomag/pkg/units"
)

const EPS = 1e-6

var (
	egm96X0, egm96X1, egm96DX float64
	egm96Y0, egm96Y1, egm96DY float64
	egm96XN, egm96YN int
	egm96Grid []float64
)

// GetEGM96GridPoint looks up the grid point nearest the desired location
// within the grid data for the EGM96 geoid grid model.
// Ignores any Height value in the passed Location.
func GetEGM96GridPoint(loc units.Location) (err error, nloc units.Location) {
	if len(egm96Grid)==0 {
		loadEGM96Grid()
	}

	lng := float64(loc.Longitude)
	lat := float64(loc.Latitude)
	nLng := int((lng-egm96X0)/egm96DX+0.5)
	nLat := int((lat-egm96Y0)/egm96DY+0.5)

	if nLng < 0 || nLng > egm96XN {
		return fmt.Errorf("requested longitude %4.2f lies outside of EGM96 longitude range %4.1f to %4.1f",
			lng, egm96X0, egm96X1), units.Location{}
	}
	if nLat < 0 || nLat > egm96YN {
		return fmt.Errorf("requested latitude %4.2f lies outside of EGM96 latitude range %4.1f to %4.1f",
			lat, egm96Y0, egm96Y1), units.Location{}
	}

	return nil, units.Location{
		Latitude:  units.Degrees(egm96Y0+egm96DY*float64(nLat)),
		Longitude: units.Degrees(egm96X0+egm96DX*float64(nLng)),
		Height:    units.Meters(egm96Grid[nLat*egm96XN+nLng]),
	}
}

func ConvertMSLToHeightAboveWGS84(loc units.Location) (err error, h units.Meters) {
	if len(egm96Grid)==0 {
		loadEGM96Grid()
	}

	lng := float64(loc.Longitude)
	lat := float64(loc.Latitude)
	nLng := int((lng-egm96X0)/egm96DX) // Grid x just below desired x
	nLat := int((lat-egm96Y0)/egm96DY) // Grid y just below desired y

	if nLng < 0 || nLng > egm96XN {
		return fmt.Errorf("requested longitude %4.2f lies outside of EGM96 longitude range %4.1f to %4.1f",
			lng, egm96X0, egm96X1), 0
	}
	if nLat < 0 || nLat > egm96YN {
		return fmt.Errorf("requested latitude %4.2f lies outside of EGM96 latitude range %4.1f to %4.1f",
			lat, egm96Y0, egm96Y1), 0
	}

	x := (lng-egm96X0)/egm96DX-float64(nLng)
	y := (lat-egm96Y0)/egm96DY-float64(nLat)
	h00 := egm96Grid[nLat*egm96XN+nLng]
	h10 := egm96Grid[nLat*egm96XN+nLng+1]
	h01 := egm96Grid[(nLat+1)*egm96XN+nLng]
	h11 := egm96Grid[(nLat+1)*egm96XN+nLng+1]

	h = units.Meters((1-x)*(1-y)*h00 + x*(1-y)*h10 + (1-x)*y*h01 + x*y*h11) + loc.Height

	return err, h
}

func loadEGM96Grid() {
	data, err := Asset("ww15mgh.grd")
	if err != nil {
		panic(err)
	}

	var (
		dat []string
		v   float64
		i   int
	)

	scanner := bufio.NewScanner(bytes.NewReader(data))
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