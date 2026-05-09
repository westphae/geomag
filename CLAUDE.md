# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Module layout

Standard Go module — `go.mod` declares `module github.com/westphae/geomag` at `go 1.22`. No external runtime dependencies. Embedded data files live under `pkg/wmm/embedded/WMM.COF`, `pkg/egm96/embedded/ww15mgh.grd`, and `pkg/wmm/wmmhr/embedded/WMMHR.COF` (only pulled in by binaries that import `pkg/wmm/wmmhr`).

Versioning is calendar-based on top of SemVer: **`v1.YYYY.x`** where YYYY is the WMM model year and x increments within that model era. v1.2025.0 was the WMM2025 baseline + Go modernization; subsequent v1.2025.x releases are bug fixes and enhancements within the WMM2025 era. The next major bump (`v1.2030.x`) will land when NOAA ships WMM2030. Avoid bumping to `v2.x.y` — Go's major-version-suffix rule would force every import path to change.

## Common commands

```sh
make test                                              # full race-tests
make test-lean                                         # also: go test -tags wmm_no_hr ./...
make vet lint                                          # vet (default + lean) + golangci-lint
make install                                           # install the three CLIs
make install-lean                                      # without HR data (lighter binaries)
go test ./...                                          # plain run all tests
go test ./pkg/wmm -run TestAllWMMHR2025TestValues      # single WMMHR test
go run ./cmd/wmm_point N89 W121 E28 2025.0             # WMM2025 fixture
go run ./cmd/wmm_point --hr N89 W121 E28 2025.0        # same point, WMMHR2025
go run ./cmd/wmm_file [--hr] f IN.txt OUT.txt          # batch processor
go run ./cmd/wmm_grid [--hr] --lat=-30,30,30 --lng=0,0,0 \
    --alt=0,0,0 --date=2026,2026,0 --element=D         # 4-D grid sweep
golangci-lint run ./...                                # static analysis
```

`wmm_point` accepts `--cof_file=<path>` and `--spherical`; lat/lng accept DMS triples (`30,30,30`) or decimal degrees with optional `NSEW` prefix; altitude is in km with optional `E` prefix for height-above-ellipsoid (default is height-above-MSL). `wmm_file` mirrors NOAA's `wmm_file f INPUT OUTPUT` syntax (with optional `e` / `--Errors` between `f` and the input filename to append uncertainty columns); each input line is `<date> <coord-system> <altitude> <lat> <lng>` with `<coord-system>` ∈ {E, M} and `<altitude>` carrying a K/M/F unit prefix (km/meters/feet). `wmm_grid` is flag-driven: each axis takes a `START,END,STEP` triple, set START=END to collapse an axis. All three CLIs accept `--cof_file=PATH` for non-default coefficients.

## Updating WMM coefficients

Drop the new `WMM.COF` into `pkg/wmm/embedded/` (replacing the existing one) and rebuild. The file is embedded at compile time via `//go:embed`; the parser handles arbitrary whitespace, precision, and number of rows. After bumping the embedded default, also:

1. Copy the previous version's `*.COF` into `pkg/wmm/testdata/` so its tests keep passing.
2. Add a per-COF entry in `pkg/wmm/errormodel.go`'s `defaultErrorModels` map keyed on the new COF's name (the second whitespace-separated token on the COF header — e.g. `"WMM-2025"`).
3. Update validity dates in `pkg/wmm/coefficients.go`'s `LoadWMMCOF` doc comment.
4. Bump the embedded WMM2025 row-1 fixture used by tests if the new model invalidates it.

## Architecture

Three library packages plus three CLI commands:

- **`pkg/egm96`** — EGM96 geoid model. Owns the shared `Location` type used by every other package. A `Location` stores geodetic (φ, λ, h) values; longitude is normalized to `[0, 360°)` at construction so `NewLocationGeodetic` accepts both GPS `[-180, 180]` and NGA `[0, 360]` conventions transparently. `HeightAboveMSL` and `NewLocationMSL` bilinearly interpolate the embedded 15′×15′ NGA grid; the grid load is `sync.Once`-protected for concurrent first-use. `NearestEGM96GridPoint` is a single-cell cousin.

