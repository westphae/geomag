# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Module layout

Standard Go module — `go.mod` declares `module github.com/westphae/geomag` at `go 1.24`. No external dependencies. Embedded data files live under `pkg/wmm/embedded/WMM.COF` and `pkg/egm96/embedded/ww15mgh.grd` and are pulled in via `//go:embed`.

## Common commands

```sh
go test ./...                                            # run all tests
go test -race ./...                                      # with race detector
go test ./pkg/wmm -run TestAll2025TestValuesFromPaper    # single test
go build ./cmd/wmm_point                                 # build CLI
go run ./cmd/wmm_point N30 W88.51 0.01 2019.5            # one-shot run
go run ./cmd/wmm_point N89 W121 E28 2025.0               # WMM2025 row-1 fixture
go run ./cmd/wmm_point                                   # interactive prompts
```

`wmm_point` accepts `--cof_file=<path>` to load an alternate WMM coefficients file and `--spherical` to print spherical (rather than ellipsoidal) field components. Lat/lng accept DMS triples (`30,30,30`) or decimal degrees; altitude is in km with optional `E` prefix for height-above-ellipsoid (default is height above MSL); date is decimal year or `MM/DD/YYYY`.

## Updating WMM coefficients

Drop the new `WMM.COF` into `pkg/wmm/embedded/` (replacing the existing one) and rebuild — that's it. The file is embedded at compile time via `//go:embed`. If you also want to keep the old version's tests, copy the previous COF and the corresponding `*_TEST_VALUES.txt` into `pkg/wmm/testdata/` before swapping. The header parser keys on whitespace tokens (`strings.Fields`), so minor whitespace differences in the NOAA-published file are tolerated.

## Architecture

Three library packages plus one CLI:

- **`pkg/egm96`** — EGM96 geoid model. Owns the shared `Location` type used by every other package. A `Location` stores geodetic (φ, λ, h) values and exposes both `Geodetic()` (lat/lng on the WGS84 ellipsoid) and `Spherical()` (φ′, λ, r — geocentric, used by the WMM math). `HeightAboveMSL()` and `NewLocationMSL` bilinearly interpolate the embedded 15′×15′ NGA grid.

- **`pkg/polynomial`** — Legendre polynomial machinery. `LegendreFunction(n, m, x)` is the hot path: it differentiates `LegendrePolynomial(n)` `m` times and caches the result in a process-wide `sync.Map` keyed by `(n, m)`. The Schmidt semi-normalization (`sqrt(2 · (n−m)! / (n+m)!)` for m>0) is applied by the **caller** in `pkg/wmm`, not here.

- **`pkg/wmm`** — WMM evaluation. The primary type is `Model`, an immutable-after-load value (sync.RWMutex-guarded) holding the parsed coefficients plus a per-location cache. Construct via `NewModel()` (embedded default), `LoadModel(path)`, or `ParseModel(io.Reader)`. Methods: `Coefficients(n, m, t)`, `MagneticField(loc, t)`. The package also exposes a thread-safe package-level default model accessible via `Default()`, plus thin backward-compat wrappers (`LoadWMMCOF`, `GetWMMCoefficients`, `CalculateWMMMagneticField`) and the legacy `Epoch`/`COFName`/`ValidDate` vars that mirror the default model's state. Per the cache design, callers iterating over many points should make **time the innermost loop**, then height/lat/lng, to maximize hits in the per-location cache (the cache stores the spherical-harmonic sum evaluated at `validDate`, then the result is linearly time-corrected on read).

- **`cmd/wmm_point`** — Mirrors NOAA's `wmm_point` CLI. Interactive when run with no positional args, otherwise expects exactly four (lat, lng, alt, date). Input parsing lives in `internal/parsing/parsing.go`.

### Coordinate conventions

The WMM math runs in **spherical** geocentric coordinates; results are rotated to **ellipsoidal** axes via `MagneticField.Ellipsoidal()`, which applies a rotation by `(spherical_lat − geodetic_lat)`. `Spherical()` returns the raw values. Most navigation use cases want `Ellipsoidal()` (or the derived `H()`, `D()`, `I()`, `GV()`).

## Validation

Test fixtures in `pkg/wmm/testdata/` are the upstream NOAA WMM test vectors (WMM2015v1, WMM2015v2, WMM2020, WMM2025). The `TestAll20XXTestValuesFromPaper` tests in `magnetic_field_test.go` exercise each set; the WMM2015 worked example from section 1.5 of the WMM technical report is also covered. Any change to the field math or coefficient parsing should keep these tests passing to within ~0.05 nT / 0.005°.
