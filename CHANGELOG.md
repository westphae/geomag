# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
This project uses a year-based versioning scheme on top of [SemVer](https://semver.org/):
**MINOR** encodes the WMM model year (`v1.2025.x` for releases that embed
WMM2025, `v1.2030.x` for the next NOAA release, …); **PATCH** increments
within a model era for data reissues, code fixes, or improvements.

## [v1.2025.5] — 2026-05-09

Accuracy release for `pkg/egm96` — geoid interpolation switches from
bilinear to a bicubic Catmull-Rom spline. Strict accuracy improvement
with no API change. Affects every caller that converts between MSL and
ellipsoidal altitude (`NewLocationMSL`, `(Location).HeightAboveMSL`).

### Changed

- **Bicubic Catmull-Rom geoid interpolation** in `pkg/egm96/egm96.go`.
  Replaces the inline bilinear formulas in `HeightAboveMSL` and
  `NewLocationMSL` with three new private helpers — `catmullRom1D`,
  `interpGeoidBilinear`, `interpGeoidBicubic`. Closes the long-standing
  TODO at `pkg/egm96/egm96.go:184`.

  Measured on the nine UNAVCO-validated reference points already
  exercised by `TestEGM96GridInterpolationAgainstKnown`:

  | Metric | Bilinear (before) | Bicubic Catmull-Rom (after) |
  |---|---|---|
  | Max absolute error vs UNAVCO | 0.0557 m | **0.0108 m** (5.1× better) |
  | Per-call compute (M1 Pro) | 2.79 ns | 14.06 ns |
  | Persistent state | none | **none** (no precompute) |
  | Stencil size | 2×2 (4 grid reads) | 4×4 (16 grid reads) |

  The 4×4 stencil is local — no precomputed coefficient tables, no
  extra RAM at load. Geoid evaluation is only invoked when the user
  passes MSL altitude; ellipsoid-altitude callers see no change. The
  +11 ns per evaluation is well below the noise floor of any
  spherical-harmonic computation (microseconds for WMM, milliseconds
  for WMMHR).

  **Edge cases:**
  - **Longitude wraparound** at the antimeridian uses a modular
    `(nx + k) mod (egm96XN-1)` stencil-index transform; indices 0 and
    `egm96XN-1` represent the same meridian, so the bicubic
    interpolation is exact across the seam without a special case.
    Covered by new `TestAntimeridianContinuity`.
  - **Polar latitudes** within ~0.5° of either pole fall back to
    bilinear, since the latitude grid (unlike longitude) doesn't wrap
    and the bicubic stencil would step off. Covered by new
    `TestPolarFallback`.

- **`TestEGM96GridInterpolationAgainstKnown` tolerance** tightened from
  0.1 m to **0.02 m** (1.85× headroom over the measured 0.0108 m
  worst case, matching the historical safety margin the bilinear test
  used).

### Behavior change to be aware of

A wider sweep across 100k+ off-grid points found that 55% of points
shift by under 0.01 m, but the delta distribution is heavy-tailed:
0.4% of points see a shift of **more than 0.2 m**, with a worst case
near **0.7 m** in regions of strong geoid gradient (e.g. parts of the
Caribbean and equatorial ocean ridges where bilinear's piecewise-linear
kink is most visible). Bicubic is closer to the published UNAVCO truth
in all measured cases — but downstream tests that pinned exact bilinear
output values will see drift up to ~0.7 m in those regions.

### Added

- `pkg/egm96/bench_test.go` with `BenchmarkHeightAboveMSL_Bicubic` and
  `BenchmarkHeightAboveMSL_Bilinear`.
- `TestPolarFallback` and `TestAntimeridianContinuity` in
  `pkg/egm96/egm96_test.go`.

## [v1.2025.4] — 2026-05-09

Performance release — magnetic-field evaluation is **roughly 1.8× faster
for WMMHR and 1.3× faster for standard WMM**, and the per-record getter
chain that batch tools (`wmm_file`, `wmm_grid`) walk for every output
row is **~7.8× faster**. No API change, no tolerance change, no model
update. Existing tests pass within the unchanged 0.05 nT / 0.005°
tolerance.

### Changed (perf)

Measured on Apple M1 Pro / Go 1.26, `go test -bench=. -benchtime=2s`:

| Benchmark | Before | After | Delta |
|---|---|---|---|
| `BenchmarkMagneticField_WMM` (uncached) | 3,322 ns | 2,620 ns | **1.27×** |
| `BenchmarkMagneticField_WMMHR` (uncached) | 264,099 ns | 147,367 ns | **1.79×** |
| `BenchmarkGrid_WMM` (5×5 grid) | 82,371 ns | 65,783 ns | **1.25×** |
| `BenchmarkGrid_WMMHR` (5×5 grid) | 6,780,129 ns | 3,831,891 ns | **1.77×** |
| `BenchmarkGetters` (D, I, F, H, dD, dI, dH, dF, GV, ErrD, Ellipsoidal) | 984 ns | 126 ns | **7.81×** |

Four optimizations, all in `pkg/wmm`:

1. **Iterative power chain.** `computeAtValidDate` was calling
   `polynomial.Pow(AGeo/hh, n+2)` once per outer-loop n, with each call
   doing O(n) multiplications — O(n²) total per location. The base value
   is constant inside the loop, so we now compute the chain
   incrementally: `pwrs[n+1] = pwrs[n] * (AGeo/hh)`, O(n) total.

2. **sin/cos angle-addition recurrence.** The inner loop was calling
   `math.Sin(mf*lambda)` and `math.Cos(mf*lambda)` O(n²) times per
   evaluation. Now precomputed via the standard
   `sin((m+1)λ) = sin(mλ)cos(λ) + cos(mλ)sin(λ)` recurrence in O(n)
   library-trig calls plus O(n) multiplies, then table-lookup in the
   inner loop. Replaces ~18,000 trig library calls per WMMHR evaluation
   with ~270.

3. **Crustal-field secular-variation skip.** WMMHR rows n=16..133 have
   dG=dH=0 (the crustal field is static). The inner loop was still
   computing `f.dx += ...`, `f.dy += ...`, `f.dz += ...` for each of
   those terms — multiplying by zero. New `Model.secVarMaxN` field is
   populated at parse time as the largest n where any (n,m) has nonzero
   dG or dH; the inner loop hoists `hasSecVar := n <= secVarMaxN` once
   per outer-n iteration and conditionally skips the secular-variation
   block. Standard WMM is unaffected (every term has secular variation,
   branch always takes the same way and is well-predicted).

4. **Ellipsoidal cache on `MagneticField`.** Six new float64 fields
   (`xE`, `yE`, `zE`, `dxE`, `dyE`, `dzE`) populated once in
   `(*Model).MagneticField` after the spherical-axis sum. The seven
   ellipsoidal-derived getters (`H`, `D`, `I`, `DH`, `DD`, `DI`, `DF`,
   plus `GV` and `ErrD` which call them) become direct field reads
   instead of triggering the spherical→ellipsoidal rotation each time.
   Most impactful for batch tools — `wmm_file` reads 11 getters per
   result line.

   One caveat: callers that construct a `MagneticField` and then read
   only `Spherical()` (not `Ellipsoidal()` or any derived value) pay
   the ~90 ns cost of the eager rotation regardless. The
   `BenchmarkMagneticField_WMMHR_Cached` benchmark shows this as
   156 ns → 247 ns. For typical batch use this is more than offset by
   the getter-chain savings.

### Added

- `pkg/wmm/bench_test.go` and `pkg/wmm/wmmhr/bench_test.go` with
  `BenchmarkMagneticField_*`, `BenchmarkGetters`, `BenchmarkGrid_*`,
  `BenchmarkParseLoad`, and a `BenchmarkMagneticField_WMMHR_Cached`
  for measuring the per-location-cache hit cost.

### Notes

- `MagneticField` grows from ~88 bytes to ~136 bytes due to the six
  cached ellipsoidal float64s. Negligible for the per-call values our
  callers deal with.
- `polynomial.SchmidtNormalizedALFTable` still allocates `preSqr`,
  `f1`, `f2`, `P`, `dP` per call (~321 KB / 540 allocs for WMMHR). A
  follow-up release will explore Model-resident scratch buffers
  (estimated additional ~3-5% on top of these wins) once benchmark
  numbers warrant it.

## [v1.2025.3] — 2026-05-09

Adds high-resolution WMMHR2025 support as a sibling to the existing
WMM2025 (degree 12) model. Library users opt in by importing the new
`pkg/wmm/wmmhr` sub-package (no cost if you don't); CLI users pass `--hr`
on `wmm_point`, `wmm_grid`, or `wmm_file` (≲4% binary-size impact, or
opt out with `make install-lean` / `-tags wmm_no_hr`).

### Added

- **`pkg/wmm/wmmhr`** sub-package — embeds NOAA's WMMHR2025 coefficients
  (≈530 KB) and exposes `New() (*wmm.Model, error)` and `Default() *wmm.Model`
  constructors that return the same `*wmm.Model` type as standard WMM.
  Importing this package is the trigger for HR data in your binary; library
  consumers who only want standard WMM never see the cost.
- **`--hr` flag** on `cmd/wmm_point`, `cmd/wmm_grid`, `cmd/wmm_file` —
  selects the embedded WMMHR model at runtime. Mutually exclusive with
  `--cof_file`. Default builds bundle both models; `make install-lean`
  (or `go install -tags wmm_no_hr ...`) produces lean binaries that
  omit the HR data and error on `--hr` with a clear message.
- **`Makefile`** with `install`, `install-lean`, `test`, `test-lean`,
  `vet`, `lint` targets — convenience wrappers around `go` commands and
  the `wmm_no_hr` build tag.
- **`(*wmm.Model).MaxN()`** accessor — returns the largest spherical-
  harmonic degree present in the loaded model (12 for standard WMM, 133
  for WMMHR).
- **`polynomial.SchmidtNormalizedALFTable(x, nMax)`** — new public
  function returning the full triangular table of Schmidt semi-normalized
  associated Legendre functions and their latitude derivatives, computed
  via Holmes & Featherstone 2002's stable recurrence. The geomagnetic
  spherical-harmonic loop in `pkg/wmm` now uses this instead of the
  `LegendreFunction(n, m, x)` differentiation path.
- **`polynomial.FactorialFloat(n int)`** — float64 counterpart of the
  existing `Factorial`. The int form silently overflows above n=20;
  `LegendrePolynomial` now uses the float form, which is correct up to
  n≈170.
- New tests: `TestAllWMMHR2025TestValues` (in `pkg/wmm`), three sanity
  tests in `pkg/wmm/wmmhr/wmmhr_test.go`, `TestAgainstNOAAHRSample` in
  `cmd/wmm_file/main_hr_test.go`. CI gains a `test-lean` job that runs
  `go vet`, `go build`, and `go test` with `-tags wmm_no_hr`.

### Changed

- **`(*wmm.Model)` parser is now degree-agnostic.** The
  `MaxLegendreOrder = 12` constant is retained for documentation but no
  longer enforced as a runtime cap; the `Model` carries its own `maxN`
  populated at parse time. Allows loading any well-formed COF file
  regardless of the spherical-harmonic degree it declares.
- **Top-level [README](README.md)** — adds a "WMM vs. WMMHR" section,
  documents the `--hr` flag and the lean-build path.

### Fixed

- **Numerical stability for high spherical-harmonic degree.** The previous
  `LegendreFunction(n, m, x)` path computed the n-th Legendre polynomial,
  m-differentiated it, then evaluated. For n>~20 the polynomial
  coefficients overflow `int64` factorial inputs and lose precision through
  catastrophic cancellation when summed at x ≠ 0. The new
  `SchmidtNormalizedALFTable` recurrence is stable for n in the thousands.
  Standard WMM's existing test values continue to pass within the prior
  tolerance (~0.05 nT / 0.005°); WMMHR's published test values now also
  pass.

## [v1.2025.2] — 2026-05-09

Patch release adding the two CLI commands missing from the v1.0.x line
(`wmm_grid`, `wmm_file`) and a small batch of audit-driven safety fixes.
No coefficient changes; WMM2025 model is unchanged.

### Added

- **`cmd/wmm_grid`** — generates a 4-D grid of magnetic field values
  over latitude, longitude, altitude, and time, evaluating one chosen
  output element (`D`, `I`, `F`, `H`, `X`, `Y`, `Z`, `dD`, `dI`, `dF`,
  `dH`, `dX`, `dY`, `dZ`) per grid point. Flag-driven for scriptability;
  matches NOAA's "collapse an axis by setting start=end" convention.
  Closes the README's "coming soon" promise that has been there since
  2016.
- **`cmd/wmm_file`** — batch-processes a coordinate file matching
  NOAA's reference syntax (`wmm_file f IN OUT`, with optional `e` /
  `--Errors` to append uncertainty columns). Output is byte-for-byte
  identical to NOAA's reference for `K` (km) and `M` (meters) altitude
  inputs; `F` (feet) inputs differ by ≲0.1 nT because we use the exact
  international-foot conversion (0.3048) where NOAA's reference C source
  has a transcription error in `wmm_file.c:606`. The discrepancy is far
  below the WMM uncertainty budget.
- `parsing.ParseNOAAAltitude` (in `internal/parsing`) — parses NOAA's
  `K`/`M`/`F` prefix convention and returns a value in meters.
- `TestMagneticFieldGetters` — direct unit test for every public scalar
  getter on `MagneticField` (D, I, F, H, X, Y, Z, secular variation,
  errors). The full `TestAll20XXTestValuesFromPaper` integration tests
  validate the math; this test exists so a regression in any individual
  accessor fails on its own with a clear name.
- `TestBoundaryCorners` and `TestAgainstNOAASample` — boundary-condition
  tests in `pkg/egm96` and a regression test diffing `wmm_file` output
  against NOAA's published `sample_output.txt`.

### Fixed

- **EGM96 bilinear-bounds off-by-one** in [pkg/egm96/egm96.go](pkg/egm96/egm96.go).
  The bounds checks in `NewLocationMSL`/`HeightAboveMSL`/`NearestEGM96GridPoint`
  used `nLng > egm96XN` / `nLat > egm96YN`, but the bilinear interpolation
  reads `[nLng+1]` and `[(nLat+1)*egm96XN]` — so the largest valid index
  is `egm96XN-2` / `egm96YN-2`. Tightened to match. Latency-masked by the
  v1.2025.1 longitude normalization, but unsafe at boundary latitudes.
- **CLI error propagation** in [cmd/wmm_point/main.go](cmd/wmm_point/main.go).
  A failure from `egm96.NewLocationMSL` was logged but execution continued
  with the zero-value `Location`, producing nonsense output. Now returns
  immediately and exits non-zero on stderr.
- **`polynomial.Derivative(n)` guard** for non-positive n. Previously
  documented as caller-responsibility; calling with `n=-1` would recurse
  forever. Now returns the input unchanged for `n<=0`, and the n=1 path
  handles the `len(p.c) <= 1` (constant or empty) case explicitly.

## [v1.2025.1] — 2026-05-09

Patch release sweeping up four open bugs (#1, #3, #4, #5/#7) and adding
CI. No coefficient changes; WMM2025 model is unchanged from v1.2025.0.

### Fixed

- **#1**: error model is no longer hardcoded as a package constant. New
  `wmm.ErrorModel` struct lives on `*Model`, populated at parse time from
  a per-COF lookup table (`defaultErrorModels`, currently keyed for
  `WMM-2025` and `WMM-2020`). Loading WMM2020 testdata now correctly
  reports `ErrF() = 148`; previously every COF returned the WMM2025
  values the binary was compiled with. Models whose COF name isn't in
  the table return zeros until `(*Model).SetErrorModel` is called.
- **#3**: input longitudes are now normalized to `[0, 360°)` at
  construction. `NewLocationGeodetic(39.86, -121.33, 0).HeightAboveMSL()`
  works (the GPS-convention longitude is wrapped to the EGM96 grid file's
  positive-east convention internally). The package was previously
  internally inconsistent — `NearestEGM96GridPoint` normalized but
  `HeightAboveMSL`/`NewLocationMSL` did not.
- **#4**: race condition in `pkg/egm96`'s lazy grid load. The pattern
  `if len(egm96Grid)==0 { loadEGM96Grid() }` against shared package state
  could be entered by multiple goroutines simultaneously, racing on the
  slice assignment in the loader. Now wrapped in `sync.Once`. Added
  `TestConcurrentHeightAboveMSL` (64 goroutines through `HeightAboveMSL`)
  which trips the race detector reliably without the fix.
- Lint findings from the new CI: unchecked `f.Close` in `LoadModel`,
  ineffectual `lng` assignment in a test.

### Added

- `wmm.ErrorModel`, `(*Model).ErrorModel()`, `(*Model).SetErrorModel(em)`,
  `(MagneticField).ErrorModel()` — see #1.
- GitHub Actions CI workflow at `.github/workflows/ci.yml`. Runs `go vet`,
  `go build`, and `go test -race -count=1 ./...` against Go 1.22 and Go
  stable on every push to master and every pull request. `golangci-lint`
  runs as a separate job (config in `.golangci.yml`), enforcing
  `errcheck`, `gosimple`, `govet`, `ineffassign`, `staticcheck`, and
  `unused`. Closing the gap that allowed PR #8's corrupt bindata to ship
  in the v1.0.x line.
- Install sections in the top-level [README](README.md), [pkg/wmm/README](pkg/wmm/README.md),
  and [pkg/egm96/README](pkg/egm96/README.md). Closes #5 and the install
  half of #7.

### Changed

- `cmd/wmm_point` migrated from the deprecated `wmm.LoadWMMCOF` /
  package-level vars to the `wmm.LoadModel` / `wmm.Default()` /
  `(*Model).MagneticField` API. Output unchanged; the CLI now serves as
  a working example of the recommended public API.

## [v1.2025.0] — 2026-05-09

First release on the new versioning scheme. Embeds WMM2025 (valid
2024-11-13 through 2029-12-31) and modernizes the codebase for current Go
toolchains.

### Embedded model

- WMM2025 (replaces WMM2020). Coefficients sourced from the official NOAA
  distribution at <https://www.ncei.noaa.gov/products/world-magnetic-model>
  (sha256 `06791cd95faba7bdf4a709808f2715a53fe689b29c23b9886bc2196fa9b3eb13`).

### Added

- `wmm.Model` type for explicit, concurrency-safe model handling.
  Constructors: `NewModel` (embedded default), `LoadModel(path)`,
  `ParseModel(io.Reader)`. Methods: `Coefficients(n, m, t)`,
  `MagneticField(loc, t)`, plus `Epoch`/`COFName`/`ValidDate` accessors.
- `wmm.Default()` returns the package-level default `*Model`.
- `pkg/wmm/embedded/WMM.COF` and `pkg/egm96/embedded/ww15mgh.grd` —
  embedded data via `//go:embed`.
- `pkg/wmm/testdata/WMM2025.COF` and `WMM2025_TEST_VALUES.txt` — official
  NOAA WMM2025 test vectors.
- `TestEmbeddedDefaultLoads` and `TestLoadWMMCOFEmptyReloadsEmbedded` —
  guard against future regressions in the embedded default (the kind of
  test that would have caught the bad bindata regeneration in PR #8).
- Tracked `go.mod` (`go 1.22` minimum).
- `CHANGELOG.md`, `CLAUDE.md`.

### Changed

- WMM error model constants updated per the WMM2025 paper:
  `errX 131→137`, `errY 94→89`, `errZ 157→141`, `errH 128→133`,
  `errF 148→138`, `errI 0.21→0.20`, `errDB 5625→5417`.
- Package `wmm` `init()` now panics on embed-load failure (was silent).
  This is a build-time invariant: a corrupt embedded `WMM.COF` is a
  programmer error, not a runtime condition.
- `pkg/polynomial.LegendreFunction` cache now uses `sync.Map` for
  safe concurrent use.
- `pkg/polynomial.Pow`, `Factorial`, `FactorialRatio`,
  `FactorialRatioFloat` rewritten as iterative loops (same signatures).
- `wmm.LoadWMMCOF`, `GetWMMCoefficients`, `CalculateWMMMagneticField`
  retained as thin wrappers over `Default()` for backward compatibility.
- Various `for i := 0; i < len(slice); i++` loops modernized to
  `for i := range slice`.
- `internal/util` directory renamed to `internal/parsing` to match
  the package declaration.
- README's "Updating Coefficients" section now describes the
  `//go:embed` flow rather than the `go-bindata` recipe.

### Removed

- `pkg/wmm/bindata.go` and `pkg/egm96/bindata.go` (replaced by
  `//go:embed`). The `go-bindata` build dependency is gone.
- `io/ioutil` imports throughout (replaced by `os.ReadFile`).

### Fixed

- `pkg/egm96/egm96_test.go`: `ExampleNearestEGM96GridPoint` and
  `ExampleConvertMSLToHeightAboveWGS84` renamed to the correct
  `ExampleType_Method` form (`ExampleLocation_NearestEGM96GridPoint`
  and `ExampleLocation_HeightAboveMSL`), fixing a pre-existing
  `go vet` failure.

### Acknowledgments

- Theo Zourzouvillys ([@zourzouvillys](https://github.com/zourzouvillys))
  — WMM2025 coefficients, error model constants, NOAA test vectors,
  documentation updates ([PR #8](https://github.com/westphae/geomag/pull/8)).

## [v1.0.1] — 2020-01-01

See git history for details on the v1.0.x series. Predates Go modules;
project was used in GOPATH layout.

## [v1.0.0] — 2019-08-24

Initial release.