- **`pkg/polynomial`** — Legendre polynomial machinery. `LegendreFunction(n, m, x)` is the hot path: it differentiates `LegendrePolynomial(n)` `m` times and caches the result in a process-wide `sync.Map` keyed by `(n, m)`. The Schmidt semi-normalization (`sqrt(2 · (n−m)! / (n+m)!)` for m>0) is applied by the **caller** in `pkg/wmm`, not here. `Pow`/`Factorial`/`FactorialRatio*` are iterative and tolerate edge inputs (negative n, m>n) without panicking.

- **`pkg/wmm`** — WMM evaluation. The primary type is `Model`, `sync.RWMutex`-guarded, holding the parsed coefficients, the published error model (loaded at parse time from `defaultErrorModels` keyed on the COF name), and a per-location cache. Construct via `NewModel()` (embedded default), `LoadModel(path)`, or `ParseModel(io.Reader)`. Methods: `Coefficients(n, m, t)`, `MagneticField(loc, t)`, `Epoch()`, `COFName()`, `ValidDate()`, `MaxN()`, `ErrorModel()`, `SetErrorModel(em)`. The package also exposes a thread-safe package-level `Default()` plus thin back-compat wrappers (`LoadWMMCOF`, `GetWMMCoefficients`, `CalculateWMMMagneticField`) and the legacy `Epoch`/`COFName`/`ValidDate` vars that mirror the default model's state. **`init()` panics if the embedded `WMM.COF` fails to parse** — a build-time invariant that prevents silent embed corruption (this is exactly the regression mode that allowed PR #8's broken bindata to ship in the v1.0.x line). Per the cache design, callers iterating over many points should make **time the innermost loop**, then height/lat/lng, to maximize hits in the per-location cache (the cache stores the spherical-harmonic sum evaluated at `validDate`, then the result is linearly time-corrected on read). The parser handles arbitrary spherical-harmonic degree (the `MaxLegendreOrder = 12` constant is documentation-only since v1.2025.3); models with degree up to ~170 are supported.

- **`pkg/wmm/wmmhr`** — WMMHR (high-resolution) sibling. Embeds NOAA's WMMHR2025 coefficients (≈530 KB, degree 133) and exposes `New() (*wmm.Model, error)` and `Default() *wmm.Model` returning the same `*wmm.Model` type as standard WMM. Importing this sub-package is what pulls the HR data into a binary; library callers who don't import it pay zero cost. The CLIs (`wmm_point`, `wmm_grid`, `wmm_file`) accept `--hr` to select the embedded HR model at runtime; mutually exclusive with `--cof_file`. The `wmm_no_hr` build tag (via `make install-lean` or `go install -tags wmm_no_hr ...`) omits HR data from CLI binaries, with `--hr` then erroring with a clear message at runtime.

- **`cmd/wmm_point`** — Single-point CLI. Interactive when run with no positional args; takes lat, lng, alt, date as positional args otherwise. Returns non-zero on `egm96.NewLocationMSL` failure (does not silently continue with a zero-value Location).

- **`cmd/wmm_file`** — Batch processor matching NOAA's `wmm_file f INPUT OUTPUT` syntax. Output is byte-for-byte identical to NOAA's reference for K/M altitudes; F (feet) altitudes differ by ≲0.1 nT because we use the exact international-foot conversion (0.3048) where NOAA's reference C source has a transcription error in `wmm_file.c:606` (`3280.0839895` instead of `1000/0.3048 = 3280.83989501…`). Documented in the command's godoc.

- **`cmd/wmm_grid`** — 4-D grid sweep over (lat × lng × alt × time), one chosen output element per row. Flag-driven (`--lat=START,END,STEP`, etc.); collapse an axis by setting START=END. Output via stdout or `--output=FILE`.

### Coordinate conventions

The WMM math runs in **spherical** geocentric coordinates; results are rotated to **ellipsoidal** axes via `MagneticField.Ellipsoidal()`, which applies a rotation by `(spherical_lat − geodetic_lat)`. `Spherical()` returns the raw values. Most navigation use cases want `Ellipsoidal()` (or the derived `H()`, `D()`, `I()`, `GV()`).

