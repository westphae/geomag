# geomag
geomag is an implementation in Go of the NOAA World Magnetic Model.

The World Magnetic Model home is at https://www.ngdc.noaa.gov/geomag/WMM/DoDWMM.shtml.

The coefficients for 2025-2029 can be downloaded at https://www.ncei.noaa.gov/products/world-magnetic-model

## Install

Requires Go 1.22 or newer. The current WMM coefficients (WMM2025) are embedded
in the package, so there is no separate data file to download.

To install the `wmm_point` command-line tool:

```sh
go install github.com/westphae/geomag/cmd/wmm_point@latest
```

The binary lands in `$(go env GOBIN)` (or `$(go env GOPATH)/bin` if `GOBIN` is
unset) — make sure that directory is on your `PATH`. Pin to a specific release
by replacing `@latest` with `@v1.2025.0`, etc.

To use the library in your own Go project:

```sh
go get github.com/westphae/geomag@latest
```

then `import "github.com/westphae/geomag/pkg/wmm"` and/or
`"github.com/westphae/geomag/pkg/egm96"`. See the per-package examples
below.

## Commands
geomag provides three command line programs, modeled after the command line programs in the official NOAA software.

`wmm_point` calculates magnetic field values for a single location and time:
```
> wmm_point N30 W88.51 0.01 2019.5

Results For

Latitude:       30.00N
Longitude:      88.51W
Altitude:        0.010 kilometers above mean sea level
Date:           2019.5

       Main Field             Secular Change
       F    =  46944.3 nT ± 152.0 nT  -118.8 nT/yr
       H    =  24074.6 nT ± 133.0 nT    -6.8 nT/yr
       X    =  24060.2 nT ± 138.0 nT    -8.0 nT/yr
       Y    =   -831.0 nT ±  89.0 nT   -36.3 nT/yr
       Z    =  40301.2 nT ± 165.0 nT  -134.3 nT/yr
       Decl =     -1º 59' ± 19'         -5.2'/yr
       Incl =     59º  9' ± 13'         -4.6'/yr

       Grid Variation =  -1º 59'
```

`wmm_file` batch-processes a coordinate file (one record per line) into a
matching output file. Input format mirrors NOAA's `wmm_file` reference:

```
> cat coords.txt
2025.5 E K100  70.3 30.8
2026.5 E M1042 70.3 30.8
2027.5 E F30000 70.3 30.8
> wmm_file f coords.txt out.txt
> head -2 out.txt
Date Coord-System Altitude Latitude Longitude D_deg D_min I_deg I_min H_nT X_nT Y_nT Z_nT F_nT dD_min dI_min dH_nT dX_nT dY_nT dZ_nT dF_nT
2025.5 E K100 70.3 30.8    16d 48m    79d 13m    9827.3   9407.8   2840.6  51603.9  52531.3    15.0       1.4        -11.5    -23.4     37.7     51.1     48.1
```

Each input line is `<date> <coord-system> <altitude> <lat> <lng>` where
`<coord-system>` is `E` (height above WGS84 ellipsoid) or `M` (height
above mean sea level), and `<altitude>` is a single-character unit prefix
(`K`=km, `M`=meters, `F`=feet) followed by a signed decimal magnitude.
Add `e` (or `--Errors`) between the `f` and the input filename to append
seven uncertainty columns.

`wmm_grid` generates a 4-D grid of magnetic field values over latitude,
longitude, altitude, and time, evaluating one chosen output element at
each grid point:

```
> wmm_grid --lat=-80,80,40 --lng=0,0,0 --alt=0,0,0 --date=2026,2026,0 --element=F --errors
-80 0 0 2026 46061.398... 138
-40 0 0 2026 23369.279... 138
0 0 0 2026 31813.131... 138
40 0 0 2026 45193.628... 138
80 0 0 2026 55208.584... 138
```

Each axis flag takes a comma-separated `START,END,STEP` triple. Setting
START=END collapses the axis (e.g. `--alt=0,0,0` for sea level only,
`--date=2026,2026,0` for a single instant). `--element` accepts D, I, F,
H, X, Y, Z, dD, dI, dF, dH, dX, dY, dZ. `--errors` adds an uncertainty
column (only for the static-field elements). Output goes to stdout
unless you pass `--output=PATH`.

## Packages
Two packages are provided by this library:

### egm96
Package egm96 provides a representation of the 1996 Earth Gravitational Model (EGM96),
a geopotential model of the Earth.

EGM96 is the geoid reference model component of the World Geodetic System (WGS84).
It consists of n=m=360 spherical harmonic coefficients as published by the
National Geospatial-Intelligence Agency (NGA).  The NGA also publishes a raster grid
of the calculated heights which can be interpolated to approximate the geoid height
at any location.

In effect, this model provides the height of sea level above the WGS84 reference ellipsoid.
It is used, for example, in GPS navigation to provide the height above sea level.

This package is based on the NGA-provided 15'x15' resolution grid encoding
the heights of the geopotential surface at each lat/long, and interpolates between grid
points using a bilinear interpolation.

usage:
```
import "github.com/westphae/geomag/pkg/egm96"

// Calculate height above MSL for a point at a
// latitude of 12.25 South, longitude of 82.75 East, and
// altitude of 1000m above the WGS84 ellipsoid (i.e. GPS altitude)
h, err := egm96.NewLocationGeodetic(-12.25, 82.75, 1000).HeightAboveMSL()
```

### wmm
Package wmm provides a representation of the 2025 World Magnetic Model (WMM),
a mathematical model of the magnetic field produced by the Earth's core and
its variation over time.

WMM is the magnetic model component of the World Geodetic System (WGS84).
It consists of n=m=12 spherical harmonic coefficients as published by the
National Geospatial-Intelligence Agency (NGA).

This model evaluates all magnetic field components and their rates of change
for any location on the Earth's surface.  These field components include the
X, Y, and Z values and the declination D and inclination I.
The Declination is used, for example, in correcting a Magnetic Heading to a
True Heading.

usage:
```
import "github.com/westphae/geomag/pkg/egm96"
import "github.com/westphae/geomag/pkg/wmm"

tt := wmm.DecimalYear(2019.5)
loc := egm96.NewLocationGeodetic(-12.25, 82.75, 1000)

mag, _ := wmm.CalculateWMMMagneticField(loc, tt.ToTime())
fmt.Printf("Declination at your location: %2.2f\n", mag.D())
```

## Validation
The library code is fully tested.
In particular, all test values provided with the official NOAA WMM are tested here,
as well as the detailed example in the WMM technical paper.
Please submit an issue on github if you notice any other issues.

## Updating Coefficients
Coefficients are currently updated through the WMM2025 model. To bump to a
newer release, drop the new `WMM.COF` into `pkg/wmm/embedded/` (replacing the
existing one) and rebuild — the file is embedded at compile time via
`//go:embed`. Add the historical version to `pkg/wmm/testdata/` if you want
to keep its tests alongside the new ones.

## License Info
This software is based on the NOAA World Magnetic Model.
The source code in this project is not based on the source code provided by NOAA, but on the
equations provided in the World Magnetic Model reference paper.

The WMM source code is not subject to copyright protection: https://www.ngdc.noaa.gov/geomag/WMM/license.shtml

The WMM source code is in the public domain and not licensed or under copyright. The information and software may be used freely by the public. As required by 17 U.S.C. 403, third parties producing copyrighted works consisting predominantly of the material produced by U.S. government agencies must provide notice with such work(s) identifying the U.S. Government material incorporated and stating that such material is not subject to copyright protection.
