# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
This project uses a year-based versioning scheme on top of [SemVer](https://semver.org/):
**MINOR** encodes the WMM model year (`v1.2025.x` for releases that embed
WMM2025, `v1.2030.x` for the next NOAA release, …); **PATCH** increments
within a model era for data reissues, code fixes, or improvements.

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
