# geomag
Golang implementation of the NOAA World Magnetic Model

The World Magnetic Model home is at https://www.ngdc.noaa.gov/geomag/WMM/DoDWMM.shtml.

## TODO
1. Write function to calculate Legendre Polynomials and validate.
2. Write function to calculate Associated Legendre Functions and validate.
3. Write function to calculate height above WGS84 ellipsoid from MSL height.
4. Write function to calculate spherical geocentric coordinates from geodetic coordinates.
5. Write function to read in COF file
6. Write function to calculate Gauss coefficients at time t from model coefficients.
7. Write function to calculate the magnetic field components X,Y,Z.
8. Write function to calculate the derivatives of X,Y,Z.
9. Write function to rotate geocentric values X,Y,Z into ellipsoidal reference frame.
10. Calculate projected components H,F,I,D and their derivatives.
11. Write test module to test against WMM test values.
12. Handle situation near poles.

## Notes
* Legendre function coefficients should be calculated once when first called and then cached.
* WMM coefficients read into 4 [][]float64 slices g,h,gg,hh, triangular in shape

## Structure

### Types
* Polynomial{c}
  * method to calculate polynomial value as function of x
  * method to calculate polynomial derivative
  * LegendrePolynomial(n) function to return a Polynomial corresponding to Legendre Polynomials
* GeodeticLocation{Lambda,Phi,H,T}
  * method to convert to spherical geocentric coordinates
  * method to convert T (type Time) to fractional years
* GeoMag{X,Y,Z,H,F,I,D,GV}
