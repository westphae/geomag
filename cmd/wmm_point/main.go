// wmm_point estimates the strength and direction of Earth's main Magnetic field for a given point/area.
//
// Usage is
//  wmm_point --cof_file=WMM2020.COF --spherical [latitude] [longitude] [altitude] [date]
//
// The World Magnetic Model (WMM) for 2020
// is a model of Earth's main Magnetic field.  The WMM
// is recomputed every five (5) years, in years divisible by
// five (i.e. 2010, 2015, 2020).
//
// Information on the model is available at https://www.ngdc.noaa.gov/geomag/WMM/DoDWMM.shtml
//
// Input required is the location in geodetic latitude and
// longitude (positive for northern latitudes and eastern
// longitudes), geodetic altitude in meters, and the date of
// interest in years.
//
// The program computes the estimated Magnetic Declination
// (Decl) which is sometimes called MagneticVAR, Inclination (Incl), Total
// Intensity (F or TI), Horizontal Intensity (H or HI), Vertical
// Intensity (Z), and Grid Variation (GV). Declination and Grid
// Variation are measured in units of degrees and are considered
// positive when east or north.  Inclination is measured in units
// of degrees and is considered positive when pointing down (into
// the Earth).  The WMM is referenced to the WGS-84 ellipsoid and
// is valid for 5 years after the base epoch. Uncertainties for the
// WMM are one standard deviation uncertainties averaged over the globe.
// We represent the uncertainty as constant values in Incl, F, H, X,
// Y, and Z. Uncertainty in Declination varies depending on the strength
// of the horizontal field.  For more information see the WMM Technical
// Report.
//
// It is very important to note that a  degree and  order 12 model,
// such as WMM, describes only the long  wavelength spatial Magnetic
// fluctuations due to  Earth's core.  Not included in the WMM series
// models are intermediate and short wavelength spatial fluctuations
// that originate in Earth's mantle and crust. Consequently, isolated
// angular errors at various  positions on the surface (primarily over
// land, along continental margins and  over oceanic sea-mounts, ridges and
// trenches) of several degrees may be expected.  Also not included in
// the model are temporal fluctuations of magnetospheric and ionospheric
// origin. On the days during and immediately following Magnetic storms,
// temporal fluctuations can cause substantial deviations of the Geomagnetic
// field  from model  values.  If the required  declination accuracy  is
// more stringent than the WMM  series of models provide, the user is
// advised to request special (regional or local) surveys be performed
// and models prepared. The World Magnetic Model is a joint product of
// the United States’ National Geospatial-Intelligence Agency (NGA) and
// the United Kingdom’s Defence Geographic Centre (DGC). The WMM was
// developed jointly by the National Centers for Environmental Information (NCEI, Boulder
// CO, USA) and the British Geological Survey (BGS, Edinburgh, Scotland).
//
// Sample output:
//  Results For
//  
//  Latitude:       30.00N
//  Longitude:      88.51W
//  Altitude:        0.010 kilometers above mean sea level
//  Date:           2019.5
//  
//         Main Field             Secular Change
//         F    =  46944.3 nT ± 152.0 nT  -118.8 nT/yr
//         H    =  24074.6 nT ± 133.0 nT    -6.8 nT/yr
//         X    =  24060.2 nT ± 138.0 nT    -8.0 nT/yr
//         Y    =   -831.0 nT ±  89.0 nT   -36.3 nT/yr
//         Z    =  40301.2 nT ± 165.0 nT  -134.3 nT/yr
//         Decl =     -1º 59' ± 19'         -5.2'/yr
//         Incl =     59º  9' ± 13'         -4.6'/yr
//  
//         Grid Variation =  -1º 59'
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/westphae/geomag/internal/util"
	"github.com/westphae/geomag/pkg/egm96"
	"github.com/westphae/geomag/pkg/wmm"
)

