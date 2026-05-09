# wmmhr

Package wmmhr provides the WMMHR (World Magnetic Model High Resolution)
coefficients as a sibling to the standard WMM model in
[`github.com/westphae/geomag/pkg/wmm`](../).

## What WMMHR is

Where standard WMM models the geomagnetic field as a spherical-harmonic
series of degree 12 (~91 coefficients), WMMHR extends that to **degree 133**
(~9,045 coefficients). The extra degrees capture finer-scale variations in
the field — particularly improvements in declination accuracy near regional
crustal anomalies.

NOAA's WMMHR2025 release is valid from 2025.0 through 2030.0 (same window
as the standard WMM2025).

## When to use it

For most navigation use cases, standard WMM is fine and a lot faster. Reach
for WMMHR when:

- You're computing declination near a known magnetic anomaly (e.g. parts of
  Eastern Canada, Siberia, or near oceanic ridges) where the local crustal
  field meaningfully diverges from the long-wavelength model.
- You're producing a high-resolution map and want the model to capture
  features at scales below ~3000 km.
- Published WMMHR uncertainty is tighter (e.g. F: ±134 nT vs WMM's ±138 nT).

## Cost

Importing this package adds the embedded `WMMHR.COF` (≈530 KB) to your
binary. Library users who only need standard WMM should not import this
package; they'll see no HR cost.

A single magnetic-field evaluation at degree 133 runs ~85× more inner-loop
iterations than at degree 12. Per-call latency goes from microseconds to
fractions-of-a-millisecond. The per-location cache on `*Model` (built into
the parent `pkg/wmm` package) absorbs that cost completely for time-sweep
loops at a fixed location.

## Usage

```go
import (
    "time"
    "github.com/westphae/geomag/pkg/egm96"
    "github.com/westphae/geomag/pkg/wmm/wmmhr"
)

loc := egm96.NewLocationGeodetic(89, -121, 28*1000)
field, err := wmmhr.Default().MagneticField(loc, time.Now())
decl := field.D() // declination in degrees, computed at HR resolution
```

`wmmhr.Default()` returns a process-wide shared `*wmm.Model` with HR data,
lazy-initialized on first call. `wmmhr.New()` returns a fresh `*wmm.Model`
allocated per-call (use this if you want an independent per-location cache,
e.g. for goroutine-isolated work).

## Numerics

WMMHR's degree 133 strains the polynomial-coefficient approach (alternating
signs, magnitudes up to 10⁸⁰) used historically in `pkg/polynomial`'s
`LegendreFunction`. The geomagnetic math uses a numerically stable
recurrence (`polynomial.SchmidtNormalizedALFTable`) ported from NOAA's
reference C library — the Holmes & Featherstone 2002 algorithm with an
underflow-protective scale factor. Both WMM and WMMHR now go through the
same path, so existing WMM behavior is bit-equivalent within rounding.

## Source data

`embedded/WMMHR.COF` is verbatim NOAA's WMMHR2025 coefficient file from
the official distribution at
<https://www.ncei.noaa.gov/products/world-magnetic-model-high-resolution>.
sha256 of `embedded/WMMHR.COF`:
`8851d40e57a1d948cb56d49b837612844890a941f93a73846a122b6c1182d504`.
