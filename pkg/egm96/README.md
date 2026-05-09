# EGM96

This package egm96 provides a Go representation of the EGM96 geopotential model of the Earth.
It calculates the geoid height of the 1996 Earth Gravitational Model (EGM96) for a given latitude and longitude.

## The EGM96 Model
The EGM96 model is a component of the 1984 World Geodetic System (WGS84).
The EGM96 homepage is at https://cddis.nasa.gov/926/egm96/egm96.html.

WGS84 defines a datum surface which is an ellipsoid whose center coincides with the Earth's center of mass.
EGM96 defines a "geoid," a gravitational equipotential surface, relative to this datum surface.
As an equipotential surface, the geoid also corresponds to Mean Sea Level.

EGM96 is specified as a spherical harmonics series of degree 360.
The National Geospatial-Intelligence Agency (NGA), which is responsible for the model,
also publishes a grid of the computed geoid heights at a 15'x15' resolution, from which
the geoid height at any location can be interpolated.

This package calculates the geoid height at any location via interpolation of the NGA grid.
Currently, a bilinear interpolation is used.

## Install

```sh
go get github.com/westphae/geomag@latest
```

Requires Go 1.22 or newer. The 15'x15' NGA grid is embedded in the package —
no separate data file to download. See the top-level
[README](../../README.md) for the full install story.

## Usage
The most common usage is to create a location corresponding to a GPS-derived
latitude, longitude, and height-above-ellipsoid, and then calculate the
height above MSL:

```go
import "github.com/westphae/geomag/pkg/egm96"

loc := egm96.NewLocationGeodetic(-12.25, 82.75, 1000) // lat, lng (deg), height (m above ellipsoid)
h, err := loc.HeightAboveMSL()
```

Negative longitudes (the GPS/WGS84 `[-180, 180]` convention) are accepted —
the constructor wraps to `[0, 360)` internally.

## Testing and Validation
The heights produced by this program have been validated against online calculator at
https://www.unavco.org/software/geodetic-utilities/geoid-height-calculator/geoid-height-calculator.html

### Copyright
The EGM96 model and associated data files are produced by the US Government and are not subject to copyright.
The software in this package is provided under the MIT license where applicable.