const (
	usage = "wmm_point --cof_file=WMM2020.COF --spherical [latitude] [longitude] [altitude] [date]"
	cofUsage = "COF coefficients file to use, empty for the built-in one"
	sphericalUsage = "Output spherical values instead of ellipsoidal"
	lngErr = "Error: Degree input is outside legal range. The legal range is from -180 to 360."
	fieldWarn = "Warning: The Horizontal Field strength at this location is only 0.000000. " +
		"Compass readings have VERY LARGE uncertainties in areas where where H is smaller than 1000 nT"
)

var prompt = map[string]string{
	"latitude": "Please enter latitude North Latitude positive. " +
		"For example: 30, 30, 30 (D,M,S) or 30.508 (Decimal Degrees) (both are north). ",
	"longitude": "Please enter longitude East longitude positive, West negative. " +
		"For example: -100.5 or -100, 30, 0 for 100.5 degrees west. ",
	"altitude": "Please enter height above mean sea level (in kilometers). " +
		"[For height above WGS-84 Ellipsoid prefix E, for example (E20.1)]. ",
	"date": "Please enter the decimal year or calendar date (YYYY.yyy, MM DD YYYY or MM/DD/YYYY) ",
}

var (
	cofFile    string
	spherical  bool
	latitude   float64
	longitude  float64
	altitude   float64
	hae        bool
	dYear      float64
	ErrHelp    error
	err        error
	loc        egm96.Location
	x, y, z    float64
	dx, dy, dz float64
)

func init() {
	flag.StringVar(&cofFile, "cof_file", "", cofUsage)
	flag.StringVar(&cofFile, "c", "", cofUsage)

	flag.BoolVar(&spherical, "spherical", false, sphericalUsage)
	flag.BoolVar(&spherical, "s", false, sphericalUsage)

	ErrHelp = errors.New(usage)
}

