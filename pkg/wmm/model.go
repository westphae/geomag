package wmm

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/westphae/geomag/pkg/egm96"
	"github.com/westphae/geomag/pkg/polynomial"
)

// Model holds a parsed WMM coefficients file and evaluates the geomagnetic
// field at any (location, time) within the model's validity period.
//
// Methods on *Model are safe for concurrent use.
type Model struct {
	mu        sync.RWMutex
	epoch     DecimalYear
	cofName   string
	validDate time.Time
	maxN      int // largest n seen in the COF; 12 for standard WMM, 133 for WMMHR
	cGnm      [][]float64
	cHnm      [][]float64
	cDGnm     [][]float64
	cDHnm     [][]float64
	errors    ErrorModel // populated from defaultErrorModels[cofName] at load; zero if unknown

	cacheMu   sync.Mutex
	haveCache bool
	cacheLoc  egm96.Location
	cache     MagneticField // field evaluated at validDate; time-corrected on read
}

// NewModel returns a Model loaded from the embedded default coefficients.
func NewModel() (*Model, error) {
	return ParseModel(bytes.NewReader(defaultCOF))
}

// LoadModel returns a Model loaded from the given .COF file path.
func LoadModel(path string) (*Model, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }() // read-only file; close errors are not actionable
	return ParseModel(f)
}

// ParseModel returns a Model parsed from the given .COF reader.
func ParseModel(r io.Reader) (*Model, error) {
	m := &Model{}
	if err := m.parse(r); err != nil {
		return nil, err
	}
	return m, nil
}

// Epoch returns the model's epoch (e.g. 2025.0).
func (m *Model) Epoch() DecimalYear {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.epoch
}

// COFName returns the model's name (e.g. "WMM-2025").
func (m *Model) COFName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cofName
}

// ValidDate returns the model's start-of-validity date.
func (m *Model) ValidDate() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.validDate
}

// MaxN returns the largest spherical-harmonic degree present in the loaded
// model. Standard WMM is 12; WMMHR is 133.
func (m *Model) MaxN() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.maxN
}

// ErrorModel returns the model's published global-average uncertainty values.
// For models whose COF name is not in the package's lookup table, this returns
// a zero ErrorModel until SetErrorModel is called.
func (m *Model) ErrorModel() ErrorModel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.errors
}

// SetErrorModel overrides the model's error model. Use this for custom or
// future WMM releases not in the package's defaultErrorModels lookup.
func (m *Model) SetErrorModel(em ErrorModel) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors = em
}

// Coefficients returns the spherical-harmonic coefficients G(n,m), H(n,m) and
// their rates of change dG(n,m), dH(n,m) at time t.
//
// If the request n,m are out of range or t lies outside the model's validity
// period, an error is returned. The returned values are still computed in the
// validity-period case.
func (m *Model) Coefficients(n, mm int, t time.Time) (g, h, dg, dh float64, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if n < 0 || n > m.maxN || mm < 0 || mm > m.maxN {
		return 0, 0, 0, 0, fmt.Errorf("n, m = (%d,%d) must be between 0 and %d",
			n, mm, m.maxN)
	}
	if mm > n {
		return 0, 0, 0, 0, fmt.Errorf("m=%d must be less than n=%d", mm, n)
	}
	if t.Sub(m.validDate) < 0 || TimeToDecimalYears(t) > m.epoch+5 {
		err = fmt.Errorf("requested date %v is outside of validity period beginning %v of WMM.COF file",
			t, m.validDate)
	}
	dt := float64(TimeToDecimalYears(t) - m.epoch)
	g = m.cGnm[n][mm] + dt*m.cDGnm[n][mm]
	h = m.cHnm[n][mm] + dt*m.cDHnm[n][mm]
	dg = m.cDGnm[n][mm]
	dh = m.cDHnm[n][mm]
	return g, h, dg, dh, err
}

// MagneticField evaluates the geomagnetic field at the given location and
// time. The model caches the per-location spherical-harmonic sum, so iterating
// over time at a fixed location is essentially free.
//
// The WMM is valid at heights from -1km to +850km relative to the WGS84
// ellipsoid; outside that range or outside the validity period, the function
// returns an informational error along with the calculated field.
func (m *Model) MagneticField(loc egm96.Location, t time.Time) (MagneticField, error) {
	m.cacheMu.Lock()
	if !m.haveCache || !loc.Equals(m.cacheLoc) {
		m.cache = m.computeAtValidDate(loc)
		m.cacheLoc = loc
		m.haveCache = true
	}
	cached := m.cache
	m.cacheMu.Unlock()

	m.mu.RLock()
	validDate := m.validDate
	epoch := m.epoch
	errors := m.errors
	m.mu.RUnlock()

	var err error
	if t.Sub(validDate) < 0 || TimeToDecimalYears(t) > epoch+5 {
		err = fmt.Errorf("requested date %v is outside of validity period beginning %v of WMM.COF file",
			t, validDate)
	}

	dt := float64(TimeToDecimalYears(t) - TimeToDecimalYears(validDate))
	field := MagneticField{
		l:      loc,
		x:      cached.x + dt*cached.dx,
		y:      cached.y + dt*cached.dy,
		z:      cached.z + dt*cached.dz,
		dx:     cached.dx,
		dy:     cached.dy,
		dz:     cached.dz,
		errors: errors,
	}
	return field, err
}

