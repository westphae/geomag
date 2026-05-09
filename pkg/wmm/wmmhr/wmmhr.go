// Package wmmhr provides the WMMHR (World Magnetic Model High Resolution)
// coefficients as a sibling to the standard WMM model in pkg/wmm.
//
// WMMHR extends the spherical-harmonic series from degree 12 (standard WMM)
// to degree 133, capturing finer-scale variations in the geomagnetic field
// — particularly improvements in declination accuracy near regional crustal
// anomalies. The trade-off is computational cost: a single magnetic-field
// evaluation runs ~85× more inner-loop iterations than standard WMM. Per-
// location caching on *Model absorbs that cost completely for time-sweep
// loops, but for one-off evaluations expect a per-call latency increase
// from microseconds to fractions-of-a-millisecond.
//
// Importing this package adds the embedded WMMHR.COF (≈530 KB) to your
// binary. Library consumers who only need standard WMM should import
// github.com/westphae/geomag/pkg/wmm and not import this package, paying
// no HR cost.
//
// The returned values are *wmm.Model, the same type as the standard model,
// so all the existing methods (MagneticField, Coefficients, ErrorModel,
// COFName, Epoch, ValidDate, MaxN, etc.) work unchanged.
//
// Usage:
//
//	import (
//	    "time"
//	    "github.com/westphae/geomag/pkg/egm96"
//	    "github.com/westphae/geomag/pkg/wmm/wmmhr"
//	)
//
//	loc := egm96.NewLocationGeodetic(89, -121, 28*1000)
//	field, err := wmmhr.Default().MagneticField(loc, time.Now())
//	decl := field.D() // declination in degrees, computed at HR resolution
//
// WMMHR2025 is valid from 2025.0 through 2030.0. Source data downloaded
// from https://www.ncei.noaa.gov/products/world-magnetic-model-high-resolution
// (sha256 8851d40e57a1d948cb56d49b837612844890a941f93a73846a122b6c1182d504).
package wmmhr

import (
	"bytes"
	_ "embed"
	"fmt"
	"sync"

	"github.com/westphae/geomag/pkg/wmm"
)

//go:embed embedded/WMMHR.COF
var defaultCOF []byte

// New returns a *wmm.Model loaded from a fresh parse of the embedded
// WMMHR2025 coefficients. Each call allocates a new model; callers
// iterating over many points should use Default instead to share the
// per-location cache.
func New() (*wmm.Model, error) {
	return wmm.ParseModel(bytes.NewReader(defaultCOF))
}

// Default returns the package-level default WMMHR Model, lazy-initialized
// on first call and shared thereafter. Subsequent calls return the same
// *wmm.Model — useful so its per-location cache can be hit across calls
// from anywhere in the program.
//
// Panics on first-call parse failure: the embedded WMMHR.COF is loaded
// from data baked into the binary at compile time, so a parse failure
// indicates a build-time corruption and is not a recoverable runtime
// condition. This matches the behavior of pkg/wmm's package-level init().
func Default() *wmm.Model {
	defaultOnce.Do(func() {
		m, err := New()
		if err != nil {
			panic(fmt.Sprintf("wmmhr: failed to load embedded WMMHR coefficients: %v", err))
		}
		defaultModel = m
	})
	return defaultModel
}

var (
	defaultOnce  sync.Once
	defaultModel *wmm.Model
)