func main() {
	flag.Parse()

	if cofFile!="" {
		if err = wmm.LoadWMMCOF(cofFile); err != nil {
			fmt.Println(err)
			return
		}
	}
	fmt.Printf("COF File: %v, Epoch: %v, Valid Date: %d/%d/%d\n", wmm.COFName, wmm.Epoch,
		wmm.ValidDate.Month(), wmm.ValidDate.Day(), wmm.ValidDate.Year())

	if flag.NArg() == 0 {
		userInput()
	} else if flag.NArg() == 4 {
		if latitude, err = parsing.ParseLatLng(flag.Arg(0)); err!=nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}
		if longitude, err = parsing.ParseLatLng(flag.Arg(1)); err!=nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}
		if altitude, hae, err = parsing.ParseAltitude(flag.Arg(2)); err!=nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}
		if dYear, err = parsing.ParseTime(flag.Arg(3)); err!=nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "You must specify a latitude, longitude, altitude and date in that order")
		return
	}
	for longitude < 0 {
		longitude += 360
	}
	if longitude >= 360 {
		_, _ = fmt.Fprintln(os.Stderr, lngErr)
	}
	altitude *= 1000 // Convert to meters

	if hae {
		loc = egm96.NewLocationGeodetic(latitude, longitude, altitude)
	} else {
		loc, err = egm96.NewLocationMSL(latitude, longitude, altitude)
		if err != nil {
			fmt.Printf("Error making location: %s\n", err)
		}
	}
	mf, err := wmm.CalculateWMMMagneticField(
		loc,
		wmm.DecimalYear(dYear).ToTime(),
		)

	fmt.Println("Results For")
	fmt.Println()
	lat, lng, hh := loc.Geodetic()
	qualifier := "N"
	quantity := lat/egm96.Deg
	if quantity<0 {
		qualifier = "S"
		quantity = -quantity
	}
	fmt.Printf("Latitude:\t%4.2f%s\n", quantity, qualifier)

	qualifier = "E"
	quantity = lng/egm96.Deg
	if quantity>=180 {
		qualifier = "W"
		quantity = 360-quantity
	}
	fmt.Printf("Longitude:\t%4.2f%s\n", quantity, qualifier)

	relationship := "above"
	quantity = hh
	qualifier = "the WGS-84 ellipsoid"
	if !hae {
		quantity, _ = loc.HeightAboveMSL()
		qualifier = "mean sea level"
	}
	if quantity<0 {
		relationship = "below"
		quantity = -quantity
	}
	fmt.Printf("Altitude:\t%6.3f kilometers %s %s\n", quantity/1000, relationship, qualifier)

	fmt.Printf("Date:\t\t%5.1f\n", dYear)

	qualifier = ""
	if spherical {
		qualifier = "(Spherical)"
	}
	fmt.Println()

	if err != nil {
		fmt.Printf("Warning: %s\n\n", err)
	}

	if spherical {
		x, y, z, dx, dy, dz = mf.Spherical()
	} else {
		x, y, z, dx, dy, dz = mf.Ellipsoidal()
	}

	dD, dM, dS := egm96.DegreesToDMS(mf.D())
	iD, iM, iS := egm96.DegreesToDMS(mf.I())
	gvD, gvM, gvS := egm96.DegreesToDMS(mf.GV(loc))
	fmt.Println("       Main Field             Secular Change")
	fmt.Printf("F    = %8.1f nT ± %5.1f nT  %6.1f nT/yr\n", mf.F(), mf.ErrF(), mf.DF())
	if !spherical {
		fmt.Printf("H    = %8.1f nT ± %5.1f nT  %6.1f nT/yr\n", mf.H(), mf.ErrH(), mf.DH())
	}
	fmt.Printf("X    = %8.1f nT ± %5.1f nT  %6.1f nT/yr %s\n", x, mf.ErrX(), dx, qualifier)
	fmt.Printf("Y    = %8.1f nT ± %5.1f nT  %6.1f nT/yr %s\n", y, mf.ErrY(), dy, qualifier)
	fmt.Printf("Z    = %8.1f nT ± %5.1f nT  %6.1f nT/yr %s\n", z, mf.ErrZ(), dz, qualifier)
	if !spherical {
		fmt.Printf("Decl =    %3.0fº %2.0f' ± %2.0f'         %4.1f'/yr\n", dD, dM+dS/60, mf.ErrD()*60, mf.DD()*60)
		fmt.Printf("Incl =    %3.0fº %2.0f' ± %2.0f'         %4.1f'/yr\n", iD, iM+iS/60, mf.ErrI()*60, mf.DI()*60)
		fmt.Println()
		fmt.Printf("Grid Variation =  %2.0fº %2.0f'\n", gvD, gvM+gvS/60)
	}
}

func userInput() {
	var (
		input string
		err   error
	)

	err = fmt.Errorf("")
	for err!=nil {
		input = readUserInput(prompt["latitude"])
		if input == "q" {
			fmt.Println("Goodbye")
			os.Exit(1)
		}
		latitude, err = parsing.ParseLatLng(input)
		if err!=nil {
			fmt.Println(err)
		}
	}

	err = fmt.Errorf("")
	for err!=nil {
		input = readUserInput(prompt["longitude"])
		if input == "q" {
			fmt.Println("Goodbye")
			os.Exit(1)
		}
		longitude, err = parsing.ParseLatLng(input)
		if err!=nil {
			fmt.Println(err)
		}
	}

	err = fmt.Errorf("")
	for err!=nil {
		input = readUserInput(prompt["altitude"])
		if input == "q" {
			fmt.Println("Goodbye")
			os.Exit(1)
		}
		altitude, hae, err = parsing.ParseAltitude(input)
		if err!=nil {
			fmt.Println(err)
		}
	}

	err = fmt.Errorf("")
	for err!=nil {
		input = readUserInput(prompt["date"])
		if input == "q" {
			fmt.Println("Goodbye")
			os.Exit(1)
		}
		dYear, err = parsing.ParseTime(input)
		if err!=nil {
			fmt.Println(err)
		}
	}

}

func readUserInput(prompt string) (inp string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	inp, _ = reader.ReadString('\n')
	inp = strings.TrimSpace(inp)
	return inp
}