// computeAtValidDate evaluates the spherical-harmonic sum at validDate
// (a fixed reference time) so the result can be linearly time-corrected later.
// Caller must NOT hold m.mu.
func (m *Model) computeAtValidDate(loc egm96.Location) MagneticField {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var f MagneticField
	phi, lambda, hh := loc.Spherical()
	sinPhi := math.Sin(phi)
	cosPhi := math.Cos(phi)

	// Schmidt-normalized associated Legendre table (and its φ-derivative)
	// computed via Holmes & Featherstone's stable recurrence. This replaces
	// the older per-(n,m) LegendreFunction differentiation, which suffered
	// catastrophic cancellation above n≈20 — fine for standard WMM but
	// useless at WMMHR's degree 133.
	P, dP := polynomial.SchmidtNormalizedALFTable(sinPhi, m.maxN)
	if P == nil {
		// At |lat| == 90° the recurrence's dP/dφ is undefined; fall back
		// to whatever the cache had (zero-init MagneticField).
		return f
	}

	dtRef := float64(TimeToDecimalYears(m.validDate) - m.epoch)
	for n := 1; n <= m.maxN; n++ {
		nn := float64(n + 1)
		pwr := polynomial.Pow(AGeo/hh, n+2)
		for mm := 0; mm <= n; mm++ {
			mf := float64(mm)
			p := P[n][mm]
			dp := dP[n][mm]
			g := m.cGnm[n][mm] + dtRef*m.cDGnm[n][mm]
			h := m.cHnm[n][mm] + dtRef*m.cDHnm[n][mm]
			dg := m.cDGnm[n][mm]
			dh := m.cDHnm[n][mm]
			sinMLambda := math.Sin(mf * lambda)
			cosMLambda := math.Cos(mf * lambda)
			f.x += -pwr * (g*cosMLambda + h*sinMLambda) * dp
			f.y += pwr / cosPhi * mf * (g*sinMLambda - h*cosMLambda) * p
			f.z += -nn * pwr * (g*cosMLambda + h*sinMLambda) * p
			f.dx += -pwr * (dg*cosMLambda + dh*sinMLambda) * dp
			f.dy += pwr / cosPhi * mf * (dg*sinMLambda - dh*cosMLambda) * p
			f.dz += -nn * pwr * (dg*cosMLambda + dh*sinMLambda) * p
		}
	}
	return f
}

func (m *Model) parse(r io.Reader) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return fmt.Errorf("could not read header line in WMM coefficient file")
	}
	dat := strings.Fields(scanner.Text())
	epoch, err := strconv.ParseFloat(dat[0], 64)
	if err != nil {
		return fmt.Errorf("bad header epoch: %w", err)
	}
	m.epoch = DecimalYear(epoch)
	m.cofName = dat[1]
	if m.validDate, err = time.Parse("01/02/2006", dat[2]); err != nil {
		return fmt.Errorf("bad header valid date: %w", err)
	}
	m.errors = defaultErrorModels[m.cofName] // zero ErrorModel if unknown

	// Initialize the n=0 row only; subsequent rows are allocated as the
	// parser sees them, supporting arbitrary degrees up to whatever the
	// COF declares. WMM is degree 12; WMMHR is degree 133.
	m.cGnm = [][]float64{{0}}
	m.cHnm = [][]float64{{0}}
	m.cDGnm = [][]float64{{0}}
	m.cDHnm = [][]float64{{0}}

	for scanner.Scan() {
		s := strings.Fields(scanner.Text())
		if len(s) < 6 {
			continue
		}
		n, err := strconv.Atoi(s[0])
		if err != nil {
			return fmt.Errorf("bad n value: %w", err)
		}
		mm, err := strconv.Atoi(s[1])
		if err != nil {
			return fmt.Errorf("bad m value: %w", err)
		}
		// Grow each slice up through index n, allocating the inner row
		// for the new top-level index. Idempotent on re-visit (no
		// reallocation when n stays the same).
		for len(m.cGnm) <= n {
			next := len(m.cGnm)
			m.cGnm = append(m.cGnm, make([]float64, next+1))
			m.cHnm = append(m.cHnm, make([]float64, next+1))
			m.cDGnm = append(m.cDGnm, make([]float64, next+1))
			m.cDHnm = append(m.cDHnm, make([]float64, next+1))
		}
		if n > m.maxN {
			m.maxN = n
		}
		if m.cGnm[n][mm], err = strconv.ParseFloat(s[2], 64); err != nil {
			return fmt.Errorf("bad Gnm value: %w", err)
		}
		if m.cHnm[n][mm], err = strconv.ParseFloat(s[3], 64); err != nil {
			return fmt.Errorf("bad Hnm value: %w", err)
		}
		if m.cDGnm[n][mm], err = strconv.ParseFloat(s[4], 64); err != nil {
			return fmt.Errorf("bad DGnm value: %w", err)
		}
		if m.cDHnm[n][mm], err = strconv.ParseFloat(s[5], 64); err != nil {
			return fmt.Errorf("bad DHnm value: %w", err)
		}
	}
	return scanner.Err()
}
