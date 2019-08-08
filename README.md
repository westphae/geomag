# geomag
Golang implementation of the NOAA World Magnetic Model

The World Magnetic Model home is at https://www.ngdc.noaa.gov/geomag/WMM/DoDWMM.shtml.
The coefficients for 2015-2020 can be downloaded at https://www.ngdc.noaa.gov/geomag/WMM/data/WMM2015/WMM2015v2COF.zip

## TODO
1. Write function to calculate Legendre Polynomials and validate. DONE
2. Write function to calculate Associated Legendre Functions and validate. DONE
3. Write function to calculate height above WGS84 ellipsoid from MSL height. DONE
4. Write function to calculate spherical geocentric coordinates from geodetic coordinates. DONE
5. Write function to read in COF file. DONE
6. Write function to calculate Gauss coefficients at time t from model coefficients. DONE
7. Write function to calculate the magnetic field components X,Y,Z. DONE
8. Write function to calculate the derivatives of X,Y,Z. DONE
9. Write function to rotate geocentric values X,Y,Z into ellipsoidal reference frame. DONE
10. Calculate projected components H,F,I,D and their derivatives. DONE
11. Allow WMM coefficients file to be used instead of bindata, use for tests. DONE
11. Write test module to test against WMM test values. DONE
12. Handle grivation near poles. DONE
13. Complete documentation.
14. Refactor to fully handle iterations over time, height, longitude, latitude.
15. Write a command line utility to calculate values for a given location/time.
16. Write a command line utility to calculate values for a range of locations/times.

## Notes
* Legendre function coefficients should be calculated once when first called and then cached.
* WMM coefficients read into 4 [][]float64 slices g,h,gg,hh, triangular in shape

## License Info
This software is based on the NOAA World Magnetic Model.
The source code in this project is not based on the source code provided by NOAA, but on the
equations provided in the World Magnetic Model reference paper.

The WMM source code is not subject to copyright protection: https://www.ngdc.noaa.gov/geomag/WMM/license.shtml

The WMM source code is in the public domain and not licensed or under copyright. The information and software may be used freely by the public. As required by 17 U.S.C. 403, third parties producing copyrighted works consisting predominantly of the material produced by U.S. government agencies must provide notice with such work(s) identifying the U.S. Government material incorporated and stating that such material is not subject to copyright protection.
