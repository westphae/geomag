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
	cGnm      [][]float64
	cHnm      [][]float64
	cDGnm     [][]float64
	cDHnm     [][]float64

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
	defer f.Close()
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

// Coefficients returns the spherical-harmonic coefficients G(n,m), H(n,m) and
// their rates of change dG(n,m), dH(n,m) at time t.
//
// If the request n,m are out of range or t lies outside the model's validity
// period, an error is returned. The returned values are still computed in the
// validity-period case.
func (m *Model) Coefficients(n, mm int, t time.Time) (g, h, dg, dh float64, err error) {
	if n < 0 || n > MaxLegendreOrder || mm < 0 || mm > MaxLegendreOrder {
		return 0, 0, 0, 0, fmt.Errorf("n, m = (%d,%d) must be between 0 and %d",
			n, mm, MaxLegendreOrder)
	}
	if mm > n {
		return 0, 0, 0, 0, fmt.Errorf("m=%d must be less than n=%d", mm, n)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
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
	m.mu.RUnlock()

	var err error
	if t.Sub(validDate) < 0 || TimeToDecimalYears(t) > epoch+5 {
		err = fmt.Errorf("requested date %v is outside of validity period beginning %v of WMM.COF file",
			t, validDate)
	}

	dt := float64(TimeToDecimalYears(t) - TimeToDecimalYears(validDate))
	field := MagneticField{
		l:  loc,
		x:  cached.x + dt*cached.dx,
		y:  cached.y + dt*cached.dy,
		z:  cached.z + dt*cached.dz,
		dx: cached.dx,
		dy: cached.dy,
		dz: cached.dz,
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
	dtRef := float64(TimeToDecimalYears(m.validDate) - m.epoch)
	for n := 1; n <= MaxLegendreOrder; n++ {
		nn := float64(n + 1)
		pwr := polynomial.Pow(AGeo/hh, n+2)
		for mm := 0; mm <= n; mm++ {
			mf := float64(mm)
			p := polynomial.LegendreFunction(n, mm, sinPhi)
			q := polynomial.LegendreFunction(n+1, mm, sinPhi)
			if mm > 0 {
				s := math.Sqrt(2 / polynomial.FactorialRatioFloat(n+mm, n-mm))
				p *= s
				q *= s
			}
			dp := nn*math.Tan(phi)*p - (nn-mf)/cosPhi*q
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

	m.cGnm = make([][]float64, MaxLegendreOrder+1)
	m.cGnm[0] = []float64{0}
	m.cHnm = make([][]float64, MaxLegendreOrder+1)
	m.cHnm[0] = []float64{0}
	m.cDGnm = make([][]float64, MaxLegendreOrder+1)
	m.cDGnm[0] = []float64{0}
	m.cDHnm = make([][]float64, MaxLegendreOrder+1)
	m.cDHnm[0] = []float64{0}

	curN := 0
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
		if n > curN {
			m.cGnm[n] = make([]float64, n+1)
			m.cHnm[n] = make([]float64, n+1)
			m.cDGnm[n] = make([]float64, n+1)
			m.cDHnm[n] = make([]float64, n+1)
			curN = n
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
