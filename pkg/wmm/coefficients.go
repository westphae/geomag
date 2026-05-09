package wmm

import (
	_ "embed"
	"fmt"
	"sync"
	"time"
)

const (
	MaxLegendreOrder = 12
)

//go:embed embedded/WMM.COF
var defaultCOF []byte

// Package-level state mirrors the most-recently-loaded default Model. New code
// should construct an explicit *Model via NewModel or LoadModel; these vars
// are kept for backward compatibility and are not safe under concurrent
// LoadWMMCOF calls.
var (
	Epoch     DecimalYear // Epoch of the loaded coefficients file, e.g. 2025.0
	COFName   string      // Filename of the loaded COF file
	ValidDate time.Time   // Beginning of the validity period
)

var (
	defaultModelMu sync.RWMutex
	defaultModel   *Model
)

func init() {
	m, err := NewModel()
	if err != nil {
		// The default model is loaded from data embedded at compile time via
		// //go:embed. A parse failure here means the embedded WMM.COF has been
		// corrupted in source — this is a build-time bug, not a runtime
		// condition, so we surface it loudly rather than leave the package in
		// an empty state where every subsequent call would panic with
		// confusing "index out of range" errors.
		panic(fmt.Sprintf("wmm: failed to load embedded WMM coefficients: %v", err))
	}
	setDefaultModel(m)
}

func setDefaultModel(m *Model) {
	defaultModelMu.Lock()
	defaultModel = m
	Epoch = m.epoch
	COFName = m.cofName
	ValidDate = m.validDate
	defaultModelMu.Unlock()
}

// Default returns the package-level default Model (the one most recently
// loaded via LoadWMMCOF, or the embedded default if none has been loaded).
func Default() *Model {
	defaultModelMu.RLock()
	defer defaultModelMu.RUnlock()
	return defaultModel
}

// LoadWMMCOF replaces the package-level default model with one loaded from
// the given path. Passing "" reloads the embedded default.
//
// Deprecated: prefer LoadModel/NewModel for new code; the package-level
// default model is global mutable state.
func LoadWMMCOF(fn string) error {
	var (
		m   *Model
		err error
	)
	if fn == "" {
		m, err = NewModel()
	} else {
		m, err = LoadModel(fn)
	}
	if err != nil {
		return err
	}
	setDefaultModel(m)
	return nil
}

// GetWMMCoefficients calculates the spherical harmonic coefficients G(n,m),
// H(n,m), and their rates of change at time t using the package-level default
// model.
func GetWMMCoefficients(n, m int, t time.Time) (gnm, hnm, dgnm, dhnm float64, err error) {
	return Default().Coefficients(n, m, t)
}