### Per-Model error model

The published global-average uncertainties (`errX`, `errY`, `errZ`, `errH`, `errF`, `errI`, `errDA`, `errDB`) change with each WMM release, so they live on `*Model`, not as package constants. They're populated at parse time from a per-COF lookup (`pkg/wmm/errormodel.go::defaultErrorModels`, keyed on the COF name token like `"WMM-2025"`). Loading WMM2020 testdata correctly yields WMM2020 uncertainties; models whose COF name isn't in the table get a zero `ErrorModel` (and `MagneticField.ErrX()` etc. return zero) until `(*Model).SetErrorModel(em)` is called. `MagneticField.Err{X,Y,Z,H,F,I,D}` all read from this per-`*Model` state, not from package constants.

### Input parsing

`internal/parsing` provides:
- `ParseLatLng` — DMS triples or decimal degrees with optional NSEW prefix
- `ParseAltitude` — km, optional `E` prefix for height-above-ellipsoid (`wmm_point` convention)
- `ParseNOAAAltitude` — K/M/F prefix + signed decimal magnitude, returning meters (`wmm_file` convention)
- `ParseTime` — decimal year or `MM DD YYYY` / `MM/DD/YYYY`

The package directory is `internal/parsing` and the package name is `parsing` (no longer the historical `internal/util` mismatch).

## Validation

Test fixtures in `pkg/wmm/testdata/` are the upstream NOAA WMM test vectors (WMM2015v1, WMM2015v2, WMM2020, WMM2025). The `TestAll20XXTestValuesFromPaper` tests in `magnetic_field_test.go` exercise each set; the WMM2015 worked example from section 1.5 of the WMM technical report is also covered. `TestEmbeddedDefaultLoads` and `TestLoadWMMCOFEmptyReloadsEmbedded` in `coefficients_test.go` guard the embedded default against silent corruption (the regression mode that bit PR #8 in the v1.0.x line). `TestMagneticFieldGetters` exercises every public `MagneticField` getter against a single fixture point. `TestBoundaryCorners` exercises the EGM96 grid edges. `TestConcurrentHeightAboveMSL` drives `pkg/egm96` from many goroutines under `-race` to guard the `sync.Once` grid load. `cmd/wmm_file/main_test.go::TestAgainstNOAASample` diffs our output against NOAA's reference `sample_output.txt` line-by-line with column tolerances tuned to the foot-conversion divergence. Any change to the field math or coefficient parsing should keep these tests passing to within ~0.05 nT / 0.005°.

## CI

`.github/workflows/ci.yml` runs on every push to master and every PR:

- **`test`** matrix: `go vet`, `go build`, `go test -race -count=1 ./...` against Go 1.22 (the declared minimum) and Go stable
- **`golangci-lint`**: action `@v8` with golangci-lint pinned to v2.12.2 (action v6 only supports golangci-lint v1; v2 config schema requires v7+ action). Config in `.golangci.yml` enables `errcheck`, `gosimple`, `govet`, `ineffassign`, `staticcheck`, `unused`. The `errcheck` linter would have caught the `_ = LoadWMMCOF(...)` pattern that hid PR #8's broken init in the v1.0.x line.

To run locally before pushing: `brew install golangci-lint && golangci-lint run ./...`. The bundled binary is golangci-lint v2.x; v1 configs (no `version: "2"` field) are rejected.

## Conventions

- **Commit messages**: imperative-mood subject under 70 chars; bullet body explaining *why*. Reference issues with `Closes #N` / `Resolves #N` in the trailer to trigger GitHub auto-close on push to master.
- **Direct push to master**: this repo's normal flow (no PR-required policy). The CI runs after the push and we react to failures.
- **Tags are annotated** (`git tag -a v1.YYYY.x -m "…"`), referencing the CHANGELOG entry. Releases are created via `gh release create vTAG --notes-file <CHANGELOG-section>`.
- **CHANGELOG.md** uses [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) sections (Added/Changed/Fixed/Removed/Acknowledgments). The first paragraph of each release section is a one-line summary; the rest is structured.
