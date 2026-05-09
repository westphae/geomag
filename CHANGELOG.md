# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
This project uses a year-based versioning scheme on top of [SemVer](https://semver.org/):
**MINOR** encodes the WMM model year (`v1.2025.x` for releases that embed
WMM2025, `v1.2030.x` for the next NOAA release, …); **PATCH** increments
within a model era for data reissues, code fixes, or improvements.

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
